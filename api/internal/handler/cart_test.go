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

func sampleAddCartItemRow(id pgtype.UUID, quantity int32, stockOK bool) repository.AddCartItemRow {
	row := repository.AddCartItemRow{
		ProductID: id,
		Name:      "Kaos Polos",
		Price:     service.Float64ToNumeric(19.99),
		ImageUrl:  pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
	}
	if stockOK {
		row.Quantity = pgtype.Int4{Int32: quantity, Valid: true}
	}
	return row
}

func sampleUpdateCartItemQuantityRow(id pgtype.UUID, quantity int32, itemExists, stockOK bool) repository.UpdateCartItemQuantityRow {
	row := repository.UpdateCartItemQuantityRow{
		ProductID:  id,
		Name:       "Kaos Polos",
		Price:      service.Float64ToNumeric(19.99),
		ImageUrl:   pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
		ItemExists: itemExists,
	}
	if itemExists && stockOK {
		row.Quantity = pgtype.Int4{Int32: quantity, Valid: true}
	}
	return row
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

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			Return(sampleAddCartItemRow(pgProductID, 2, true), nil)

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

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).Return(repository.AddCartItemRow{}, pgx.ErrNoRows)

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

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).Return(sampleAddCartItemRow(pgProductID, 0, false), nil)

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

		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(sampleUpdateCartItemQuantityRow(pgProductID, 3, true, true), nil)

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

		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(sampleUpdateCartItemQuantityRow(pgProductID, 0, false, false), nil)

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

		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(sampleUpdateCartItemQuantityRow(pgProductID, 0, true, false), nil)

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
