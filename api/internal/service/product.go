package service

import (
	"context"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProductService struct {
	repo repository.Querier
}

func NewProductService(repo repository.Querier) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.ProductResponse, error) {
	var pgErr *pgconn.PgError

	createdProduct, err := s.repo.CreateProduct(ctx, repository.CreateProductParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price: Float64ToNumeric(req.Price),
		Stock: int32(req.Stock),
		Sku:   req.SKU,
		Category: pgtype.Text{
			String: req.Category,
			Valid:  true,
		},
	})

	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, ErrSKUTaken
		}
		return nil, fmt.Errorf("error in creating product")
	}

	return &model.ProductResponse{
		ID:          createdProduct.ID.Bytes,
		Name:        createdProduct.Name,
		Description: createdProduct.Description.String,
		Price:       NumericToString(createdProduct.Price),
		Stock:       int(createdProduct.Stock),
		SKU:         createdProduct.Sku,
		Category:    createdProduct.Category.String,
		IsActive:    createdProduct.IsActive,
		CreatedAt:   createdProduct.CreatedAt.Time,
		UpdatedAt:   createdProduct.UpdatedAt.Time,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req model.UpdateProductRequest, id uuid.UUID) (*model.ProductResponse, error) {
	var pgErr *pgconn.PgError

	updatedProduct, err := s.repo.UpdateProduct(ctx, repository.UpdateProductParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price: Float64ToNumeric(req.Price),
		Stock: int32(req.Stock),
		Sku:   req.SKU,
		Category: pgtype.Text{
			String: req.Category,
			Valid:  true,
		},
		IsActive: req.IsActive,
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	})

	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, ErrSKUTaken
		} else if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in updating product")
	}

	return &model.ProductResponse{
		ID:          updatedProduct.ID.Bytes,
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description.String,
		Price:       NumericToString(updatedProduct.Price),
		Stock:       int(updatedProduct.Stock),
		SKU:         updatedProduct.Sku,
		Category:    updatedProduct.Category.String,
		IsActive:    updatedProduct.IsActive,
		CreatedAt:   updatedProduct.CreatedAt.Time,
		UpdatedAt:   updatedProduct.UpdatedAt.Time,
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.SoftDeleteProduct(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}
		return fmt.Errorf("error in deleting product")
	}

	return nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error) {
	existingProduct, err := s.repo.GetProductByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in getting product by id: %w", err)
	}

	return &model.ProductResponse{
		ID:          existingProduct.ID.Bytes,
		Name:        existingProduct.Name,
		Description: existingProduct.Description.String,
		Price:       NumericToString(existingProduct.Price),
		Stock:       int(existingProduct.Stock),
		SKU:         existingProduct.Sku,
		Category:    existingProduct.Category.String,
		IsActive:    existingProduct.IsActive,
		CreatedAt:   existingProduct.CreatedAt.Time,
		UpdatedAt:   existingProduct.UpdatedAt.Time,
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, search, category string, page, limit int) (model.ListProductsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit >= 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	listProducts, err := s.repo.ListProducts(ctx, repository.ListProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		Search: pgtype.Text{
			String: search,
			Valid:  search != "",
		},
		Category: pgtype.Text{
			String: category,
			Valid:  category != "",
		},
	})

	if err != nil {
		return model.ListProductsResponse{}, fmt.Errorf("error in listing products: %w", err)
	}

	totalItems, err := s.repo.CountProducts(ctx, repository.CountProductsParams{
		Search: pgtype.Text{
			String: search,
			Valid:  search != "",
		},
		Category: pgtype.Text{
			String: category,
			Valid:  category != "",
		},
	})

	if err != nil {
		return model.ListProductsResponse{}, fmt.Errorf("error in counting products: %w", err)
	}

	listProductResponse := []model.ProductResponse{}

	for _, p := range listProducts {
		listProductResponse = append(listProductResponse, model.ProductResponse{
			ID:          p.ID.Bytes,
			Name:        p.Name,
			Description: p.Description.String,
			Price:       NumericToString(p.Price),
			Stock:       int(p.Stock),
			SKU:         p.Sku,
			Category:    p.Category.String,
			IsActive:    p.IsActive,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		})
	}

	return model.ListProductsResponse{
		Data: listProductResponse,
		Meta: model.MetaListProductResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: int(totalItems),
			TotalPages: (int(totalItems) + limit - 1) / limit,
		},
	}, nil
}
