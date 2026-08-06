package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMidtransClient(t *testing.T, handler http.HandlerFunc) *MidtransClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &MidtransClient{
		serverKey:  "test-server-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
}

func TestMidtransClient_CreateSnapTransaction(t *testing.T) {
	t.Run("sends order id, gross amount and basic auth, returns the token", func(t *testing.T) {
		client := newTestMidtransClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/transactions", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-server-key:"))
			assert.Equal(t, wantAuth, r.Header.Get("Authorization"))

			var body snapTransactionRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "order-123", body.TransactionDetails.OrderID)
			assert.Equal(t, int64(150000), body.TransactionDetails.GrossAmount)

			json.NewEncoder(w).Encode(snapTransactionResponse{Token: "snap-token-abc", RedirectURL: "https://example.com/pay"})
		})

		token, err := client.CreateSnapTransaction(context.Background(), "order-123", 150000)

		require.NoError(t, err)
		assert.Equal(t, "snap-token-abc", token)
	})

	t.Run("non-2xx status is surfaced as an error", func(t *testing.T) {
		client := newTestMidtransClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error_messages":["gross_amount is not equal"]}`))
		})

		token, err := client.CreateSnapTransaction(context.Background(), "order-123", 150000)

		assert.Error(t, err)
		assert.Empty(t, token)
	})
}

func TestBaseURLForEnv(t *testing.T) {
	assert.Equal(t, sandboxBaseURL, baseURLForEnv("sandbox"))
	assert.Equal(t, sandboxBaseURL, baseURLForEnv(""))
	assert.Equal(t, productionBaseURL, baseURLForEnv("production"))
}
