package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/relationskat/auth-service/internal/models/dto"
	"go.uber.org/zap"
)

type Store struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func New(pool *pgxpool.Pool, log *zap.Logger) *Store {
	return &Store{
		pool: pool,
		log:  log.Named("store"),
	}
}

func (s *Store) CreateUser(ctx context.Context, user *dto.User) error {
	const op = "CreateUser"

	_, err := s.pool.Exec(ctx, createUserQuery,
		user.ID,
		user.LastName,
		user.FirstName,
		user.MiddleName,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			s.log.Debug("failed to added user in db, user alredy exists")
			return ErrUserAlreadyExists
		}
		s.log.Debug("failed to added user in db", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Debug("user added in db", zap.String("user_id", user.ID))

	return nil
}

func (s *Store) AcceptEmail(ctx context.Context, email string) error {
	const op = "AcceptEmail"

	tag, err := s.pool.Exec(ctx, acceptEmailQuery, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: email not found", op)
	}

	return nil
}

func (s *Store) Login(ctx context.Context, email string) (string, string, error) {
	const op = "Login"

	var ID, passwordHash string

	err := s.pool.QueryRow(ctx, loginQuery, email).Scan(&ID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return ID, passwordHash, nil
}

func (s *Store) Update(ctx context.Context, user *dto.UserUpdate) (*dto.UserUpdate, error) {
	const op = "UpdateUser"

	var res dto.UserUpdate
	err := s.pool.QueryRow(ctx, updateUserQuery,
		user.ID, user.LastName, user.FirstName, user.MiddleName,
	).Scan(&res.ID, &res.LastName, &res.FirstName, &res.MiddleName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &res, nil
}

func (s *Store) GetPasswordHash(ctx context.Context, id string) (string, error) {
	const op = "GetPasswordHash"

	var hash string
	err := s.pool.QueryRow(ctx, getPasswordHashByIDQuery, id).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return hash, nil
}

func (s *Store) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	const op = "UpdatePassword"

	tag, err := s.pool.Exec(ctx, updatePasswordByIDQuery, id, passwordHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) UpdatePasswordByEmail(ctx context.Context, email, passwordHash string) error {
	const op = "UpdatePasswordByEmail"

	tag, err := s.pool.Exec(ctx, updatePasswordByEmailQuery, email, passwordHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) DeleteUser(ctx context.Context, id string) error {
	const op = "DeleteUser"

	tag, err := s.pool.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) GetEmailStatus(ctx context.Context, email string) (string, bool, error) {
	const op = "GetEmailStatus"

	var (
		firstName string
		confirmed bool
	)
	err := s.pool.QueryRow(ctx, getEmailStatusQuery, email).Scan(&firstName, &confirmed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, ErrUserNotFound
		}
		return "", false, fmt.Errorf("%s: %w", op, err)
	}

	return firstName, confirmed, nil
}

func (s *Store) RestoreUserByEmail(ctx context.Context, email string) error {
	const op = "RestoreUserByEmail"

	tag, err := s.pool.Exec(ctx, restoreUserByEmailQuery, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) RestoreUser(ctx context.Context, id string) error {
	const op = "RestoreUser"

	tag, err := s.pool.Exec(ctx, restoreUserByIDQuery, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*dto.UserProfile, error) {
	const op = "GetUserByID"

	var u dto.UserProfile
	err := s.pool.QueryRow(ctx, getUserByIDQuery, id).Scan(
		&u.ID, &u.LastName, &u.FirstName, &u.MiddleName, &u.Email, &u.EmailConfirmed,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &u, nil
}
