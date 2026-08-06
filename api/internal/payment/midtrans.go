package payment

import (
	"bytes"
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	sandboxBaseURL    = "https://app.sandbox.midtrans.com/snap/v1"
	productionBaseURL = "https://app.midtrans.com/snap/v1"
)

// MidtransClient calls the Snap REST API directly; no official Go SDK exists.
type MidtransClient struct {
	serverKey  string
	baseURL    string
	httpClient *http.Client
}

func NewMidtransClient(serverKey, env string) *MidtransClient {
	return &MidtransClient{
		serverKey:  serverKey,
		baseURL:    baseURLForEnv(env),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func baseURLForEnv(env string) string {
	if env == "production" {
		return productionBaseURL
	}
	return sandboxBaseURL
}

type snapTransactionRequest struct {
	TransactionDetails snapTransactionDetails `json:"transaction_details"`
}

type snapTransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type snapTransactionResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func (c *MidtransClient) CreateSnapTransaction(ctx context.Context, orderID string, grossAmount int64) (string, error) {
	payload, err := json.Marshal(snapTransactionRequest{
		TransactionDetails: snapTransactionDetails{
			OrderID:     orderID,
			GrossAmount: grossAmount,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error in encoding midtrans request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/transactions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("error in building midtrans request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.serverKey, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in calling midtrans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("midtrans returned status %d: %s", resp.StatusCode, body)
	}

	var result snapTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error in decoding midtrans response: %w", err)
	}

	return result.Token, nil
}

// VerifySignature is the only auth on the webhook: SHA512(order_id+status_code+gross_amount+server_key).
func (c *MidtransClient) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	raw := orderID + statusCode + grossAmount + c.serverKey
	sum := sha512.Sum512([]byte(raw))
	computed := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(signatureKey)) == 1
}
