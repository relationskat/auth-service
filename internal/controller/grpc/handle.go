package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/relationskat/auth-service/internal/domain"
	"github.com/relationskat/auth-service/internal/models/dto"
	"github.com/relationskat/auth-service/internal/pkg/provider"
	"github.com/relationskat/auth-service/internal/repository/store"
	"github.com/relationskat/auth-service/internal/service/production"
	authv1 "github.com/relationskat/auth-service/pkg/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (grpc *GRPC) Register(ctx context.Context, in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if in.GetLastName() == "" || in.GetFirstName() == "" {
		return nil, status.Error(codes.InvalidArgument, "last_name and first_name are required")
	}
	if in.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	req := &domain.RegisterRequest{
		LastName:   in.GetLastName(),
		FirstName:  in.GetFirstName(),
		MiddleName: in.MiddleName,
		Email:      in.GetEmail(),
		Password:   in.GetPassword(),
	}

	resp, err := grpc.service.CreateUser(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			grpc.log.Error("register failed", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.RegisterResponse{
		Id:    resp.ID,
		Email: resp.Email,
	}, nil
}

func (grpc *GRPC) AcceptEmail(ctx context.Context, in *authv1.AcceptEmailRequest) (*authv1.Response, error) {
	const op = "grpc.AcceptEmail"

	if in.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	ok, err := grpc.service.AcceptEmail(ctx, in.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, production.ErrTokenNotFound):
			grpc.log.Debug("token not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "token not found or expired")
		case errors.Is(err, production.ErrUserNotFound):
			grpc.log.Debug("user not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			grpc.log.Error("accept email failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	if !ok {
		return nil, status.Error(codes.NotFound, "token validation failed")
	}

	grpc.log.Info("email successfully confirmed", zap.String("token", in.GetToken()))
	return &authv1.Response{
		Ok: true,
	}, nil
}

func (grpc *GRPC) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	const op = "grpc.Login"

	if in.GetEmail() == "" || in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if !strings.Contains(in.GetEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "email should be valid")
	}

	accessToken, refreshToken, err := grpc.service.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, store.ErrUserNotFound),
			errors.Is(err, production.ErrInvalidCredentials):
			grpc.log.Debug("login failed", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		default:
			grpc.log.Error("login failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (grpc *GRPC) RefreshToken(ctx context.Context, in *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	const op = "grpc.RefreshToken"

	if in.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	accessToken, refreshToken, err := grpc.service.RefreshToken(ctx, in.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, provider.ErrTokenExpired):
			grpc.log.Debug("refresh token expired", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "refresh token expired")
		case errors.Is(err, provider.ErrTokenInvalid):
			grpc.log.Debug("invalid refresh token", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		default:
			grpc.log.Error("refresh failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (grpc *GRPC) Auth(ctx context.Context, in *authv1.AuthRequest) (*authv1.AuthResponse, error) {
	const op = "grpc.Auth"

	if in.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	id, err := grpc.service.Auth(ctx, in.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, provider.ErrTokenExpired):
			return nil, status.Error(codes.Unauthenticated, "token expired")
		case errors.Is(err, provider.ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, "token invalid")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.AuthResponse{Id: id}, nil
}

func (grpc *GRPC) Update(ctx context.Context, in *authv1.UpdateRequest) (*authv1.UpdateResponse, error) {
	const op = "grpc.Update"

	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user := &dto.UserUpdate{
		ID:         in.GetId(),
		LastName:   in.GetLastName(),
		FirstName:  in.GetFirstName(),
		MiddleName: in.MiddleName,
	}

	updated, err := grpc.service.Update(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	resp := &authv1.UpdateResponse{
		Id:         updated.ID,
		MiddleName: updated.MiddleName,
		LastName:   updated.LastName,
		FirstName:  updated.FirstName,
	}

	return resp, nil
}

func (grpc *GRPC) ChangePassword(ctx context.Context, in *authv1.ChangePasswordRequest) (*authv1.Response, error) {
	const op = "grpc.ChangePassword"

	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if in.GetOldPassword() == "" || in.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "old_password and new_passord are required")
	}
	if len(in.GetNewPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "new_password must be at 8 characters")
	}

	if err := grpc.service.ChangePassword(ctx, in.GetId(), in.GetOldPassword(),
		in.GetNewPassword()); err != nil {
		switch {
		case errors.Is(err, production.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, production.ErrSamePassword):
			return nil, status.Error(codes.InvalidArgument, "new password should be diffrient")
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			grpc.log.Error("change password failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) DeleteUser(ctx context.Context, in *authv1.DeleteUserRequest) (*authv1.Response, error) {
	const op = "grpc.DeleteUser"

	if in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := grpc.service.DeleteUser(ctx, in.GetId()); err != nil {
		switch {
		case errors.Is(err, production.ErrInvalidCredentials):
			return nil, status.Error(codes.InvalidArgument, "invalid id")
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			grpc.log.Error("delete user failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) ResendConfirmationEmail(ctx context.Context, in *authv1.ResendConfirmationEmailRequest) (*authv1.Response, error) {
	const op = "grpc.ResendConfirmationEmail"

	if in.GetEmail() == "" || !strings.Contains(in.GetEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}

	if err := grpc.service.ResendConfirmationEmail(ctx, in.GetEmail()); err != nil {
		switch {
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, production.ErrEmailAlreadyConfirmed):
			return nil, status.Error(codes.FailedPrecondition, "email already confirmed")
		default:
			grpc.log.Error("resend confirmation email failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) RequestAccountRestore(ctx context.Context, in *authv1.RequestAccountRestoreRequest) (*authv1.Response, error) {
	const op = "grpc.RequestAccountRestore"

	if in.GetEmail() == "" || !strings.Contains(in.GetEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}

	if err := grpc.service.RequestAccountRestore(ctx, in.GetEmail()); err != nil {
		grpc.log.Error("request account restore failed", zap.String("op", op), zap.Error(err))
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) ConfirmAccountRestore(ctx context.Context, in *authv1.ConfirmAccountRestoreRequest) (*authv1.Response, error) {
	const op = "grpc.ConfirmAccountRestore"

	if in.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := grpc.service.ConfirmAccountRestore(ctx, in.GetToken()); err != nil {
		switch {
		case errors.Is(err, production.ErrTokenNotFound):
			return nil, status.Error(codes.NotFound, "restore token not found or expired")
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "deleted account not found")
		default:
			grpc.log.Error("confirm account restore failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) RequestPasswordReset(ctx context.Context, in *authv1.RequestPasswordResetRequest) (*authv1.Response, error) {
	const op = "grpc.RequestPasswordReset"

	if in.GetEmail() == "" || !strings.Contains(in.GetEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}

	if err := grpc.service.RequestPasswordReset(ctx, in.GetEmail()); err != nil {
		grpc.log.Error("request password reset failed", zap.String("op", op), zap.Error(err))
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) ConfirmPasswordReset(ctx context.Context, in *authv1.ConfirmPasswordResetRequest) (*authv1.Response, error) {
	const op = "grpc.ConfirmPasswordReset"

	if in.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}
	if len(in.GetNewPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "new_password must be at least 8 characters")
	}

	if err := grpc.service.ConfirmPasswordReset(ctx, in.GetToken(), in.GetNewPassword()); err != nil {
		switch {
		case errors.Is(err, production.ErrTokenNotFound):
			return nil, status.Error(codes.NotFound, "reset token not found or expired")
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			grpc.log.Error("confirm password reset failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.Response{Ok: true}, nil
}

func (grpc *GRPC) GetMe(ctx context.Context, in *authv1.GetMeRequest) (*authv1.GetMeResponse, error) {
	const op = "grpc.GetMe"

	if in.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := grpc.service.GetMe(ctx, in.Id)
	if err != nil {
		switch {
		case errors.Is(err, provider.ErrTokenExpired):
			return nil, status.Error(codes.Unauthenticated, "token expired")
		case errors.Is(err, provider.ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, "token invalid")
		case errors.Is(err, store.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			grpc.log.Error("get me failed", zap.String("op", op), zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &authv1.GetMeResponse{
		Id:             user.ID,
		LastName:       user.LastName,
		FirstName:      user.FirstName,
		MiddleName:     user.MiddleName,
		Email:          user.Email,
		EmailConfirmed: user.EmailConfirmed,
	}, nil
}
