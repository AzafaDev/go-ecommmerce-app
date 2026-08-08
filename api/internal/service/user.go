package service

import (
	"context"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/pkg/security"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	repo  repository.Querier
	cfg   *config.Config
	email EmailSender
}

func NewUserService(repo repository.Querier, cfg *config.Config, email EmailSender) *UserService {
	return &UserService{
		repo:  repo,
		cfg:   cfg,
		email: email,
	}
}

func toUserResponse(u repository.User) *model.UserResponse {
	var emailVerifiedAt *time.Time
	if u.EmailVerifiedAt.Valid {
		t := u.EmailVerifiedAt.Time
		emailVerifiedAt = &t
	}

	return &model.UserResponse{
		ID:              u.ID.Bytes,
		FullName:        u.FullName,
		Email:           u.Email,
		Role:            u.Role,
		CreatedAt:       u.CreatedAt.Time,
		UpdatedAt:       u.UpdatedAt.Time,
		EmailVerifiedAt: emailVerifiedAt,
	}
}

func (u *UserService) Register(ctx context.Context, req model.RegisterUserRequest) (*model.UserResponse, error) {
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	trimEmail := strings.ToLower(strings.TrimSpace(req.Email))

	createdUser, err := u.repo.CreateUser(ctx, repository.CreateUserParams{
		FullName:     req.FullName,
		Email:        trimEmail,
		PasswordHash: passwordHash,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("error in creating user: %w", err)
	}

	randomString, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("error in generatin random string token: %w", err)
	}
	hashedToken := security.HashRefreshToken(randomString)

	_, err = u.repo.CreateVericationEmail(ctx, repository.CreateVericationEmailParams{
		UserID:    createdUser.ID,
		TokenHash: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error in creating verification email token: %w", err)
	}

	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := u.email.SendVerificationEmail(emailCtx, createdUser.Email, randomString); err != nil {
			slog.Error("failed to send verification email", "error", err, "user_id", createdUser.ID)
		}
	}()

	return toUserResponse(createdUser), nil
}

func (u *UserService) Login(ctx context.Context, req model.LoginUserRequest) (*model.UserResponse, string, string, error) {
	trimEmail := strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := u.repo.GetUserByEmail(ctx, trimEmail)
	if err != nil {
		slog.Warn("security_event", "event", "login_failed", "reason", "invalid_credentials", "email", trimEmail)
		return nil, "", "", fmt.Errorf("error in getting user by email: %w", err)
	}

	if err := security.ComparePassword(existingUser.PasswordHash, req.Password); err != nil {
		slog.Warn("security_event", "event", "login_failed", "reason", "invalid_credentials", "email", trimEmail)
		return nil, "", "", fmt.Errorf("error in comparing password: %w", err)
	}

	if !existingUser.EmailVerifiedAt.Valid {
		slog.Warn("security_event", "event", "login_failed", "reason", "email_not_verified", "email", trimEmail)
		return nil, "", "", ErrUserNotVerified
	}

	signedToken, err := security.GenerateToken(u.cfg.JWTSecret, u.cfg.JWTExpiry, existingUser.ID.Bytes, existingUser.Role)

	if err != nil {
		return nil, "", "", fmt.Errorf("error in generating token JWT: %w", err)
	}

	rawToken, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("error in generating refresh token: %w", err)
	}
	hashedToken := security.HashRefreshToken(rawToken)

	if err := u.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    existingUser.ID,
		TokenHash: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(u.cfg.RefreshTokenExpiry),
			Valid: true,
		},
	}); err != nil {
		return nil, "", "", fmt.Errorf("error in creating refresh token: %w", err)
	}

	slog.Info("security_event", "event", "login_success", "user_id", existingUser.ID)

	return toUserResponse(existingUser), signedToken, rawToken, nil
}

func (u *UserService) Refresh(ctx context.Context, rawToken string) (string, string, error) {
	tokenHash := security.HashRefreshToken(rawToken)

	existingRefreshToken, err := u.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if anyToken, lookupErr := u.repo.GetRefreshTokenAnyStatus(ctx, tokenHash); lookupErr == nil && anyToken.RevokedAt.Valid {
			slog.Error("security_event", "event", "refresh_token_reuse_detected", "user_id", anyToken.UserID)
			if revokeErr := u.repo.RevokeRefreshTokenByUserID(ctx, anyToken.UserID); revokeErr != nil {
				return "", "", fmt.Errorf("error in revoking refresh tokens after reuse detected: %w", revokeErr)
			}
		}
		return "", "", fmt.Errorf("error in get refresh token: %w", err)
	}

	if err := u.repo.RevokeRefreshToken(ctx, existingRefreshToken.TokenHash); err != nil {
		return "", "", fmt.Errorf("error in get refresh token: %w", err)
	}

	newRawToken, err := security.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("error in generating refresh token: %w", err)
	}
	hashedToken := security.HashRefreshToken(newRawToken)

	if err := u.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    existingRefreshToken.UserID,
		TokenHash: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(u.cfg.RefreshTokenExpiry),
			Valid: true,
		},
	}); err != nil {
		return "", "", fmt.Errorf("error in creating refresh token: %w", err)
	}

	existingUser, err := u.repo.GetUserByID(ctx, existingRefreshToken.UserID)
	if err != nil {
		return "", "", fmt.Errorf("error in getting user by id: %w", err)
	}

	accessToken, err := security.GenerateToken(u.cfg.JWTSecret, u.cfg.JWTExpiry, existingRefreshToken.UserID.Bytes, existingUser.Role)
	if err != nil {
		return "", "", fmt.Errorf("error in generating token: %w", err)
	}

	return accessToken, newRawToken, nil
}

func (u *UserService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := security.HashRefreshToken(rawToken)

	if err := u.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("error in revoking refresh token: %w", err)
	}

	return nil
}

func (u *UserService) UpdateRole(ctx context.Context, userID pgtype.UUID, role string) (*model.UserResponse, error) {
	updatedUser, err := u.repo.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
		Role: role,
		ID:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("error in updating user role: %w", err)
	}

	return toUserResponse(updatedUser), nil
}

func (u *UserService) Me(ctx context.Context, userID pgtype.UUID) (*model.UserResponse, error) {
	existingUser, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error in getting user by id: %w", err)
	}

	return toUserResponse(existingUser), nil
}

func (u *UserService) LogoutAllDevices(ctx context.Context, userID pgtype.UUID) error {
	if err := u.repo.RevokeRefreshTokenByUserID(ctx, userID); err != nil {
		return fmt.Errorf("error in revoking all refresh tokens: %w", err)
	}

	return nil
}

func (u *UserService) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := security.HashRefreshToken(rawToken)

	existingEmailToken, err := u.repo.GetEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("error in getting email verification token: %w", err)
	}

	_, err = u.repo.SetUserVerified(ctx, existingEmailToken.UserID)
	if err != nil {
		return fmt.Errorf("error in verifying email: %w", err)
	}

	if err := u.repo.DeleteEmailVerificationByTokenHash(ctx, existingEmailToken.TokenHash); err != nil {
		return fmt.Errorf("error in deleting email verification token: %w", err)
	}

	return nil
}

func (u *UserService) ResendVerification(ctx context.Context, email string) error {
	trimEmail := strings.ToLower(strings.TrimSpace(email))

	existingUser, err := u.repo.GetUserByEmail(ctx, trimEmail)
	if err != nil {
		return fmt.Errorf("error in getting user by email: %w", err)
	}

	if existingUser.EmailVerifiedAt.Valid {
		return fmt.Errorf("user is already verified")
	}

	randomString, err := security.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("error in generatin random string token: %w", err)
	}
	hashedToken := security.HashRefreshToken(randomString)

	_, err = u.repo.CreateVericationEmail(ctx, repository.CreateVericationEmailParams{
		UserID:    existingUser.ID,
		TokenHash: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("error in creating verification email token: %w", err)
	}

	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.email.SendVerificationEmail(emailCtx, existingUser.Email, randomString); err != nil {
			slog.Error("failed to resend verification email", "error", err, "user_id", existingUser.ID)
		}
	}()
	return nil
}

func (u *UserService) ForgotPassword(ctx context.Context, email string) error {
	trimEmail := strings.ToLower(strings.TrimSpace(email))

	existingUser, err := u.repo.GetUserByEmail(ctx, trimEmail)
	if err != nil {
		return fmt.Errorf("error in getting user by email: %w", err)
	}

	randomString, err := security.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("error in generatin random string token: %w", err)
	}
	hashedToken := security.HashRefreshToken(randomString)

	_, err = u.repo.CreateResetPasswordToken(ctx, repository.CreateResetPasswordTokenParams{
		UserID:    existingUser.ID,
		TokenHash: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(15 * time.Minute),
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("error in creating verification email token: %w", err)
	}

	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.email.SendPasswordResetEmail(emailCtx, existingUser.Email, randomString); err != nil {
			slog.Error("failed to send password reset email", "error", err, "user_id", existingUser.ID)
		}
	}()

	slog.Info("security_event", "event", "password_reset_requested", "user_id", existingUser.ID)

	return nil
}

func (u *UserService) ResetPassword(ctx context.Context, rawToken, password string) error {
	tokenHash := security.HashRefreshToken(rawToken)

	exisitngPasswordToken, err := u.repo.GetResetPasswordTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("error in getting email verification token: %w", err)
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("error in hashing password: %w", err)
	}

	_, err = u.repo.UpdatePasswordUser(ctx, repository.UpdatePasswordUserParams{
		PasswordHash: passwordHash,
		ID:           exisitngPasswordToken.UserID,
	})
	if err != nil {
		return fmt.Errorf("error in updating password user: %w", err)
	}

	if err := u.LogoutAllDevices(ctx, exisitngPasswordToken.UserID); err != nil {
		return err
	}

	if err := u.repo.DeletePasswordTokenByTokenHash(ctx, exisitngPasswordToken.TokenHash); err != nil {
		return fmt.Errorf("error in deleting pasword token: %w", err)
	}

	slog.Info("security_event", "event", "password_reset_succeeded", "user_id", exisitngPasswordToken.UserID)

	return nil
}

func (u *UserService) ChangePassword(ctx context.Context, req model.ChangePasswordRequest, userID uuid.UUID) error {

	existingUser, err := u.repo.GetUserByID(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})

	if err != nil {
		return fmt.Errorf("error in getting use by id: %w", err)
	}

	if err := security.ComparePassword(existingUser.PasswordHash, req.OldPassword); err != nil {
		slog.Warn("security_event", "event", "change_password_failed", "reason", "invalid_old_password", "user_id", existingUser.ID)
		return ErrInvalidOldPassword
	}

	newPasswordHash, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("error in hashing new password: %w", err)
	}

	updatedUser, err := u.repo.UpdatePasswordUser(ctx, repository.UpdatePasswordUserParams{
		PasswordHash: newPasswordHash,
		ID:           existingUser.ID,
	})

	if err != nil {
		return fmt.Errorf("error in updating password user: %w", err)
	}

	if err := u.LogoutAllDevices(ctx, updatedUser.ID); err != nil {
		return fmt.Errorf("error in logout all devices service method: %w", err)
	}

	slog.Info("security_event", "event", "password_changed", "user_id", updatedUser.ID)

	return nil
}
