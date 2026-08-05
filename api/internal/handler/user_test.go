package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/handler"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"
	repomocks "go-ecommerce-app/internal/repository/mocks"
	"go-ecommerce-app/internal/service"
	emailmocks "go-ecommerce-app/internal/service/mocks"
	"go-ecommerce-app/pkg/security"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// jsonResponse mirrors the shape both response.WriteJSON and
// response.WriteErrorJSON produce, so a single decode target works for
// success and error paths alike.
type jsonResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func newTestHandler(t *testing.T) (*handler.UserHandler, *repomocks.MockQuerier, *emailmocks.MockEmailSender) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	mockEmail := emailmocks.NewMockEmailSender(ctrl)

	cfg := &config.Config{
		JWTSecret:          "test-jwt-secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 168 * time.Hour,
	}
	svc := service.NewUserService(mockRepo, cfg, mockEmail)

	// Login/Register never touch rdb directly - rate limiting is wired as
	// separate middleware around the route, not inside the handler method -
	// so a nil *redis.Client is safe here.
	h := handler.NewUserHandler(svc, cfg.JWTSecret, cfg.RefreshTokenExpiry, nil, "test")
	return h, mockRepo, mockEmail
}

func newJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) jsonResponse {
	t.Helper()
	var got jsonResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	return got
}

func TestUserHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, mockRepo, _ := newTestHandler(t)

		const rawPassword = "correct-horse-battery-staple"
		hashed, err := security.HashPassword(rawPassword)
		require.NoError(t, err)

		mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(repository.User{
			ID:              pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Email:           "jane@example.com",
			PasswordHash:    hashed,
			Role:            "customer",
			EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)
		mockRepo.EXPECT().CreateRefreshToken(gomock.Any(), gomock.Any()).Return(nil)

		req := newJSONRequest(t, http.MethodPost, "/users/login", model.LoginUserRequest{
			Email:    "jane@example.com",
			Password: rawPassword,
		})
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotEmpty(t, got.Data["accessToken"])
		assert.NotNil(t, got.Data["user"])

		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "refresh_token", cookies[0].Name)
		assert.NotEmpty(t, cookies[0].Value)
	})

	t.Run("bad credentials return 400", func(t *testing.T) {
		h, mockRepo, _ := newTestHandler(t)

		hashed, err := security.HashPassword("the-real-password")
		require.NoError(t, err)

		mockRepo.EXPECT().GetUserByEmail(gomock.Any(), "jane@example.com").Return(repository.User{
			Email:           "jane@example.com",
			PasswordHash:    hashed,
			EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		req := newJSONRequest(t, http.MethodPost, "/users/login", model.LoginUserRequest{
			Email:    "jane@example.com",
			Password: "totally-wrong-password",
		})
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		got := decodeResponse(t, rec)
		assert.False(t, got.Success)
	})
}

func TestUserHandler_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, mockRepo, mockEmail := newTestHandler(t)

		mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{
			ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			FullName: "Jane Doe",
			Email:    "jane@example.com",
		}, nil)
		mockRepo.EXPECT().CreateVericationEmail(gomock.Any(), gomock.Any()).Return(repository.EmailVerificationToken{}, nil)
		mockEmail.EXPECT().SendVerificationEmail(gomock.Any(), "jane@example.com", gomock.Any()).Return(nil)

		req := newJSONRequest(t, http.MethodPost, "/users/register", model.RegisterUserRequest{
			FullName: "Jane Doe",
			Email:    "jane@example.com",
			Password: "supersecretpassword",
		})
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["user"])
	})

	t.Run("email already taken returns 409", func(t *testing.T) {
		h, mockRepo, _ := newTestHandler(t)

		mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{}, &pgconn.PgError{Code: "23505"})

		req := newJSONRequest(t, http.MethodPost, "/users/register", model.RegisterUserRequest{
			FullName: "Jane Doe",
			Email:    "jane@example.com",
			Password: "supersecretpassword",
		})
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
		got := decodeResponse(t, rec)
		assert.False(t, got.Success)
		assert.Equal(t, service.ErrEmailTaken.Error(), got.Message)
	})
}
