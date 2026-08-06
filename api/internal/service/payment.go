package service

import "context"

//go:generate go tool mockgen -source=payment.go -destination=mocks/payment/payment_mock.go -package=paymentmocks
type PaymentClient interface {
	CreateSnapTransaction(ctx context.Context, orderID string, grossAmount int64) (string, error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}
