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
		Category:  "apparel",
		ImageUrl:  pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true},
		IsActive:  true,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	req := model.CreateProductRequest{
		Name:        "Kaos Polos",
		Description: "Kaos katun combed 30s",
		Price:       "19.99",
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

	t.Run("empty image_url is forwarded as SQL NULL, not empty string", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.CreateProductParams) (repository.Product, error) {
				assert.False(t, arg.ImageUrl.Valid, "empty image_url must be NULL")
				return sampleProduct(), nil
			})

		_, err := svc.CreateProduct(context.Background(), req)

		require.NoError(t, err)
	})

	t.Run("non-empty image_url is forwarded as-is", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		withImage := req
		withImage.ImageURL = "https://cdn.example.com/kaos.jpg"

		mockRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.CreateProductParams) (repository.Product, error) {
				assert.Equal(t, pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true}, arg.ImageUrl)
				return sampleProduct(), nil
			})

		_, err := svc.CreateProduct(context.Background(), withImage)

		require.NoError(t, err)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	// SKU is deliberately absent: it's immutable after create, and
	// model.UpdateProductRequest has no SKU field to carry one anyway.
	req := model.UpdateProductRequest{
		Name:     "Kaos Polos Updated",
		Price:    "24.5",
		Stock:    5,
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

	t.Run("sku is never part of the update params", func(t *testing.T) {
		// Regression guard: UpdateProductParams has no Sku field at all
		// (compile-time enforced), so an update can never touch it. This
		// pins the intent so a future sqlc regen that reintroduces the
		// field gets caught here instead of silently making SKU mutable.
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.UpdateProductParams) (repository.Product, error) {
				p := sampleProduct()
				p.Category = arg.Category
				return p, nil
			})

		_, err := svc.UpdateProduct(context.Background(), req, id)
		require.NoError(t, err)
	})

	t.Run("empty image_url means delete: forwarded as SQL NULL", func(t *testing.T) {
		// Regression guard: PUT is full-replace, so image_url: "" must clear
		// the image (Valid:false -> NULL), not persist an empty string.
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.UpdateProductParams) (repository.Product, error) {
				assert.False(t, arg.ImageUrl.Valid, "empty image_url must be NULL")
				return sampleProduct(), nil
			})

		_, err := svc.UpdateProduct(context.Background(), req, id)

		require.NoError(t, err)
	})

	t.Run("non-empty image_url is forwarded as-is", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		withImage := req
		withImage.ImageURL = "https://cdn.example.com/kaos.jpg"

		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.UpdateProductParams) (repository.Product, error) {
				assert.Equal(t, pgtype.Text{String: "https://cdn.example.com/kaos.jpg", Valid: true}, arg.ImageUrl)
				return sampleProduct(), nil
			})

		_, err := svc.UpdateProduct(context.Background(), withImage, id)

		require.NoError(t, err)
	})
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

	t.Run("public: success uses the is_active-filtered query", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(sampleProduct(), nil)

		product, err := svc.GetProductByID(context.Background(), id, false)

		require.NoError(t, err)
		require.NotNil(t, product)
		assert.Equal(t, "19.99", product.Price)
		assert.Equal(t, "https://cdn.example.com/kaos.jpg", product.ImageURL)
	})

	t.Run("public: not found (also covers inactive product, filtered at query level)", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(repository.Product{}, pgx.ErrNoRows)

		product, err := svc.GetProductByID(context.Background(), id, false)

		assert.ErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, product)
	})

	t.Run("admin: includeInactive=true hits the unfiltered admin query", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		inactive := sampleProduct()
		inactive.IsActive = false
		mockRepo.EXPECT().AdminGetProductByID(gomock.Any(), gomock.Any()).Return(inactive, nil)
		// The public query must NOT be called on the admin path.

		product, err := svc.GetProductByID(context.Background(), id, true)

		require.NoError(t, err)
		require.NotNil(t, product)
		assert.False(t, product.IsActive)
	})

	t.Run("unexpected db error", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).Return(repository.Product{}, errors.New("connection reset"))

		product, err := svc.GetProductByID(context.Background(), id, false)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, service.ErrProductNotFound)
		assert.Nil(t, product)
	})
}

func toListRow(p repository.Product, total int64) repository.ListProductsRow {
	return repository.ListProductsRow{
		ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price,
		Stock: p.Stock, Sku: p.Sku, Category: p.Category, IsActive: p.IsActive,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, ImageUrl: p.ImageUrl,
		TotalCount: total,
	}
}

func toAdminListRow(p repository.Product, total int64) repository.AdminListProductsRow {
	return repository.AdminListProductsRow{
		ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price,
		Stock: p.Stock, Sku: p.Sku, Category: p.Category, IsActive: p.IsActive,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, ImageUrl: p.ImageUrl,
		TotalCount: total,
	}
}

func TestProductService_ListProducts(t *testing.T) {
	t.Run("empty search/category must reach the repo as NULL, not empty-string", func(t *testing.T) {
		// Regression guard: Valid:true with an empty string turns the SQL
		// filter into `category = ''`, which matches nothing. Empty filters
		// must be forwarded as SQL NULL (Valid:false) to mean "no filter".
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.False(t, arg.Search.Valid, "empty search must be NULL")
				assert.False(t, arg.Category.Valid, "empty category must be NULL")
				return []repository.ListProductsRow{toListRow(sampleProduct(), 1)}, nil
			})

		res, err := svc.ListProducts(context.Background(), "", "", 1, 10, false)

		require.NoError(t, err)
		assert.Len(t, res.Data, 1)
	})

	t.Run("non-empty search/category are forwarded as-is", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, pgtype.Text{String: "kaos", Valid: true}, arg.Search)
				assert.Equal(t, pgtype.Text{String: "apparel", Valid: true}, arg.Category)
				return []repository.ListProductsRow{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		_, err := svc.ListProducts(context.Background(), "kaos", "apparel", 1, 10, false)

		require.NoError(t, err)
	})

	t.Run("page <= 0 defaults to page 1, not a negative offset", func(t *testing.T) {
		// Regression guard: page=0 previously stayed 0, producing
		// offset = (0-1)*limit, a negative OFFSET that Postgres rejects.
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(0), arg.Offset)
				return []repository.ListProductsRow{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 0, 10, false)

		require.NoError(t, err)
		assert.Equal(t, 1, res.Meta.Page)
	})

	t.Run("limit <= 0 defaults to 10", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(10), arg.Limit)
				return []repository.ListProductsRow{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 1, 0, false)

		require.NoError(t, err)
		assert.Equal(t, 10, res.Meta.Limit)
	})

	t.Run("limit=100 is honored exactly, not reset to the default", func(t *testing.T) {
		// Regression guard: a prior `limit >= 100` check reset the boundary
		// value itself to the default (10) instead of clamping above it.
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(100), arg.Limit)
				return []repository.ListProductsRow{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 1, 100, false)

		require.NoError(t, err)
		assert.Equal(t, 100, res.Meta.Limit)
	})

	t.Run("limit > 100 is clamped to 100, not reset to the default", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(100), arg.Limit)
				return []repository.ListProductsRow{}, nil
			})
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 1, 999, false)

		require.NoError(t, err)
		assert.Equal(t, 100, res.Meta.Limit)
	})

	t.Run("pagination metadata is computed from total_count on the returned rows, no fallback count needed", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(5), arg.Offset) // page 2, limit 5 -> offset 5
				return []repository.ListProductsRow{toListRow(sampleProduct(), 23)}, nil
			})
		// CountProducts deliberately NOT EXPECT()'d: total_count already rode
		// along on the row, so no fallback query should fire.

		res, err := svc.ListProducts(context.Background(), "", "", 2, 5, false)

		require.NoError(t, err)
		assert.Equal(t, 2, res.Meta.Page)
		assert.Equal(t, 5, res.Meta.Limit)
		assert.Equal(t, 23, res.Meta.TotalItems)
		assert.Equal(t, 5, res.Meta.TotalPages)
	})

	t.Run("no matching rows at all falls back to CountProducts, which also reports zero", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return([]repository.ListProductsRow{}, nil)
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 1, 10, false)

		require.NoError(t, err)
		assert.Equal(t, 0, res.Meta.TotalItems)
	})

	t.Run("page past the end of the results falls back to CountProducts for the true total", func(t *testing.T) {
		// Regression guard: matches exist (23 of them), but page 5 at
		// limit 10 is past the end, so ListProducts returns zero rows and
		// there's no row to carry total_count. Without the fallback, this
		// would incorrectly report TotalItems = 0 instead of 23.
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return([]repository.ListProductsRow{}, nil)
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(23), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 5, 10, false)

		require.NoError(t, err)
		assert.Empty(t, res.Data)
		assert.Equal(t, 23, res.Meta.TotalItems)
		assert.Equal(t, 3, res.Meta.TotalPages)
	})

	t.Run("list error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection reset"))
		// CountProducts deliberately NOT EXPECT()'d: must short-circuit on the first error.

		_, err := svc.ListProducts(context.Background(), "", "", 1, 10, false)

		assert.Error(t, err)
	})

	t.Run("fallback count error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).Return([]repository.ListProductsRow{}, nil)
		mockRepo.EXPECT().CountProducts(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("connection reset"))

		_, err := svc.ListProducts(context.Background(), "", "", 1, 10, false)

		assert.Error(t, err)
	})

	t.Run("includeInactive=true routes to the admin query", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		inactive := sampleProduct()
		inactive.IsActive = false

		mockRepo.EXPECT().AdminListProducts(gomock.Any(), gomock.Any()).Return([]repository.AdminListProductsRow{toAdminListRow(inactive, 1)}, nil)
		// The public ListProducts/CountProducts must NOT be called on the admin path.

		res, err := svc.ListProducts(context.Background(), "", "", 1, 10, true)

		require.NoError(t, err)
		require.Len(t, res.Data, 1)
		assert.False(t, res.Data[0].IsActive)
		assert.Equal(t, 1, res.Meta.TotalItems)
	})

	t.Run("includeInactive=true page past the end falls back to AdminCountProducts", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().AdminListProducts(gomock.Any(), gomock.Any()).Return([]repository.AdminListProductsRow{}, nil)
		mockRepo.EXPECT().AdminCountProducts(gomock.Any(), gomock.Any()).Return(int64(7), nil)

		res, err := svc.ListProducts(context.Background(), "", "", 5, 10, true)

		require.NoError(t, err)
		assert.Equal(t, 7, res.Meta.TotalItems)
	})
}

func TestProductService_ListCategories(t *testing.T) {
	t.Run("public uses the active-only query", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetDistinctCategories(gomock.Any()).Return([]string{"apparel", "shoes"}, nil)

		res, err := svc.ListCategories(context.Background(), false)

		require.NoError(t, err)
		assert.Equal(t, []string{"apparel", "shoes"}, res.Categories)
	})

	t.Run("admin uses the unfiltered query", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().AdminGetDistinctCategories(gomock.Any()).Return([]string{"apparel", "discontinued"}, nil)

		res, err := svc.ListCategories(context.Background(), true)

		require.NoError(t, err)
		assert.Equal(t, []string{"apparel", "discontinued"}, res.Categories)
	})

	t.Run("nil result is normalized to an empty slice, not null", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetDistinctCategories(gomock.Any()).Return(nil, nil)

		res, err := svc.ListCategories(context.Background(), false)

		require.NoError(t, err)
		assert.NotNil(t, res.Categories)
		assert.Empty(t, res.Categories)
	})

	t.Run("error is propagated", func(t *testing.T) {
		svc, mockRepo := newTestProductService(t)
		mockRepo.EXPECT().GetDistinctCategories(gomock.Any()).Return(nil, errors.New("connection reset"))

		_, err := svc.ListCategories(context.Background(), false)

		assert.Error(t, err)
	})
}
