package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-ecommerce-app/internal/handler"
	"go-ecommerce-app/internal/repository"
	repomocks "go-ecommerce-app/internal/repository/mocks"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/security"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const orderTestSecretKey = "test-jwt-secret"

// newTestOrderRouter mounts OrderRoutes on a real chi.Mux so path params
// (chi.URLParam) and the auth middleware behave like production.
func newTestOrderRouter(t *testing.T) (*chi.Mux, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	svc := service.NewOrderService(mockRepo, nil, nil)
	h := handler.NewOrderHandler(svc, orderTestSecretKey)

	r := chi.NewRouter()
	h.OrderRoutes(r)
	return r, mockRepo
}

func orderUserToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	token, err := security.GenerateToken(orderTestSecretKey, time.Hour, userID, "customer")
	require.NoError(t, err)
	return token
}

func sampleOrderRow(id, userID uuid.UUID, status string) repository.Order {
	return repository.Order{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Status:      status,
		TotalAmount: service.Float64ToNumeric(19.99),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestOrderHandler_ListOrders(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestOrderRouter(t)
		userID := uuid.New()

		mockRepo.EXPECT().ListOrdersByUser(gomock.Any(), gomock.Any()).Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
		req.Header.Set("Authorization", "Bearer "+orderUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["orders"])
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestOrderRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestOrderHandler_GetOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestOrderRouter(t)
		userID := uuid.New()
		orderID := uuid.New()
		order := sampleOrderRow(orderID, userID, service.OrderStatusPaid)

		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(order, nil)
		mockRepo.EXPECT().ListOrderItemsByOrder(gomock.Any(), gomock.Any()).Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+orderUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["order"])
	})

	t.Run("not found for another user's order", func(t *testing.T) {
		router, mockRepo := newTestOrderRouter(t)
		userID := uuid.New()
		orderID := uuid.New()
		order := sampleOrderRow(orderID, uuid.New(), service.OrderStatusPaid)

		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(order, nil)
		// ListOrderItemsByOrder deliberately NOT EXPECT()'d: ownership check must short-circuit first.

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+orderUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("unknown order id returns 404", func(t *testing.T) {
		router, mockRepo := newTestOrderRouter(t)
		userID := uuid.New()
		orderID := uuid.New()

		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(repository.Order{}, pgx.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+orderUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid id format", func(t *testing.T) {
		router, _ := newTestOrderRouter(t)
		userID := uuid.New()
		// GetOrderByID deliberately NOT EXPECT()'d: uuid.Parse must fail before the service is called.

		req := httptest.NewRequest(http.MethodGet, "/orders/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+orderUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestOrderRouter(t)
		orderID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
