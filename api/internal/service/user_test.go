package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"
	repomocks "go-ecommerce-app/internal/repository/mocks"
	"go-ecommerce-app/internal/service"
	emailmocks "go-ecommerce-app/internal/service/mocks"
	"go-ecommerce-app/pkg/security"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:          "test-jwt-secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 168 * time.Hour,
	}
}

func randomUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

func newTestService(t *testing.T) (*service.UserService, *repomocks.MockQuerier, *emailmocks.MockEmailSender) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	mockEmail := emailmocks.NewMockEmailSender(ctrl)
	svc := service.NewUserService(mockRepo, newTestConfig(), mockEmail)
	return svc, mockRepo, mockEmail
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(repo *repomocks.MockQuerier, email *emailmocks.MockEmailSender)
		assertErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(repo *repomocks.MockQuerier, email *emailmocks.MockEmailSender) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Return(repository.User{ID: randomUUID(), Email: "jane@example.com", FullName: "Jane Doe"}, nil)
				repo.EXPECT().CreateVericationEmail(gomock.Any(), gomock.Any()).
					Return(repository.EmailVerificationToken{}, nil)
				email.EXPECT().SendVerificationEmail(gomock.Any(), "jane@example.com", gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "email already taken",
			setup: func(repo *repomocks.MockQuerier, email *emailmocks.MockEmailSender) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Return(repository.User{}, &pgconn.PgError{Code: "23505"})
				// CreateVericationEmail and SendVerificationEmail deliberately NOT EXPECT()'d.
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrEmailTaken)
			},
		},
		{
			name: "email sending fails",
			setup: func(repo *repomocks.MockQuerier, email *emailmocks.MockEmailSender) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Return(repository.User{ID: randomUUID(), Email: "jane@example.com"}, nil)
				repo.EXPECT().CreateVericationEmail(gomock.Any(), gomock.Any()).
					Return(repository.EmailVerificationToken{}, nil)
				email.EXPECT().SendVerificationEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("smtp down"))
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo, mockEmail := newTestService(t)
			tt.setup(mockRepo, mockEmail)

			req := model.RegisterUserRequest{
				FullName: "Jane Doe",
				Email:    "jane@example.com",
				Password: "supersecretpassword",
			}

			user, err := svc.Register(context.Background(), req)

			if tt.assertErr != nil {
				tt.assertErr(t, err)
				assert.Nil(t, user)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, user)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	const rawPassword = "correct-horse-battery-staple"
	hashed, err := security.HashPassword(rawPassword)
	require.NoError(t, err)

	verifiedUser := repository.User{
		ID:              randomUUID(),
		Email:           "jane@example.com",
		PasswordHash:    hashed,
		Role:            "customer",
		EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	tests := []struct {
		name    string
		req     model.LoginUserRequest
		setup   func(repo *repomocks.MockQuerier)
		wantOK  bool
		wantErr error
	}{
		{
			name: "success",
			// Mixed case + surrounding whitespace exercises the trim/lowercase
			// normalization before hitting the repo.
			req: model.LoginUserRequest{Email: "  Jane@Example.com  ", Password: rawPassword},
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(verifiedUser, nil)
				repo.EXPECT().CreateRefreshToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantOK: true,
		},
		{
			name: "wrong password",
			req:  model.LoginUserRequest{Email: verifiedUser.Email, Password: "totally-wrong-password"},
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), verifiedUser.Email).Return(verifiedUser, nil)
			},
		},
		{
			name: "user not verified",
			req:  model.LoginUserRequest{Email: "unverified@example.com", Password: rawPassword},
			setup: func(repo *repomocks.MockQuerier) {
				unverified := verifiedUser
				unverified.Email = "unverified@example.com"
				unverified.EmailVerifiedAt = pgtype.Timestamptz{}
				repo.EXPECT().GetUserByEmail(gomock.Any(), "unverified@example.com").Return(unverified, nil)
			},
			wantErr: service.ErrUserNotVerified,
		},
		{
			name: "user not found",
			req:  model.LoginUserRequest{Email: "ghost@example.com", Password: rawPassword},
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "ghost@example.com").Return(repository.User{}, errors.New("no rows in result set"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo, _ := newTestService(t)
			tt.setup(mockRepo)

			user, accessToken, refreshToken, err := svc.Login(context.Background(), tt.req)

			if tt.wantOK {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
				return
			}

			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestUserService_Refresh(t *testing.T) {
	t.Run("normal rotation", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		userID := randomUUID()
		existingToken := repository.RefreshToken{
			UserID:    userID,
			TokenHash: security.HashRefreshToken("raw-refresh-token"),
		}

		mockRepo.EXPECT().GetRefreshToken(gomock.Any(), existingToken.TokenHash).Return(existingToken, nil)
		mockRepo.EXPECT().RevokeRefreshToken(gomock.Any(), existingToken.TokenHash).Return(nil)
		mockRepo.EXPECT().CreateRefreshToken(gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(repository.User{ID: userID, Role: "customer"}, nil)

		accessToken, newRawToken, err := svc.Refresh(context.Background(), "raw-refresh-token")

		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, newRawToken)
	})

	t.Run("refresh token reuse is detected and revokes every session", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		userID := randomUUID()
		tokenHash := security.HashRefreshToken("stolen-raw-refresh-token")
		revokedToken := repository.RefreshToken{
			UserID:    userID,
			TokenHash: tokenHash,
			RevokedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		mockRepo.EXPECT().GetRefreshToken(gomock.Any(), tokenHash).Return(repository.RefreshToken{}, errors.New("not found"))
		mockRepo.EXPECT().GetRefreshTokenAnyStatus(gomock.Any(), tokenHash).Return(revokedToken, nil)
		// gomock defaults to Times(1): this fails the test automatically if
		// the reuse-detection branch doesn't actually call it.
		mockRepo.EXPECT().RevokeRefreshTokenByUserID(gomock.Any(), userID).Return(nil)

		_, _, err := svc.Refresh(context.Background(), "stolen-raw-refresh-token")

		require.Error(t, err)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	const oldPassword = "the-old-password-123"
	oldHash, err := security.HashPassword(oldPassword)
	require.NoError(t, err)

	userID := uuid.New()
	existingUser := repository.User{
		ID:           pgtype.UUID{Bytes: userID, Valid: true},
		PasswordHash: oldHash,
	}

	t.Run("success", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		mockRepo.EXPECT().GetUserByID(gomock.Any(), existingUser.ID).Return(existingUser, nil)
		mockRepo.EXPECT().UpdatePasswordUser(gomock.Any(), gomock.Any()).Return(existingUser, nil)
		mockRepo.EXPECT().RevokeRefreshTokenByUserID(gomock.Any(), existingUser.ID).Return(nil)

		err := svc.ChangePassword(context.Background(), model.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: "brand-new-password-456",
		}, userID)

		require.NoError(t, err)
	})

	t.Run("wrong old password is rejected without touching the password", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		mockRepo.EXPECT().GetUserByID(gomock.Any(), existingUser.ID).Return(existingUser, nil)
		// UpdatePasswordUser deliberately NOT EXPECT()'d: if the code regresses
		// and calls it anyway, gomock fails the test.

		err := svc.ChangePassword(context.Background(), model.ChangePasswordRequest{
			OldPassword: "totally-wrong-old-password",
			NewPassword: "brand-new-password-456",
		}, userID)

		assert.ErrorIs(t, err, service.ErrInvalidOldPassword)
	})
}

func TestUserService_ResendVerification(t *testing.T) {
	t.Run("already verified user is not re-sent a token", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		verifiedUser := repository.User{
			Email:           "jane@example.com",
			EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}
		mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(verifiedUser, nil)
		// CreateVericationEmail and SendVerificationEmail deliberately NOT
		// EXPECT()'d - this is the regression guard for the "already verified
		// users still get re-sent a token" bug.

		err := svc.ResendVerification(context.Background(), "jane@example.com")

		assert.Error(t, err)
	})

	t.Run("unverified user gets a new token", func(t *testing.T) {
		svc, mockRepo, mockEmail := newTestService(t)

		unverifiedUser := repository.User{
			ID:    randomUUID(),
			Email: "jane@example.com",
		}
		mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(unverifiedUser, nil)
		mockRepo.EXPECT().CreateVericationEmail(gomock.Any(), gomock.Any()).Return(repository.EmailVerificationToken{}, nil)
		mockEmail.EXPECT().SendVerificationEmail(gomock.Any(), "jane@example.com", gomock.Any()).Return(nil)

		err := svc.ResendVerification(context.Background(), "jane@example.com")

		require.NoError(t, err)
	})
}

func TestUserService_ForgotPassword(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		svc, mockRepo, mockEmail := newTestService(t)

		user := repository.User{ID: randomUUID(), Email: "jane@example.com"}
		mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(user, nil)
		mockRepo.EXPECT().CreateResetPasswordToken(gomock.Any(), gomock.Any()).Return(repository.PasswordResetToken{}, nil)
		mockEmail.EXPECT().SendPasswordResetEmail(gomock.Any(), "jane@example.com", gomock.Any()).Return(nil)

		err := svc.ForgotPassword(context.Background(), "jane@example.com")

		require.NoError(t, err)
	})
}

func TestUserService_ResetPassword(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		userID := randomUUID()
		tokenHash := security.HashRefreshToken("valid-raw-reset-token")
		resetToken := repository.PasswordResetToken{UserID: userID, TokenHash: tokenHash}

		mockRepo.EXPECT().GetResetPasswordTokenByTokenHash(gomock.Any(), tokenHash).Return(resetToken, nil)
		mockRepo.EXPECT().UpdatePasswordUser(gomock.Any(), gomock.Any()).Return(repository.User{ID: userID}, nil)
		mockRepo.EXPECT().RevokeRefreshTokenByUserID(gomock.Any(), userID).Return(nil)
		mockRepo.EXPECT().DeletePasswordTokenByTokenHash(gomock.Any(), tokenHash).Return(nil)

		err := svc.ResetPassword(context.Background(), "valid-raw-reset-token", "brand-new-password-789")

		require.NoError(t, err)
	})

	t.Run("expired or invalid token never touches the password", func(t *testing.T) {
		svc, mockRepo, _ := newTestService(t)

		tokenHash := security.HashRefreshToken("expired-raw-reset-token")
		mockRepo.EXPECT().GetResetPasswordTokenByTokenHash(gomock.Any(), tokenHash).Return(repository.PasswordResetToken{}, errors.New("not found"))
		// UpdatePasswordUser deliberately NOT EXPECT()'d.

		err := svc.ResetPassword(context.Background(), "expired-raw-reset-token", "brand-new-password-789")

		assert.Error(t, err)
	})
}
