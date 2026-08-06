package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CartService struct {
	repo repository.Querier
}

func NewCartService(repo repository.Querier) *CartService {
	return &CartService{
		repo: repo,
	}
}

func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, req model.AddCartItemRequest) (*model.CartItemResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	pgProductID := pgtype.UUID{Bytes: req.ProductID, Valid: true}

	row, err := s.repo.AddCartItem(ctx, repository.AddCartItemParams{
		UserID:    pgUserID,
		ProductID: pgProductID,
		Quantity:  int32(req.Quantity),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in adding cart item: %w", err)
	}
	if !row.Quantity.Valid {
		return nil, ErrInsufficientStock
	}

	return toCartItemResponse(req.ProductID, row.Name, row.Price, row.ImageUrl, row.Quantity.Int32), nil
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItemResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

	row, err := s.repo.UpdateCartItemQuantity(ctx, repository.UpdateCartItemQuantityParams{
		Quantity:  int32(quantity),
		UserID:    pgUserID,
		ProductID: pgProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in updating cart item quantity: %w", err)
	}
	if !row.ItemExists {
		return nil, ErrCartItemNotFound
	}
	if !row.Quantity.Valid {
		return nil, ErrInsufficientStock
	}

	return toCartItemResponse(productID, row.Name, row.Price, row.ImageUrl, row.Quantity.Int32), nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, productID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteCartItem(ctx, repository.DeleteCartItemParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ProductID: pgtype.UUID{Bytes: productID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error in deleting cart item: %w", err)
	}
	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.ClearCart(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil {
		return fmt.Errorf("error in clearing cart: %w", err)
	}
	return nil
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*model.CartResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}

	items, err := s.repo.ListCartItems(ctx, pgUserID)
	if err != nil {
		return nil, fmt.Errorf("error in listing cart items: %w", err)
	}

	itemResponses := []model.CartItemResponse{}
	totalItems := 0
	total := pgtype.Numeric{Int: big.NewInt(0), Valid: true}
	for _, item := range items {
		itemResponses = append(itemResponses, model.CartItemResponse{
			ProductID: item.ProductID.Bytes,
			Name:      item.Name,
			Price:     NumericToString(item.Price),
			ImageURL:  item.ImageUrl.String,
			IsActive:  item.IsActive,
			Quantity:  int(item.Quantity),
			Subtotal:  NumericToString(item.Subtotal),
		})
		totalItems += int(item.Quantity)
		total = addNumeric(total, item.Subtotal)
	}

	return &model.CartResponse{
		Items:      itemResponses,
		TotalItems: totalItems,
		Total:      NumericToString(total),
	}, nil
}

// IsActive is hardcoded true: both source queries already filter on is_active = true.
func toCartItemResponse(productID uuid.UUID, name string, price pgtype.Numeric, imageURL pgtype.Text, quantity int32) *model.CartItemResponse {
	return &model.CartItemResponse{
		ProductID: productID,
		Name:      name,
		Price:     NumericToString(price),
		ImageURL:  imageURL.String,
		IsActive:  true,
		Quantity:  int(quantity),
		Subtotal:  NumericToString(multiplyNumericByInt(price, quantity)),
	}
}
