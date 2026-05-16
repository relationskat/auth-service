package repository

import (
	"context"

	"github.com/relationskat/auth-service/internal/models/dto"
)

type Repository interface {
	CreateUser(ctx context.Context, user *dto.User) error
	Login(ctx context.Context, email string) (string, string, error)
	AcceptEmail(ctx context.Context, email string) error
	Update(ctx context.Context, user *dto.UserUpdate) (*dto.UserUpdate, error)
	GetPasswordHash(ctx context.Context, id string) (string, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	UpdatePasswordByEmail(ctx context.Context, email, passwordHash string) error
	DeleteUser(ctx context.Context, id string) error
	GetEmailStatus(ctx context.Context, email string) (string, bool, error)
	RestoreUserByEmail(ctx context.Context, email string) error
	RestoreUser(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (*dto.UserProfile, error)
}
