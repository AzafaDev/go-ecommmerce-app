package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type ResendClient struct {
	client      *resend.Client
	fromAddress string
	frontendURL string
}

func NewResendClient(apiKey, fromAddress, frontendURL string) *ResendClient {
	return &ResendClient{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		frontendURL: frontendURL,
	}
}

func (r *ResendClient) SendVerificationEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", r.frontendURL, token)

	params := &resend.SendEmailRequest{
		From:    r.fromAddress,
		To:      []string{to},
		Subject: "Verify your email",
		Html: fmt.Sprintf(
			`<p>Click the link below to verify your email. This link expires in 24 hours.</p><p><a href="%s">%s</a></p>`,
			link, link,
		),
	}

	if _, err := r.client.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("error in sending verification email: %w", err)
	}

	return nil
}

func (r *ResendClient) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", r.frontendURL, token)

	params := &resend.SendEmailRequest{
		From:    r.fromAddress,
		To:      []string{to},
		Subject: "Reset your password",
		Html: fmt.Sprintf(
			`<p>Click the link below to reset your password. This link expires in 30 minutes. If you didn't request this, you can ignore this email.</p><p><a href="%s">%s</a></p>`,
			link, link,
		),
	}

	if _, err := r.client.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("error in sending password reset email: %w", err)
	}

	return nil
}
