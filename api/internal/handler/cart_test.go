package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-ecommerce-app/internal/handler"
	"go-ecommerce-app/internal/model"
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

const cartTestSecretKey = "test-jwt-secret"

// newTestCartRouter mounts CartRoutes on a real chi.Mux so path params
// (chi.URLParam) and the auth middleware behave like production.
func newTestCartRouter(t *testing.T) (*chi.Mux, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	svc := service.NewCartService(mockRepo)
	h := handler.NewCartHandler(svc, cartTestSecretKey)

	r := chi.NewRouter()
	h.CartRoutes(r)
	return r, mockRepo
}

func cartUserToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	token, err := security.GenerateToken(cartTestSecretKey, time.Hour, userID, "customer")
	require.NoError(t, err)
	return token
}

func sampleCartRepoProduct(id pgtype.UUID) repository.Product {
	return repository.Product{
		ID:       id,
		Name:     "Kaos Polos",
		Price:    service.Float64ToNumeric(19.99),
		Stock:    10,
		Sku:      "SKU-001",
		Category: "apparel",
		ImageUrl: pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
		IsActive: true,
	}
}

func TestCartHandler_GetCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()

		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return([]repository.ListCartItemsRow{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/cart/", nil)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["cart"])
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestCartRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/cart/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestCartHandler_AddItem(t *testing.T) {
	productID := uuid.New()
	payload := model.AddCartItemRequest{ProductID: productID, Quantity: 2}

	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(sampleCartRepoProduct(pgProductID), nil)
		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			Return(repository.CartItem{ProductID: pgProductID, Quantity: 2}, nil)

		req := newJSONRequest(t, http.MethodPost, "/cart/items/", payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["item"])
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestCartRouter(t)

		req := newJSONRequest(t, http.MethodPost, "/cart/items/", payload)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		router, _ := newTestCartRouter(t)
		userID := uuid.New()

		req := newJSONRequest(t, http.MethodPost, "/cart/items/", model.AddCartItemRequest{})
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("product not found returns 404", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(repository.Product{}, pgx.ErrNoRows)

		req := newJSONRequest(t, http.MethodPost, "/cart/items/", payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("insufficient stock returns 409", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		pgProductID := pgtype.UUID{Bytes: productID, Valid: true}
		lowStock := sampleCartRepoProduct(pgProductID)
		lowStock.Stock = 1

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(lowStock, nil)
		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).Return(repository.CartItem{}, pgx.ErrNoRows)

		req := newJSONRequest(t, http.MethodPost, "/cart/items/", payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
	})
}

func TestCartHandler_UpdateItemQuantity(t *testing.T) {
	payload := model.UpdateCartItemRequest{Quantity: 3}

	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()
		pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(sampleCartRepoProduct(pgProductID), nil)
		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(repository.CartItem{ProductID: pgProductID, Quantity: 3}, nil)

		req := newJSONRequest(t, http.MethodPatch, "/cart/items/"+productID.String(), payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid uuid returns 400", func(t *testing.T) {
		router, _ := newTestCartRouter(t)
		userID := uuid.New()

		req := newJSONRequest(t, http.MethodPatch, "/cart/items/not-a-uuid", payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("item not in cart returns 404", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()
		pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(sampleCartRepoProduct(pgProductID), nil)
		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(repository.CartItem{}, pgx.ErrNoRows)

		req := newJSONRequest(t, http.MethodPatch, "/cart/items/"+productID.String(), payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("quantity exceeding stock returns 409", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()
		pgProductID := pgtype.UUID{Bytes: productID, Valid: true}
		lowStock := sampleCartRepoProduct(pgProductID)
		lowStock.Stock = 1

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(lowStock, nil)

		req := newJSONRequest(t, http.MethodPatch, "/cart/items/"+productID.String(), payload)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("quantity=0 is rejected by validation, not treated as delete", func(t *testing.T) {
		router, _ := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()

		req := newJSONRequest(t, http.MethodPatch, "/cart/items/"+productID.String(), model.UpdateCartItemRequest{Quantity: 0})
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestCartHandler_RemoveItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()

		mockRepo.EXPECT().DeleteCartItem(gomock.Any(), gomock.Any()).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+productID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("item not found returns 404", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()
		productID := uuid.New()

		mockRepo.EXPECT().DeleteCartItem(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+productID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid uuid returns 400", func(t *testing.T) {
		router, _ := newTestCartRouter(t)
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/cart/items/not-a-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestCartHandler_ClearCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestCartRouter(t)
		userID := uuid.New()

		mockRepo.EXPECT().ClearCart(gomock.Any(), gomock.Any()).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart/", nil)
		req.Header.Set("Authorization", "Bearer "+cartUserToken(t, userID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestCartRouter(t)

		req := httptest.NewRequest(http.MethodDelete, "/cart/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
