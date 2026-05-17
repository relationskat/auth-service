package production

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/relationskat/auth-service/internal/cache"
	"github.com/relationskat/auth-service/internal/cache/memcached"
	"github.com/relationskat/auth-service/internal/domain"
	"github.com/relationskat/auth-service/internal/models/dto"
	"github.com/relationskat/auth-service/internal/pkg/provider"
	"github.com/relationskat/auth-service/internal/repository"
	"github.com/relationskat/auth-service/internal/smtp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	log        *zap.Logger
	repository repository.Repository
	tokens     cache.Cache
	mailer     *smtp.Mailer
	provider   *provider.Provider
	baseURL    string
}

func New(
	log *zap.Logger,
	repository repository.Repository,
	tokens cache.Cache,
	mailer *smtp.Mailer,
	provider *provider.Provider,
	baseURL string,
) (*Service, error) {
	return &Service{
		log:        log.Named("Service"),
		repository: repository,
		tokens:     tokens,
		mailer:     mailer,
		provider:   provider,
		baseURL:    baseURL,
	}, nil
}

func (s *Service) CreateUser(ctx context.Context, user *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	const op = "production.CreateUser"

	passwordHash, err := hashPassword(user.Password)
	if err != nil {
		s.log.Debug("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	useDTO := &dto.User{
		ID:           uuid.New().String(),
		LastName:     user.LastName,
		FirstName:    user.FirstName,
		MiddleName:   user.MiddleName,
		Email:        user.Email,
		PasswordHash: passwordHash,
	}

	if err = s.repository.CreateUser(ctx, useDTO); err != nil {
		s.log.Debug("failed to add user in db", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	emailToken := generateSecureToken(64)

	if err := s.tokens.SetToken(emailToken, useDTO.Email, 24*time.Hour); err != nil {
		s.log.Error("failed to store email token", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf(
		`<p>Здравствуйте, %s!</p><p>Для подтверждения почты перейдите по ссылке:</p>`+
			`<p><a href="%s/verify_email?token=%s">Подтвердить email</a></p>`,
		useDTO.FirstName, s.baseURL, emailToken,
	)

	if err := s.mailer.Send(ctx, []string{useDTO.Email}, subject, body); err != nil {
		s.log.Debug("failed to send confirmation email", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &domain.RegisterResponse{
		ID:    useDTO.ID,
		Email: useDTO.Email,
	}, nil
}

func (s *Service) AcceptEmail(ctx context.Context, token string) (bool, error) {
	const op = "production.AcceptEmail"

	email, err := s.tokens.GetToken(token)
	if err != nil {
		s.log.Debug("failed to get token", zap.Error(err))
		if errors.Is(err, memcached.ErrNotFound) {
			return false, fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	err = s.tokens.DeleteToken(token)
	if err != nil {
		s.log.Debug("failed to delete token", zap.Error(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	err = s.repository.AcceptEmail(ctx, email)
	if err != nil {
		s.log.Debug("failed to accept email", zap.Error(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (string, string, error) {
	const op = "production.Login"

	id, hashedPassword, err := s.repository.Login(ctx, email)
	if err != nil {
		s.log.Debug("login: user lookup failed", zap.Error(err))
		return "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err !=
		nil {
		s.log.Debug("login: password mismatch", zap.String("email", email))
		return "", "", ErrInvalidCredentials
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		s.log.Error("login: bad user id from db", zap.String("id", id), zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := s.provider.NewJwt(userID, "", true)
	if err != nil {
		s.log.Error("login: failed to create access token", zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.provider.NewJwt(userID, "", false)
	if err != nil {
		s.log.Error("login: failed to create refresh token", zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, token string) (string, string, error) {
	const op = "production.RefreshToken"

	c, err := s.provider.ValidateToken(token, false)
	if err != nil {
		s.log.Debug("refresh: invalid token", zap.Error(err))
		return "", "", err
	}

	userID, err := uuid.Parse(c.ID)
	if err != nil {
		s.log.Error("refresh: bad user id in token", zap.String("id", c.ID), zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, provider.ErrTokenInvalid)
	}

	accessToken, err := s.provider.NewJwt(userID, "", true)
	if err != nil {
		s.log.Error("refresh: failed to create access token", zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.provider.NewJwt(userID, "", false)
	if err != nil {
		s.log.Error("refresh: failed to create refresh token", zap.Error(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) Auth(ctx context.Context, token string) (string, error) {
	const op = "production.Auth"

	c, err := s.provider.ValidateToken(token, true)
	if err != nil {
		s.log.Debug("auth: invalid token", zap.Error(err))
		return "", err
	}

	if _, err := uuid.Parse(c.ID); err != nil {
		s.log.Error("auth: bad user id in token", zap.String("id", c.ID), zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, provider.ErrTokenInvalid)
	}

	return c.ID, nil
}

func (s *Service) Update(ctx context.Context, user *dto.UserUpdate) (*dto.UserUpdate, error) {
	const op = "production.Update"

	if _, err := uuid.Parse(user.ID); err != nil {
		s.log.Debug("update: bad user id", zap.String("id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	updated, err := s.repository.Update(ctx, user)
	if err != nil {
		s.log.Debug("update: failed to update user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updated, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	const op = "production.RequestPasswordReset"

	resetToken := generateSecureToken(64)

	if err := s.tokens.SetToken(resetToken, email, 1*time.Hour); err != nil {
		s.log.Error("reset: failed to store token", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	subject := "Сброс пароля"
	body := fmt.Sprintf(
		`<p>Вы запросили сброс пароля.</p>`+
			`<p><a href="%s/reset-password?token=%s">Установить новый пароль<p>`+
			`<p>Если это были не вы — проигнорируйте письмо. Ссылка действует 1 час.<p>`,
		s.baseURL, resetToken,
	)

	if err := s.mailer.Send(ctx, []string{email}, subject, body); err != nil {
		s.log.Debug("reset: failed to send email", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	const op = "production.ConfirmPasswordReset"

	email, err := s.tokens.GetToken(token)
	if err != nil {
		s.log.Debug("reset: failed to get token", zap.Error(err))
		if errors.Is(err, memcached.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		s.log.Debug("reset: failed to hash password", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.repository.UpdatePasswordByEmail(ctx, email, passwordHash); err != nil {
		s.log.Debug("reset: failed to update password", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.tokens.DeleteToken(token); err != nil {
		s.log.Warn("reset: failed to delete used token", zap.Error(err))
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	const op = "production.ChangePassword"

	if _, err := uuid.Parse(id); err != nil {
		s.log.Debug("change password: bad user id", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	currentHash, err := s.repository.GetPasswordHash(ctx, id)
	if err != nil {
		s.log.Debug("change password: failed to get current hash", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		s.log.Debug("change password: old password mismatch", zap.String("id", id))
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if oldPassword == newPassword {
		return fmt.Errorf("%s: %w", op, ErrSamePassword)
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		s.log.Debug("change password: failed to hash new password", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.repository.UpdatePassword(ctx, id, newHash); err != nil {
		s.log.Debug("change password: failed to update password", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	const op = "production.DeleteUser"

	if _, err := uuid.Parse(id); err != nil {
		s.log.Debug("delete user: bad user id", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if err := s.repository.DeleteUser(ctx, id); err != nil {
		s.log.Debug("delete user: failed", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ResendConfirmationEmail(ctx context.Context, email string) error {
	const op = "production.ResendConfirmationEmail"

	firstName, confirmed, err := s.repository.GetEmailStatus(ctx, email)
	if err != nil {
		s.log.Debug("resend: failed to get email status", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if confirmed {
		return fmt.Errorf("%s: %w", op, ErrEmailAlreadyConfirmed)
	}

	emailToken := generateSecureToken(64)

	if err := s.tokens.SetToken(emailToken, email, 24*time.Hour); err != nil {
		s.log.Error("resend: failed to store email token", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf(
		`<p>Здравствуйте, %s!</p><p>Для подтверждения почты перейдите по ссылке:</p>`+
			`<p><a href="%s/verify_email?token=%s">Подтвердить email</a></p>`+
			`<p>Ссылка действует 24 часа.</p>`,
		firstName, s.baseURL, emailToken,
	)

	if err := s.mailer.Send(ctx, []string{email}, subject, body); err != nil {
		s.log.Debug("resend: failed to send confirmation email", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ConfirmAccountRestore(ctx context.Context, token string) error {
	const op = "production.ConfirmAccountRestore"

	email, err := s.tokens.GetToken(token)
	if err != nil {
		s.log.Debug("restore: failed to get token", zap.Error(err))
		if errors.Is(err, memcached.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, ErrTokenNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.repository.RestoreUserByEmail(ctx, email); err != nil {
		s.log.Debug("restore: failed to restore user", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.tokens.DeleteToken(token); err != nil {
		s.log.Warn("restore: failed to delete used token", zap.Error(err))
	}

	return nil
}

func (s *Service) RestoreUser(ctx context.Context, id string) error {
	const op = "production.RestoreUser"

	if _, err := uuid.Parse(id); err != nil {
		s.log.Debug("restore user: bad user id", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if err := s.repository.RestoreUser(ctx, id); err != nil {
		s.log.Debug("restore user: failed", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GetMe(ctx context.Context, id string) (*dto.UserProfile, error) {
	const op = "production.GetMe"

	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		s.log.Debug("get me: failed to get user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Service) RequestAccountRestore(ctx context.Context, email string) error {
	const op = "production.RequestAccountRestore"

	restoreToken := generateSecureToken(64)

	if err := s.tokens.SetToken(restoreToken, email, 1*time.Hour); err != nil {
		s.log.Error("restore: failed to store token", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	subject := "Восстановление аккаунта"
	body := fmt.Sprintf(
		`<p>Поступил запрос на восстановление вашего аккаунта.</p>`+
			`<p><a href="%s/restore-account?token=%s">Восстановить аккаунт</a></p>`+
			`<p>Если это были не вы — проигнорируйте письмо. Ссылка действует 1 час.</p>`,
		s.baseURL, restoreToken,
	)

	if err := s.mailer.Send(ctx, []string{email}, subject, body); err != nil {
		s.log.Debug("restore: failed to send email", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
