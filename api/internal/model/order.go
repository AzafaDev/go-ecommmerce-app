package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Price       string    `json:"price"`
	Quantity    int       `json:"quantity"`
	Subtotal    string    `json:"subtotal"`
}

// MidtransNotificationRequest declares only the fields we use; Midtrans sends more.
type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

type OrderResponse struct {
	ID          uuid.UUID           `json:"id"`
	Status      string              `json:"status"`
	TotalAmount string              `json:"total_amount"`
	SnapToken   string              `json:"snap_token,omitempty"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   time.Time           `json:"created_at"`
	PaidAt      *time.Time          `json:"paid_at,omitempty"`
}
