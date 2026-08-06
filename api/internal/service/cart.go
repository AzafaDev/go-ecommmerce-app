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

	product, err := s.repo.GetProductByID(ctx, pgProductID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in getting product by id: %w", err)
	}

	// Stock is validated inside the AddCartItem query itself (against the
	// existing row under lock), so no separate lookup+check is needed here.
	cartItem, err := s.repo.AddCartItem(ctx, repository.AddCartItemParams{
		UserID:    pgUserID,
		ProductID: pgProductID,
		Quantity:  int32(req.Quantity),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInsufficientStock
		}
		return nil, fmt.Errorf("error in adding cart item: %w", err)
	}

	return toCartItemResponse(req.ProductID, product, cartItem.Quantity), nil
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItemResponse, error) {
	pgUserID := pgtype.UUID{Bytes: userID, Valid: true}
	pgProductID := pgtype.UUID{Bytes: productID, Valid: true}

	product, err := s.repo.GetProductByID(ctx, pgProductID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in getting product by id: %w", err)
	}

	if int32(quantity) > product.Stock {
		return nil, ErrInsufficientStock
	}

	cartItem, err := s.repo.UpdateCartItemQuantity(ctx, repository.UpdateCartItemQuantityParams{
		Quantity:  int32(quantity),
		UserID:    pgUserID,
		ProductID: pgProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("error in updating cart item quantity: %w", err)
	}

	return toCartItemResponse(productID, product, cartItem.Quantity), nil
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

// multiplyNumericByInt multiplies a NUMERIC by an integer factor without
// going through float64, preserving exact decimal precision for money math.
func multiplyNumericByInt(n pgtype.Numeric, factor int32) pgtype.Numeric {
	if !n.Valid || n.Int == nil {
		return pgtype.Numeric{Int: big.NewInt(0), Valid: true}
	}
	return pgtype.Numeric{
		Int:   new(big.Int).Mul(n.Int, big.NewInt(int64(factor))),
		Exp:   n.Exp,
		Valid: true,
	}
}

// addNumeric adds two NUMERIC values without going through float64,
// aligning them to their common (more precise) exponent first since Numeric
// stores value as Int * 10^Exp and different rows may carry different Exp.
func addNumeric(a, b pgtype.Numeric) pgtype.Numeric {
	if !a.Valid || a.Int == nil {
		a = pgtype.Numeric{Int: big.NewInt(0), Exp: b.Exp, Valid: true}
	}
	if !b.Valid || b.Int == nil {
		return a
	}

	exp := a.Exp
	if b.Exp < exp {
		exp = b.Exp
	}

	scale := func(n *big.Int, from int32) *big.Int {
		if from == exp {
			return n
		}
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(from-exp)), nil)
		return new(big.Int).Mul(n, factor)
	}

	return pgtype.Numeric{
		Int:   new(big.Int).Add(scale(a.Int, a.Exp), scale(b.Int, b.Exp)),
		Exp:   exp,
		Valid: true,
	}
}

func toCartItemResponse(productID uuid.UUID, product repository.Product, quantity int32) *model.CartItemResponse {
	return &model.CartItemResponse{
		ProductID: productID,
		Name:      product.Name,
		Price:     NumericToString(product.Price),
		ImageURL:  product.ImageUrl.String,
		IsActive:  product.IsActive,
		Quantity:  int(quantity),
		Subtotal:  NumericToString(multiplyNumericByInt(product.Price, quantity)),
	}
}
