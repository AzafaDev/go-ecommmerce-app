package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusCancelled      = "cancelled"
	OrderStatusExpired        = "expired"
)

type OrderService struct {
	repo    repository.Querier
	dbPool  *pgxpool.Pool
	payment PaymentClient
}

func NewOrderService(repo repository.Querier, dbPool *pgxpool.Pool, payment PaymentClient) *OrderService {
	return &OrderService{
		repo:    repo,
		dbPool:  dbPool,
		payment: payment,
	}
}

// Checkout reserves stock immediately rather than on payment confirmation, so an unpaid order holds items until the sweep job expires it.
func (s *OrderService) Checkout(ctx context.Context, userID uuid.UUID) (*model.OrderResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}

	items, err := s.repo.ListCartItems(ctx, pgUserID)
	if err != nil {
		return nil, fmt.Errorf("error in listing cart items: %w", err)
	}
	if len(items) == 0 {
		return nil, ErrCartEmpty
	}

	total := pgtype.Numeric{Int: big.NewInt(0), Valid: true}
	for _, item := range items {
		if item.Quantity > item.Stock {
			return nil, ErrInsufficientStock
		}
		total = addNumeric(total, item.Subtotal)
	}

	var order repository.Order
	orderItems := make([]repository.OrderItem, 0, len(items))

	err = repository.WithTx(ctx, s.dbPool, func(q *repository.Queries) error {
		createdOrder, err := q.CreateOrder(ctx, repository.CreateOrderParams{
			UserID:      pgUserID,
			Status:      OrderStatusPendingPayment,
			TotalAmount: total,
		})
		if err != nil {
			return fmt.Errorf("error in creating order: %w", err)
		}
		order = createdOrder

		for _, item := range items {
			createdItem, err := q.CreateOrderItem(ctx, repository.CreateOrderItemParams{
				OrderID:     order.ID,
				ProductID:   item.ProductID,
				ProductName: item.Name,
				Price:       item.Price,
				Quantity:    item.Quantity,
				Subtotal:    item.Subtotal,
			})
			if err != nil {
				return fmt.Errorf("error in creating order item: %w", err)
			}

			// Re-checked under the row lock to catch a concurrent checkout racing for the same stock.
			if _, err := q.DecrementProductStock(ctx, repository.DecrementProductStockParams{
				Stock: item.Quantity,
				ID:    item.ProductID,
			}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrInsufficientStock
				}
				return fmt.Errorf("error in decrementing product stock: %w", err)
			}

			orderItems = append(orderItems, createdItem)
		}

		if err := q.ClearCart(ctx, pgUserID); err != nil {
			return fmt.Errorf("error in clearing cart: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Runs after commit since it's an external call; a failure here leaves the order pending_payment for the sweep job to expire later.
	grossAmount, err := NumericToInt64(order.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("error in converting order total for midtrans: %w", err)
	}

	midtransOrderID := uuid.UUID(order.ID.Bytes).String()
	snapToken, err := s.payment.CreateSnapTransaction(ctx, midtransOrderID, grossAmount)
	if err != nil {
		return nil, fmt.Errorf("error in creating midtrans transaction: %w", err)
	}

	order, err = s.repo.SetOrderPaymentInfo(ctx, repository.SetOrderPaymentInfoParams{
		MidtransOrderID: pgtype.Text{String: midtransOrderID, Valid: true},
		SnapToken:       pgtype.Text{String: snapToken, Valid: true},
		ID:              order.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("error in saving midtrans payment info: %w", err)
	}

	return toOrderResponse(order, orderItems), nil
}

// HandleWebhook is idempotent: orders already in a final status are acknowledged as a no-op, since Midtrans retries non-2xx deliveries.
func (s *OrderService) HandleWebhook(ctx context.Context, notif model.MidtransNotificationRequest) error {
	if !s.payment.VerifySignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey) {
		return ErrInvalidSignature
	}

	order, err := s.repo.GetOrderByMidtransID(ctx, pgtype.Text{String: notif.OrderID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("error in getting order by midtrans id: %w", err)
	}

	if isFinalOrderStatus(order.Status) {
		return nil
	}

	newStatus, shouldRestock, handled := mapMidtransTransactionStatus(notif.TransactionStatus, notif.FraudStatus)
	if !handled {
		return nil
	}

	if !shouldRestock {
		if _, err := s.repo.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
			Status: newStatus,
			PaidAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ID:     order.ID,
		}); err != nil {
			return fmt.Errorf("error in updating order status: %w", err)
		}
		return nil
	}

	return s.setStatusAndRestock(ctx, order.ID, newStatus)
}

// SweepExpiredOrders expires stale pending_payment orders and returns their stock, continuing past individual failures.
func (s *OrderService) SweepExpiredOrders(ctx context.Context, threshold time.Duration) (int, error) {
	cutoff := time.Now().Add(-threshold)

	orders, err := s.repo.ListExpiredPendingOrders(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("error in listing expired pending orders: %w", err)
	}

	var errs []error
	swept := 0
	for _, order := range orders {
		if err := s.setStatusAndRestock(ctx, order.ID, OrderStatusExpired); err != nil {
			errs = append(errs, fmt.Errorf("order %s: %w", uuid.UUID(order.ID.Bytes), err))
			continue
		}
		swept++
	}

	return swept, errors.Join(errs...)
}

// setStatusAndRestock sets a non-paid final status and returns stock in one transaction.
func (s *OrderService) setStatusAndRestock(ctx context.Context, orderID pgtype.UUID, status string) error {
	return repository.WithTx(ctx, s.dbPool, func(q *repository.Queries) error {
		if _, err := q.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
			Status: status,
			ID:     orderID,
		}); err != nil {
			return fmt.Errorf("error in updating order status: %w", err)
		}

		items, err := q.ListOrderItemsByOrder(ctx, orderID)
		if err != nil {
			return fmt.Errorf("error in listing order items: %w", err)
		}

		for _, item := range items {
			if err := q.IncrementProductStock(ctx, repository.IncrementProductStockParams{
				Stock: item.Quantity,
				ID:    item.ProductID,
			}); err != nil {
				return fmt.Errorf("error in restocking product: %w", err)
			}
		}

		return nil
	})
}

func (s *OrderService) ListOrders(ctx context.Context, userID uuid.UUID) ([]model.OrderResponse, error) {
	listOrdersResp := []model.OrderResponse{}
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	orders, err := s.repo.ListOrdersByUser(ctx, pgUserID)
	if err != nil {
		return nil, fmt.Errorf("error in listing orders by user: %w", err)
	}
	for _, order := range orders {
		orderItems, err := s.repo.ListOrderItemsByOrder(ctx, order.ID)
		if err != nil {
			return nil, fmt.Errorf("error in listing orders by order: %w", err)
		}
		listOrdersResp = append(listOrdersResp, *toOrderResponse(order, orderItems))
	}
	return listOrdersResp, nil
}

func (s *OrderService) GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*model.OrderResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	pgOrderID := pgtype.UUID{Bytes: orderID, Valid: true}

	existingOrder, err := s.repo.GetOrderByID(ctx, pgOrderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("error in getting order by id: %w", err)
	}
	if existingOrder.UserID != pgUserID {
		return nil, ErrOrderNotFound
	}

	listOrderItems, err := s.repo.ListOrderItemsByOrder(ctx, existingOrder.ID)
	if err != nil {
		return nil, fmt.Errorf("error in listing order items by order: %w", err)
	}

	return toOrderResponse(existingOrder, listOrderItems), nil
}

func isFinalOrderStatus(status string) bool {
	switch status {
	case OrderStatusPaid, OrderStatusCancelled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// handled is false for statuses like "pending" that don't map to a final order state.
func mapMidtransTransactionStatus(transactionStatus, fraudStatus string) (status string, shouldRestock, handled bool) {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "deny" {
			return OrderStatusCancelled, true, true
		}
		return OrderStatusPaid, false, true
	case "settlement":
		return OrderStatusPaid, false, true
	case "deny", "cancel":
		return OrderStatusCancelled, true, true
	case "expire":
		return OrderStatusExpired, true, true
	default:
		return "", false, false
	}
}

func toOrderResponse(order repository.Order, items []repository.OrderItem) *model.OrderResponse {
	itemResponses := make([]model.OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, model.OrderItemResponse{
			ProductID:   item.ProductID.Bytes,
			ProductName: item.ProductName,
			Price:       NumericToString(item.Price),
			Quantity:    int(item.Quantity),
			Subtotal:    NumericToString(item.Subtotal),
		})
	}

	var paidAt *time.Time
	if order.PaidAt.Valid {
		paidAt = &order.PaidAt.Time
	}

	return &model.OrderResponse{
		ID:          order.ID.Bytes,
		Status:      order.Status,
		TotalAmount: NumericToString(order.TotalAmount),
		SnapToken:   order.SnapToken.String,
		Items:       itemResponses,
		CreatedAt:   order.CreatedAt.Time,
		PaidAt:      paidAt,
	}
}
