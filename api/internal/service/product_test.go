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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestProductService(t *testing.T) (*service.ProductService, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	svc := service.NewProductService(mockRepo)
	return svc, mockRepo
}

func sampleProduct() repository.Product {
	return repository.Product{
		ID:          randomUUID(),
		Name:        "Kaos Polos",
		Description: pgtype.Text{String: "Kaos katun combed 30s", Valid: true},
		// 19.99 is the regression case for the pgtype.Numeric -> string bug:
		// reading .Int directly (dropping .Exp) turns this into "1999".
		Price:     service.Float64ToNumeric(19.99),
		Stock:     10,
		Sku:       "SKU-001",
		Category:  pgtype.Text{String: "apparel", Valid: true},
		IsActive:  true,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	req := model.CreateProductRequest{
		Name:        "Kaos Polos",
		Description: "Kaos katun combed 30s",
		Price:       19.99,
		Stock:       3,
		SKU:         "SKU-001",
		Category:    "apparel",
	}

	tests := []struct {
		name      string
		setup     func(repo *repomocks.MockQuerier)
		assertErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(sampleProduct(), nil)
			},
		},
		{
			name: "sku already taken",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
					Return(repository.Product{}, &pgconn.PgError{Code: "23505"})
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrSKUTaken)
			},
		},
		{
			name: "unexpected db error",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
					Return(repository.Product{}, errors.New("connection reset"))
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.NotErrorIs(t, err, service.ErrSKUTaken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := newTestProductService(t)
			tt.setup(mockRepo)

			product, err := svc.CreateProduct(context.Background(), req)

			if tt.assertErr != nil {
				tt.assertErr(t, err)
				assert.Nil(t, product)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, product)
			// Regression guard: Price must be a proper decimal string, not the
			// unscaled pgtype.Numeric.Int (which would render "19.99" as "1999").
			assert.Equal(t, "19.99", product.Price)
			assert.Equal(t, "SKU-001", product.SKU)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	req := model.UpdateProductRequest{
		Name:     "Kaos Polos Updated",
		Price:    24.5,
		Stock:    5,
		SKU:      "SKU-001",
		Category: "apparel",
		IsActive: true,
	}
	id := uuid.New()

	tests := []struct {
		name      string
		setup     func(repo *repomocks.MockQuerier)
		assertErr func(t *testing.T, err error)
	}{
		{
			name: "success",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).Return(sampleProduct(), nil)
			},
		},
		{
			name: "sku already taken by another product",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
					Return(repository.Product{}, &pgconn.PgError{Code: "23505"})
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrSKUTaken)
			},
		},
		{
			name: "product not found",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
					Return(repository.Product{}, pgx.ErrNoRows)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrProductNotFound)
			},
		},
		{
			name: "unexpected db error",
			setup: func(repo *repomocks.MockQuerier) {
				repo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
					Return(repository.Product{}, errors.New("connection reset"))
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.NotErrorIs(t, err, service.ErrSKUTaken)
				assert.NotErrorIs(t, err, service.ErrProductNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := newTestProductService(t)
			tt.setup(mockRepo)

			product, err := svc.UpdateProduct(context.Background(), req, id)

			if tt.assertErr != nil {
				tt.assertErr(t, err)
				assert.Nil(t, product)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, product)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().SoftDeleteProduct(gomock.Any(), gomock.Any()).Return(sampleProduct(), nil)

		err := svc.DeleteProduct(context.Background(), id)

		require.NoError(t, err)
	})

	t.Run("product not found", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().SoftDeleteProduct(gomock.Any(), gomock.Any()).Return(repository.Product{}, pgx.ErrNoRows)

		err := svc.DeleteProduct(context.Background(), id)

		assert.ErrorIs(t, err, service.ErrProductNotFound)
	})

	t.Run("unexpected db error", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().SoftDeleteProduct(gomock.Any(), gomock.Any()).Return(repository.Product{}, errors.New("connection reset"))

		err := svc.DeleteProduct(context.Background(), id)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, service.ErrProductNotFound)
	})
}

func TestProductService_GetProductByID(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(sampleProduct(), nil)

		product, err := svc.GetProductByID(context.Background(), id)

		require.NoError(t, err)
		require.NotNil(t, product)
		assert.Equal(t, "19.99", product.Price)
	})

	t.Run("product not found", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(repository.Product{}, pgx.ErrNoRows)

		product, err := svc.GetProductByID(context.Background(), id)

		assert.ErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, product)
	})

	t.Run("unexpected db error", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(repository.Product{}, errors.New("connection reset"))

		product, err := svc.GetProductByID(context.Background(), id)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, product)
	})
}

func TestProductService_ListProducts(t *testing.T) {
	t.Run("empty search/category must reach the repo as NULL, not empty-string", func(t *testing.T) {
		// Regression guard: Valid:true with an empty string turns the SQL
		// filter into `category = ''`, which matches nothing. Empty filters
		// must be forwarded as SQL NULL (Valid:false) to mean "no filter".
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.Product, error) {
				assert.False(t, arg.Search.Valid, "empty search must be NULL")
				assert.False(t, arg.Category.Valid, "empty category must be NULL")
				return []repository.Product{sampleProduct()}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.CountProductsParams) (int64, error) {
				assert.False(t, arg.Search.Valid)
				assert.False(t, arg.Category.Valid)
				return 1, nil
			})

		res, err := svc.ListProducts(context.Background(), "", "", 1, 10)

		require.NoError(t, err)
		assert.Len(t, res.Data, 1)
	})

	t.Run("non-empty search/category are forwarded as-is", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.Product, error) {
				assert.Equal(t, pgtype.Text{String: "kaos", Valid: true}, arg.Search)
				assert.Equal(t, pgtype.Text{String: "apparel", Valid: true}, arg.Category)
				return []repository.Product{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		_, err := svc.ListProducts(context.Background(), "kaos", "apparel", 1, 10)

		require.NoError(t, err)
	})

	t.Run("page <= 0 defaults to page 1, not a negative offset", func(t *testing.T) {
		// Regression guard: page=0 previously stayed 0, producing
		// offset = (0-1)*limit, a negative OFFSET that Postgres rejects.
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.Product, error) {
				assert.Equal(t, int32(0), arg.Offset)
				return []repository.Product{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 0, 10)

		require.NoError(t, err)
		assert.Equal(t, 1, res.Meta.Page)
	})

	t.Run("pagination metadata is computed from total items", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.Product, error) {
				assert.Equal(t, int32(5), arg.Offset) // page 2, limit 5 -> offset 5
				return []repository.Product{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(23), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 2, 5)

		require.NoError(t, err)
		assert.Equal(t, 2, res.Meta.Page)
		assert.Equal(t, 5, res.Meta.Limit)
		assert.Equal(t, 23, res.Meta.TotalItems)
		assert.Equal(t, 5, res.Meta.TotalPages)
	})

	t.Run("list error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))
		// CountProducts deliberately NOT EXPECT()'d: must short-circuit on the first error.

		_, err := svc.ListProducts(context.Background(), "", "", 1, 10)

		assert.Error(t, err)
	})

	t.Run("count error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return([]repository.Product{}, nil)
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("connection reset"))

		_, err := svc.ListProducts(context.Background(), "", "", 1, 10)

		assert.Error(t, err)
	})
}
