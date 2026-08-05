package service

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}
