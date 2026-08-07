package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"
	repomocks "go-ecommerce-app/internal/repository/mocks"
	"go-ecommerce-app/internal/service"
	paymentmocks "go-ecommerce-app/internal/service/mocks/payment"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestOrderService(t *testing.T) (*service.OrderService, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	// dbPool/payment are nil: these cases short-circuit before the transaction opens.
	svc := service.NewOrderService(mockRepo, nil, nil)
	return svc, mockRepo
}

func newTestOrderServiceWithPayment(t *testing.T) (*service.OrderService, *repomocks.MockQuerier, *paymentmocks.MockPaymentClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	mockPayment := paymentmocks.NewMockPaymentClient(ctrl)
	// dbPool is nil: none of these webhook cases reach the restock transaction.
	svc := service.NewOrderService(mockRepo, nil, mockPayment)
	return svc, mockRepo, mockPayment
}

func sampleOrderCartRow(productID pgtype.UUID, quantity, stock int32) repository.ListCartItemsRow {
	return repository.ListCartItemsRow{
		ProductID: productID,
		Quantity:  quantity,
		Name:      "Kaos Polos",
		Price:     service.Float64ToNumeric(19.99),
		ImageUrl:  pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
		IsActive:  true,
		Stock:     stock,
		Subtotal:  service.Float64ToNumeric(19.99 * float64(quantity)),
	}
}

func TestOrderService_Checkout(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()

	t.Run("empty cart returns ErrCartEmpty", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return(nil, nil)

		order, err := svc.Checkout(context.Background(), userID)

		assert.ErrorIs(t, err, service.ErrCartEmpty)
		assert.Nil(t, order)
	})

	t.Run("cart item quantity exceeding live stock returns ErrInsufficientStock without opening a transaction", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		row := sampleOrderCartRow(pgtype.UUID{Bytes: productID, Valid: true}, 5, 3)
		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return([]repository.ListCartItemsRow{row}, nil)
		// CreateOrder deliberately NOT EXPECT()'d: must short-circuit before writing.

		order, err := svc.Checkout(context.Background(), userID)

		assert.ErrorIs(t, err, service.ErrInsufficientStock)
		assert.Nil(t, order)
	})

	t.Run("list cart items error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListCartItems(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))

		order, err := svc.Checkout(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, order)
	})
}

func sampleNotification(orderID string, status string) model.MidtransNotificationRequest {
	return model.MidtransNotificationRequest{
		OrderID:           orderID,
		StatusCode:        "200",
		GrossAmount:       "150000.00",
		SignatureKey:      "valid-signature",
		TransactionStatus: status,
	}
}

func TestOrderService_HandleWebhook(t *testing.T) {
	orderID := uuid.New()
	pgOrderID := pgtype.UUID{Bytes: orderID, Valid: true}

	t.Run("invalid signature is rejected before touching order data", func(t *testing.T) {
		svc, _, mockPayment := newTestOrderServiceWithPayment(t)
		notif := sampleNotification(orderID.String(), "settlement")

		mockPayment.EXPECT().VerifySignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey).Return(false)
		// GetOrderByMidtransID deliberately NOT EXPECT()'d: must short-circuit before reading.

		err := svc.HandleWebhook(context.Background(), notif)

		assert.ErrorIs(t, err, service.ErrInvalidSignature)
	})

	t.Run("unknown midtrans order id returns ErrOrderNotFound", func(t *testing.T) {
		svc, mockRepo, mockPayment := newTestOrderServiceWithPayment(t)
		notif := sampleNotification(orderID.String(), "settlement")

		mockPayment.EXPECT().VerifySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
		mockRepo.EXPECT().GetOrderByMidtransID(gomock.Any(), gomock.Any()).Return(repository.Order{}, pgx.ErrNoRows)

		err := svc.HandleWebhook(context.Background(), notif)

		assert.ErrorIs(t, err, service.ErrOrderNotFound)
	})

	t.Run("order already in a final status is a no-op (idempotent retry)", func(t *testing.T) {
		svc, mockRepo, mockPayment := newTestOrderServiceWithPayment(t)
		notif := sampleNotification(orderID.String(), "settlement")

		mockPayment.EXPECT().VerifySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
		mockRepo.EXPECT().GetOrderByMidtransID(gomock.Any(), gomock.Any()).
			Return(repository.Order{ID: pgOrderID, Status: service.OrderStatusPaid}, nil)
		// UpdateOrderStatus deliberately NOT EXPECT()'d: a final-status order must not be mutated again.

		err := svc.HandleWebhook(context.Background(), notif)

		assert.NoError(t, err)
	})

	t.Run("settlement marks the order paid without opening a restock transaction", func(t *testing.T) {
		svc, mockRepo, mockPayment := newTestOrderServiceWithPayment(t)
		notif := sampleNotification(orderID.String(), "settlement")

		mockPayment.EXPECT().VerifySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
		mockRepo.EXPECT().GetOrderByMidtransID(gomock.Any(), gomock.Any()).
			Return(repository.Order{ID: pgOrderID, Status: service.OrderStatusPendingPayment}, nil)
		mockRepo.EXPECT().UpdateOrderStatus(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error) {
				assert.Equal(t, service.OrderStatusPaid, arg.Status)
				assert.True(t, arg.PaidAt.Valid)
				return repository.Order{ID: arg.ID, Status: arg.Status, PaidAt: arg.PaidAt}, nil
			})

		err := svc.HandleWebhook(context.Background(), notif)

		assert.NoError(t, err)
	})

	t.Run("pending transaction status is acknowledged without mutating the order", func(t *testing.T) {
		svc, mockRepo, mockPayment := newTestOrderServiceWithPayment(t)
		notif := sampleNotification(orderID.String(), "pending")

		mockPayment.EXPECT().VerifySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
		mockRepo.EXPECT().GetOrderByMidtransID(gomock.Any(), gomock.Any()).
			Return(repository.Order{ID: pgOrderID, Status: service.OrderStatusPendingPayment}, nil)
		// UpdateOrderStatus deliberately NOT EXPECT()'d: "pending" isn't a final state we act on.

		err := svc.HandleWebhook(context.Background(), notif)

		assert.NoError(t, err)
	})
}

func sampleOrder(id, userID uuid.UUID, status string) repository.Order {
	return repository.Order{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Status:      status,
		TotalAmount: service.Float64ToNumeric(19.99),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func sampleOrderItemRow(orderID, productID uuid.UUID) repository.OrderItem {
	return repository.OrderItem{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrderID:     pgtype.UUID{Bytes: orderID, Valid: true},
		ProductID:   pgtype.UUID{Bytes: productID, Valid: true},
		ProductName: "Kaos Polos",
		Price:       service.Float64ToNumeric(19.99),
		Quantity:    1,
		Subtotal:    service.Float64ToNumeric(19.99),
	}
}

func TestOrderService_ListOrders(t *testing.T) {
	userID := uuid.New()

	t.Run("no orders returns an empty slice, not nil", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListOrdersByUser(gomock.Any(), gomock.Any()).Return(nil, nil)

		orders, err := svc.ListOrders(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, []model.OrderResponse{}, orders)
	})

	t.Run("fetches items per order and maps them into the response", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		orderA := sampleOrder(uuid.New(), userID, service.OrderStatusPaid)
		orderB := sampleOrder(uuid.New(), userID, service.OrderStatusPendingPayment)
		productID := uuid.New()

		mockRepo.EXPECT().ListOrdersByUser(gomock.Any(), gomock.Any()).
			Return([]repository.Order{orderA, orderB}, nil)
		mockRepo.EXPECT().ListOrderItemsByOrder(gomock.Any(), orderA.ID).
			Return([]repository.OrderItem{sampleOrderItemRow(uuid.UUID(orderA.ID.Bytes), productID)}, nil)
		mockRepo.EXPECT().ListOrderItemsByOrder(gomock.Any(), orderB.ID).
			Return([]repository.OrderItem{sampleOrderItemRow(uuid.UUID(orderB.ID.Bytes), productID)}, nil)

		orders, err := svc.ListOrders(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Len(t, orders[0].Items, 1)
		assert.Len(t, orders[1].Items, 1)
	})

	t.Run("list orders error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListOrdersByUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))

		orders, err := svc.ListOrders(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, orders)
	})
}

func TestOrderService_GetOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()

	t.Run("success returns the order with its items", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		order := sampleOrder(orderID, userID, service.OrderStatusPaid)
		productID := uuid.New()

		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(order, nil)
		mockRepo.EXPECT().ListOrderItemsByOrder(gomock.Any(), order.ID).
			Return([]repository.OrderItem{sampleOrderItemRow(orderID, productID)}, nil)

		got, err := svc.GetOrder(context.Background(), userID, orderID)

		assert.NoError(t, err)
		assert.Equal(t, orderID, got.ID)
		assert.Len(t, got.Items, 1)
	})

	t.Run("unknown order id returns ErrOrderNotFound", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(repository.Order{}, pgx.ErrNoRows)

		got, err := svc.GetOrder(context.Background(), userID, orderID)

		assert.ErrorIs(t, err, service.ErrOrderNotFound)
		assert.Nil(t, got)
	})

	t.Run("order belonging to another user returns ErrOrderNotFound", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		order := sampleOrder(orderID, uuid.New(), service.OrderStatusPaid)
		mockRepo.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Return(order, nil)
		// ListOrderItemsByOrder deliberately NOT EXPECT()'d: ownership check must short-circuit first.

		got, err := svc.GetOrder(context.Background(), userID, orderID)

		assert.ErrorIs(t, err, service.ErrOrderNotFound)
		assert.Nil(t, got)
	})
}

func TestOrderService_SweepExpiredOrders(t *testing.T) {
	t.Run("nothing past the threshold sweeps zero without opening a transaction", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListExpiredPendingOrders(gomock.Any(), gomock.Any()).Return(nil, nil)

		swept, err := svc.SweepExpiredOrders(context.Background(), 30*time.Minute)

		assert.NoError(t, err)
		assert.Equal(t, 0, swept)
	})

	t.Run("list error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestOrderService(t)
		mockRepo.EXPECT().ListExpiredPendingOrders(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))

		swept, err := svc.SweepExpiredOrders(context.Background(), 30*time.Minute)

		assert.Error(t, err)
		assert.Equal(t, 0, swept)
	})
}
