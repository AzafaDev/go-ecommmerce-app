package model

import (
	"time"

	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	FullName string `json:"full_name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	FullName        string     `json:"full_name"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=12"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=customer admin"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}
