package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/response"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	srv                *service.UserService
	validate           *validator.Validate
	jwtSecretKey       string
	refreshTokenExpiry time.Duration
	rdb                *redis.Client
	env                string
}

func NewUserHandler(srv *service.UserService, secret string, refreshTokenExpiry time.Duration, rdb *redis.Client, env string) *UserHandler {
	return &UserHandler{
		srv:                srv,
		validate:           validator.New(),
		jwtSecretKey:       secret,
		refreshTokenExpiry: refreshTokenExpiry,
		rdb:                rdb,
		env:                env,
	}
}

func (u *UserHandler) refreshTokenCookie(value string) http.Cookie {
	return http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(u.refreshTokenExpiry),
		MaxAge:   int(u.refreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   u.env == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

func (u *UserHandler) clearRefreshTokenCookie() http.Cookie {
	return http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   u.env == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

func (u *UserHandler) UserRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.With(middleware.RateLimit(u.rdb, 5, 15*time.Minute, middleware.IPAndEmailKeyFunc("login"))).
			Post("/login", u.Login)
		r.With(middleware.RateLimit(u.rdb, 3, time.Hour, middleware.IPKeyFunc("register"))).
			Post("/register", u.Register)
		r.Post("/logout", u.Logout)
		r.Post("/refresh", u.Refresh)
		r.Post("/verify-email", u.VerifyEmail)
		r.With(middleware.RateLimit(u.rdb, 3, time.Hour, middleware.EmailKeyFunc("resend-verification"))).
			Post("/resend-verification", u.ResendVerification)
		r.With(middleware.RateLimit(u.rdb, 3, time.Hour, middleware.EmailKeyFunc("forgot-password"))).
			Post("/forgot-password", u.ForgotPassword)
		r.Post("/reset-password", u.ResetPassword)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(u.jwtSecretKey))
			r.Get("/me", u.Me)
			r.With(middleware.RateLimit(u.rdb, 5, 15*time.Minute, middleware.UserIDKeyFunc("change-password"))).
				Post("/change-password", u.ChangePassword)
			r.Post("/logout-all", u.LogoutAllDevices)
			r.With(middleware.RequireRole("admin")).Patch("/{id}/role", u.UpdateRole)
		})
	})
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("register", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("register", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	createdUser, err := u.srv.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			slog.Error("register", "error", err)
			response.WriteErrorJSON(err.Error(), http.StatusConflict, w)
			return
		}
		slog.Error("register", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"user": createdUser,
		},
	}, http.StatusCreated, w)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("login", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("login", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	existingUser, accessToken, refreshToken, err := u.srv.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotVerified) {
			slog.Error("login", "error", err)
			response.WriteErrorJSON("please verify your email", http.StatusForbidden, w)
			return
		}
		slog.Error("login", "error", err)
		response.WriteErrorJSON("invalid email or password", http.StatusBadRequest, w)
		return
	}

	cookie := u.refreshTokenCookie(refreshToken)
	http.SetCookie(w, &cookie)

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"accessToken": accessToken,
			"user":        existingUser,
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		slog.Error("refresh", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	rawRefreshToken := refreshTokenCookie.Value
	if rawRefreshToken == "" {
		slog.Error("refresh", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	accessToken, newRawRefreshToken, err := u.srv.Refresh(r.Context(), rawRefreshToken)
	if err != nil {
		slog.Error("refresh", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	cookie := u.refreshTokenCookie(newRawRefreshToken)
	http.SetCookie(w, &cookie)

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"accessToken": accessToken,
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := u.clearRefreshTokenCookie()
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		slog.Error("refresh", "error", err)
		http.SetCookie(w, &cookie)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]any{
				"message": "logout successfully",
			},
		}, http.StatusOK, w)
		return
	}

	rawRefreshToken := refreshTokenCookie.Value
	if rawRefreshToken == "" {
		slog.Error("refresh", "error", err)
		http.SetCookie(w, &cookie)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]any{
				"message": "logout successfully",
			},
		}, http.StatusOK, w)
		return
	}

	if err := u.srv.Logout(r.Context(), rawRefreshToken); err != nil {
		slog.Error("refresh", "error", err)
		http.SetCookie(w, &cookie)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]any{
				"message": "logout successfully",
			},
		}, http.StatusOK, w)
		return
	}

	http.SetCookie(w, &cookie)
	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"message": "logout successfully",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	user, err := u.srv.Me(r.Context(), pgtype.UUID{Bytes: claims.ID, Valid: true})
	if err != nil {
		slog.Error("me", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"user": user,
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		slog.Error("verify email", "error", fmt.Errorf("invalid request url query of token"))
		response.WriteErrorJSON("invalid or expired token", http.StatusUnauthorized, w)
		return
	}

	if err := u.srv.VerifyEmail(r.Context(), rawToken); err != nil {
		slog.Error("verify email", "error", err)
		response.WriteErrorJSON("invalid or expired token", http.StatusUnauthorized, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"message": "email verified successfully",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req model.ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("resend verification", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("resend verification", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	if err := u.srv.ResendVerification(r.Context(), req.Email); err != nil {
		slog.Error("resend verification", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "if that email exists, we sent a link",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("forgot password", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("forgot password", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	if err := u.srv.ForgotPassword(r.Context(), req.Email); err != nil {
		slog.Error("forgot password", "error", err)
		response.WriteJSON(response.JSONResponse{
			Success: true,
			Data: map[string]string{
				"message": "if that email exists, we sent a link",
			},
		}, http.StatusOK, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "if that email exists, we sent a link",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("reset password", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("reset password", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		slog.Error("reset password", "error", fmt.Errorf("invalid request url query of token"))
		response.WriteErrorJSON("invalid or expired token", http.StatusUnauthorized, w)
		return
	}

	if err := u.srv.ResetPassword(r.Context(), rawToken, req.Password); err != nil {
		slog.Error("reset password", "error", err)
		response.WriteErrorJSON("invalid or expired token", http.StatusUnauthorized, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"message": "reset password successfully",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	if err := u.srv.LogoutAllDevices(r.Context(), pgtype.UUID{Bytes: claims.ID, Valid: true}); err != nil {
		slog.Error("logout all devices", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"message": "logged out from all devices",
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteErrorJSON("invalid user id", http.StatusBadRequest, w)
		return
	}

	var req model.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("update role", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("update role", "error", err)
		response.WriteErrorJSON("invalid request payload", http.StatusBadRequest, w)
		return
	}

	updatedUser, err := u.srv.UpdateRole(r.Context(), pgtype.UUID{Bytes: targetID, Valid: true}, req.Role)
	if err != nil {
		slog.Error("update role", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"user": updatedUser,
		},
	}, http.StatusOK, w)
}

func (u *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("change password", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		slog.Error("change password", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("change password", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusBadRequest, w)
		return
	}

	err = u.srv.ChangePassword(r.Context(), req, claims.ID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOldPassword) {
			slog.Error("change password", "error", err)
			response.WriteErrorJSON(err.Error(), http.StatusForbidden, w)
			return
		}
		slog.Error("change password", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "changed password successfully",
		},
	}, http.StatusOK, w)
}
