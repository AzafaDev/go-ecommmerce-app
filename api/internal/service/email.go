package service

import "context"

//go:generate go tool mockgen -source=email.go -destination=mocks/email_mock.go -package=emailmocks
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}
