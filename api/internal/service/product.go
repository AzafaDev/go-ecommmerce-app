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

	numericPrice, err := StringToNumeric(req.Price)
	if err != nil {
		return nil, err
	}

	createdProduct, err := s.repo.CreateProduct(ctx, repository.CreateProductParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price:    numericPrice,
		Stock:    int32(req.Stock),
		Sku:      req.SKU,
		Category: req.Category,
		ImageUrl: pgtype.Text{
			String: req.ImageURL,
			Valid:  req.ImageURL != "",
		},
	})

	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, ErrSKUTaken
		}
		return nil, fmt.Errorf("error in creating product")
	}

	return toProductResponse(createdProduct), nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req model.UpdateProductRequest, id uuid.UUID) (*model.ProductResponse, error) {
	numericPrice, err := StringToNumeric(req.Price)
	if err != nil {
		return nil, err
	}

	updatedProduct, err := s.repo.UpdateProduct(ctx, repository.UpdateProductParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price:    numericPrice,
		Stock:    int32(req.Stock),
		Category: req.Category,
		IsActive: req.IsActive,
		ImageUrl: pgtype.Text{
			String: req.ImageURL,
			Valid:  req.ImageURL != "",
		},
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in updating product")
	}

	return toProductResponse(updatedProduct), nil
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

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID, includeInactive bool) (*model.ProductResponse, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	var existingProduct repository.Product
	var err error
	if includeInactive {
		existingProduct, err = s.repo.AdminGetProductByID(ctx, pgID)
	} else {
		existingProduct, err = s.repo.GetProductByID(ctx, pgID)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error in getting product by id: %w", err)
	}

	return toProductResponse(existingProduct), nil
}

func (s *ProductService) ListProducts(ctx context.Context, search, category string, page, limit int, includeInactive bool) (model.ListProductsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	searchParam := pgtype.Text{String: search, Valid: search != ""}
	categoryParam := pgtype.Text{String: category, Valid: category != ""}

	var listProducts []repository.Product
	var totalItems int64
	var err error

	if includeInactive {
		listProducts, err = s.repo.AdminListProducts(ctx, repository.AdminListProductsParams{
			Limit:    int32(limit),
			Offset:   int32(offset),
			Search:   searchParam,
			Category: categoryParam,
		})
	} else {
		listProducts, err = s.repo.ListProducts(ctx, repository.ListProductsParams{
			Limit:    int32(limit),
			Offset:   int32(offset),
			Search:   searchParam,
			Category: categoryParam,
		})
	}

	if err != nil {
		return model.ListProductsResponse{}, fmt.Errorf("error in listing products: %w", err)
	}

	if includeInactive {
		totalItems, err = s.repo.AdminCountProducts(ctx, repository.AdminCountProductsParams{
			Search:   searchParam,
			Category: categoryParam,
		})
	} else {
		totalItems, err = s.repo.CountProducts(ctx, repository.CountProductsParams{
			Search:   searchParam,
			Category: categoryParam,
		})
	}

	if err != nil {
		return model.ListProductsResponse{}, fmt.Errorf("error in counting products: %w", err)
	}

	listProductResponse := []model.ProductResponse{}
	for _, p := range listProducts {
		listProductResponse = append(listProductResponse, *toProductResponse(p))
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

func (s *ProductService) ListCategories(ctx context.Context, includeInactive bool) (model.CategoriesResponse, error) {
	var categories []string
	var err error

	if includeInactive {
		categories, err = s.repo.AdminGetDistinctCategories(ctx)
	} else {
		categories, err = s.repo.GetDistinctCategories(ctx)
	}

	if err != nil {
		return model.CategoriesResponse{}, fmt.Errorf("error in listing categories: %w", err)
	}

	if categories == nil {
		categories = []string{}
	}

	return model.CategoriesResponse{Categories: categories}, nil
}

func toProductResponse(p repository.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:          p.ID.Bytes,
		Name:        p.Name,
		Description: p.Description.String,
		Price:       NumericToString(p.Price),
		Stock:       int(p.Stock),
		SKU:         p.Sku,
		Category:    p.Category,
		IsActive:    p.IsActive,
		ImageURL:    p.ImageUrl.String,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}
