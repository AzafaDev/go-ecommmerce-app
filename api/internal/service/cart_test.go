package service_test

import (
	"context"
	"errors"
	"testing"

	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"
	repomocks "go-ecommerce-app/internal/repository/mocks"
	"go-ecommerce-app/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestCartService(t *testing.T) (*service.CartService, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	svc := service.NewCartService(mockRepo)
	return svc, mockRepo
}

// Quantity is only valid when the upsert's stock guard let the write through.
func sampleAddCartItemRow(productID pgtype.UUID, quantity int32, stockOK bool) repository.AddCartItemRow {
	row := repository.AddCartItemRow{
		ProductID: productID,
		Name:      "Kaos Polos",
		Price:     service.Float64ToNumeric(19.99),
		ImageUrl:  pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
	}
	if stockOK {
		row.Quantity = pgtype.Int4{Int32: quantity, Valid: true}
	}
	return row
}

// Quantity is only valid when itemExists and the guarded UPDATE applied.
func sampleUpdateCartItemQuantityRow(productID pgtype.UUID, quantity int32, itemExists, stockOK bool) repository.UpdateCartItemQuantityRow {
	row := repository.UpdateCartItemQuantityRow{
		ProductID:  productID,
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

func TestCartService_AddItem(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	pgProductID := pgtype.UUID{Bytes: productID, Valid: true}
	req := model.AddCartItemRequest{ProductID: productID, Quantity: 2}

	t.Run("active product with enough stock succeeds", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			Return(sampleAddCartItemRow(pgProductID, 2, true), nil)

		item, err := svc.AddItem(context.Background(), userID, req)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, 2, item.Quantity)
		assert.Equal(t, "19.99", item.Price)
	})

	t.Run("product already in cart accumulates quantity instead of replacing it", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.AddCartItemParams) (repository.AddCartItemRow, error) {
				// The upsert itself is responsible for the actual accumulation in SQL;
				// here we only verify the requested quantity is forwarded as-is (2),
				// not pre-summed with the existing 3 in Go.
				assert.Equal(t, int32(2), arg.Quantity)
				return sampleAddCartItemRow(arg.ProductID, 5, true), nil
			})

		item, err := svc.AddItem(context.Background(), userID, req)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, 5, item.Quantity)
	})

	t.Run("product not found or inactive returns ErrProductNotFound", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		// The merged query's target_product CTE finds nothing, so the whole
		// statement returns zero rows.
		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).Return(repository.AddCartItemRow{}, pgx.ErrNoRows)

		item, err := svc.AddItem(context.Background(), userID, req)

		assert.ErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, item)
	})

	t.Run("insufficient stock reported by the guarded upsert returns ErrInsufficientStock", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		// Product exists (a row comes back), but the upsert's WHERE guard
		// rejected the write, so Quantity comes back NULL via the LEFT JOIN.
		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			Return(sampleAddCartItemRow(pgProductID, 0, false), nil)

		item, err := svc.AddItem(context.Background(), userID, req)

		assert.ErrorIs(t, err, service.ErrInsufficientStock)
		assert.Nil(t, item)
	})

	t.Run("unexpected error from AddCartItem is propagated", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)

		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any()).
			Return(repository.AddCartItemRow{}, errors.New("connection reset"))

		item, err := svc.AddItem(context.Background(), userID, req)

		assert.Error(t, err)
		assert.Nil(t, item)
	})
}

func TestCartService_UpdateItemQuantity(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

	t.Run("item not in cart returns ErrCartItemNotFound", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		// Product exists and is active, but there's no cart_items row for
		// this user/product yet.
		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(sampleUpdateCartItemQuantityRow(pgProductID, 0, false, false), nil)

		item, err := svc.UpdateItemQuantity(context.Background(), userID, productID, 2)

		assert.ErrorIs(t, err, service.ErrCartItemNotFound)
		assert.Nil(t, item)
	})

	t.Run("new quantity exceeding stock returns ErrInsufficientStock", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		// The cart item exists, but the guarded UPDATE's WHERE rejected the
		// write, so Quantity comes back NULL via the LEFT JOIN.
		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(sampleUpdateCartItemQuantityRow(pgProductID, 0, true, false), nil)

		item, err := svc.UpdateItemQuantity(context.Background(), userID, productID, 5)

		assert.ErrorIs(t, err, service.ErrInsufficientStock)
		assert.Nil(t, item)
	})

	t.Run("product not found returns ErrProductNotFound", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			Return(repository.UpdateCartItemQuantityRow{}, pgx.ErrNoRows)

		item, err := svc.UpdateItemQuantity(context.Background(), userID, productID, 2)

		assert.ErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, item)
	})

	t.Run("success sets quantity absolutely", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)

		mockRepo.EXPECT().UpdateCartItemQuantity(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.UpdateCartItemQuantityParams) (repository.UpdateCartItemQuantityRow, error) {
				assert.Equal(t, int32(7), arg.Quantity)
				return sampleUpdateCartItemQuantityRow(arg.ProductID, arg.Quantity, true, true), nil
			})

		item, err := svc.UpdateItemQuantity(context.Background(), userID, productID, 7)

		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, 7, item.Quantity)
	})
}

func TestCartService_RemoveItem(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()

	t.Run("item not in cart returns ErrCartItemNotFound via zero rows affected", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().DeleteCartItem(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		err := svc.RemoveItem(context.Background(), userID, productID)

		assert.ErrorIs(t, err, service.ErrCartItemNotFound)
	})

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().DeleteCartItem(gomock.Any(), gomock.Any()).Return(int64(1), nil)

		err := svc.RemoveItem(context.Background(), userID, productID)

		require.NoError(t, err)
	})

	t.Run("unexpected db error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().DeleteCartItem(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("connection reset"))

		err := svc.RemoveItem(context.Background(), userID, productID)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, service.ErrCartItemNotFound)
	})
}

func TestCartService_ClearCart(t *testing.T) {
	userID := uuid.New()

	t.Run("success, no special case for an already-empty cart", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().ClearCart(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.ClearCart(context.Background(), userID)

		require.NoError(t, err)
	})
}

func TestCartService_GetCart(t *testing.T) {
	userID := uuid.New()

	t.Run("empty cart returns an empty slice, not nil", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return(nil, nil)

		cart, err := svc.GetCart(context.Background(), userID)

		require.NoError(t, err)
		require.NotNil(t, cart)
		assert.NotNil(t, cart.Items)
		assert.Empty(t, cart.Items)
		assert.Equal(t, 0, cart.TotalItems)
		assert.Equal(t, "0", cart.Total)
	})

	t.Run("non-empty cart maps items and total", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		productID := uuid.New()

		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return([]repository.ListCartItemsRow{
			{
				ProductID: pgtype.UUID{Bytes: productID, Valid: true},
				Quantity:  2,
				Name:      "Kaos Polos",
				Price:     service.Float64ToNumeric(19.99),
				ImageUrl:  pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
				IsActive:  true,
				Stock:     10,
				Subtotal:  service.Float64ToNumeric(39.98),
			},
		}, nil)

		cart, err := svc.GetCart(context.Background(), userID)

		require.NoError(t, err)
		require.Len(t, cart.Items, 1)
		assert.Equal(t, "19.99", cart.Items[0].Price)
		assert.Equal(t, "39.98", cart.Items[0].Subtotal)
		assert.Equal(t, 2, cart.TotalItems)
		assert.Equal(t, "39.98", cart.Total)
	})

	t.Run("list error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))

		_, err := svc.GetCart(context.Background(), userID)

		assert.Error(t, err)
	})

	t.Run("total sums across items with differing numeric scale", func(t *testing.T) {
		svc, mockRepo := newTestCartService(t)
		productA, productB := uuid.New(), uuid.New()

		subtotalA, err := service.StringToNumeric("39.98") // Exp -2
		require.NoError(t, err)
		subtotalB, err := service.StringToNumeric("5.5") // Exp -1
		require.NoError(t, err)

		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return([]repository.ListCartItemsRow{
			{ProductID: pgtype.UUID{Bytes: productA, Valid: true}, Quantity: 2, Subtotal: subtotalA},
			{ProductID: pgtype.UUID{Bytes: productB, Valid: true}, Quantity: 1, Subtotal: subtotalB},
		}, nil)

		cart, err := svc.GetCart(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, 3, cart.TotalItems)
		assert.Equal(t, "45.48", cart.Total)
	})
}
