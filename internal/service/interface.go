package service

import (
	"context"

	"github.com/relationskat/auth-service/internal/domain"
	"github.com/relationskat/auth-service/internal/models/dto"
)

type Service interface {
	CreateUser(ctx context.Context, user *domain.RegisterRequest) (*domain.RegisterResponse, error)
	AcceptEmail(ctx context.Context, token string) (bool, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	RefreshToken(ctx context.Context, token string) (string, string, error)
	Auth(ctx context.Context, token string) (string, error)
	Update(ctx context.Context, user *dto.UserUpdate) (*dto.UserUpdate, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error
	DeleteUser(ctx context.Context, id string) error
	ResendConfirmationEmail(ctx context.Context, email string) error
	ConfirmAccountRestore(ctx context.Context, token string) error
	RestoreUser(ctx context.Context, id string) error
	GetMe(ctx context.Context, id string) (*dto.UserProfile, error)
	RequestAccountRestore(ctx context.Context, email string) error
}
