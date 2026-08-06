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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const productTestSecretKey = "test-jwt-secret"

// newTestProductRouter mounts ProductRoutes on a real chi.Mux, so path
// params (chi.URLParam) and the admin auth/role middleware chain behave
// exactly like production instead of being bypassed by calling handler
// methods directly.
func newTestProductRouter(t *testing.T) (*chi.Mux, *repomocks.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := repomocks.NewMockQuerier(ctrl)
	svc := service.NewProductService(mockRepo)
	h := handler.NewProductHandler(svc, productTestSecretKey)

	r := chi.NewRouter()
	h.ProductRoutes(r)
	return r, mockRepo
}

func adminToken(t *testing.T) string {
	t.Helper()
	token, err := security.GenerateToken(productTestSecretKey, time.Hour, uuid.New(), "admin")
	require.NoError(t, err)
	return token
}

func customerToken(t *testing.T) string {
	t.Helper()
	token, err := security.GenerateToken(productTestSecretKey, time.Hour, uuid.New(), "customer")
	require.NoError(t, err)
	return token
}

func sampleRepoProduct(id pgtype.UUID) repository.Product {
	return repository.Product{
		ID:          id,
		Name:        "Kaos Polos",
		Description: pgtype.Text{String: "Kaos katun combed 30s", Valid: true},
		Price:       service.Float64ToNumeric(19.99),
		Stock:       10,
		Sku:         "SKU-001",
		Category:    "apparel",
		IsActive:    true,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
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

func TestProductHandler_GetProductByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).
			Return(sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true}), nil)

		req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// Regression guard: this is the exact request shape that silently
		// 400'd on every call before switching id extraction from
		// r.PathValue (a net/http ServeMux feature chi v1 never populates)
		// to chi.URLParam.
		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["product"])
	})

	t.Run("not found", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).
			Return(repository.Product{}, pgx.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
		got := decodeResponse(t, rec)
		assert.False(t, got.Success)
	})

	t.Run("inactive product is not found on the public route", func(t *testing.T) {
		// The public query filters is_active=true at the SQL level, so from
		// the handler's point of view an inactive product looks identical
		// to a missing one: pgx.ErrNoRows.
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).
			Return(repository.Product{}, pgx.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid uuid returns 400", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/products/not-a-uuid", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("no auth header required for public GET", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()

		mockRepo.EXPECT().GetProductByID(gomock.Any(), gomock.Any()).
			Return(sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true}), nil)

		req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
		// deliberately no Authorization header
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestProductHandler_AdminGetProductByID(t *testing.T) {
	t.Run("admin can see an inactive product", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		inactive := sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true})
		inactive.IsActive = false

		mockRepo.EXPECT().AdminGetProductByID(gomock.Any(), gomock.Any()).Return(inactive, nil)

		req := httptest.NewRequest(http.MethodGet, "/admin/products/"+id.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		id := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/admin/products/"+id.String(), nil)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		id := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/admin/products/"+id.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestProductHandler_ListProducts(t *testing.T) {
	t.Run("success with default pagination", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)

		mockRepo.EXPECT().ListProducts(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, arg repository.ListProductsParams) ([]repository.ListProductsRow, error) {
				assert.Equal(t, int32(10), arg.Limit)
				assert.Equal(t, int32(0), arg.Offset)
				return []repository.ListProductsRow{toListRow(sampleRepoProduct(pgtype.UUID{Bytes: uuid.New(), Valid: true}), 1)}, nil
			})

		req := httptest.NewRequest(http.MethodGet, "/products/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["products"])
		assert.NotNil(t, got.Data["meta"])
	})

	t.Run("invalid page format returns 400", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/products/?page=notanumber", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestProductHandler_AdminListProducts(t *testing.T) {
	t.Run("admin sees inactive products too", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		inactive := sampleRepoProduct(pgtype.UUID{Bytes: uuid.New(), Valid: true})
		inactive.IsActive = false

		mockRepo.EXPECT().AdminListProducts(gomock.Any(), gomock.Any()).Return([]repository.AdminListProductsRow{toAdminListRow(inactive, 1)}, nil)

		req := httptest.NewRequest(http.MethodGet, "/admin/products/", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/admin/products/", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestProductHandler_ListCategories(t *testing.T) {
	t.Run("public: no auth required", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		mockRepo.EXPECT().GetDistinctCategories(gomock.Any()).Return([]string{"apparel"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/products/categories", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["categories"])
	})

	t.Run("admin: requires admin role and hits the unfiltered query", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		mockRepo.EXPECT().AdminGetDistinctCategories(gomock.Any()).Return([]string{"apparel", "discontinued"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/admin/products/categories", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("admin categories: non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/admin/products/categories", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestProductHandler_CreateProduct(t *testing.T) {
	validPayload := model.CreateProductRequest{
		Name:     "Kaos Polos",
		Price:    "19.99",
		Stock:    10,
		SKU:      "SKU-001",
		Category: "apparel",
	}

	t.Run("admin can create, gets 201", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		mockRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
			Return(sampleRepoProduct(pgtype.UUID{Bytes: uuid.New(), Valid: true}), nil)

		req := newJSONRequest(t, http.MethodPost, "/products/", validPayload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		got := decodeResponse(t, rec)
		assert.True(t, got.Success)
		assert.NotNil(t, got.Data["product"])
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		// CreateProduct deliberately NOT EXPECT()'d: RequireRole must block
		// the request before it ever reaches the service.

		req := newJSONRequest(t, http.MethodPost, "/products/", validPayload)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("missing token is unauthorized", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := newJSONRequest(t, http.MethodPost, "/products/", validPayload)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("sku already taken returns 409, not 403", func(t *testing.T) {
		// Regression guard: this used to return 403 Forbidden, which is
		// semantically wrong (403 = authorization, this is a data conflict)
		// and inconsistent with user.go's "email already taken" -> 409.
		router, mockRepo := newTestProductRouter(t)
		mockRepo.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).
			Return(repository.Product{}, &pgconn.PgError{Code: "23505"})

		req := newJSONRequest(t, http.MethodPost, "/products/", validPayload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
		got := decodeResponse(t, rec)
		assert.False(t, got.Success)
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		router, _ := newTestProductRouter(t)

		req := newJSONRequest(t, http.MethodPost, "/products/", model.CreateProductRequest{})
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing category returns 400", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		payload := validPayload
		payload.Category = ""

		req := newJSONRequest(t, http.MethodPost, "/products/", payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("price with more than 2 decimal digits returns 400", func(t *testing.T) {
		// Regression guard: this is the actual hardening gap — the API must
		// reject imprecise prices instead of silently rounding them.
		router, _ := newTestProductRouter(t)
		payload := validPayload
		payload.Price = "19.995"

		req := newJSONRequest(t, http.MethodPost, "/products/", payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid image_url returns 400", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		payload := validPayload
		payload.ImageURL = "not-a-url"

		req := newJSONRequest(t, http.MethodPost, "/products/", payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	// SKU is deliberately absent from the payload: model.UpdateProductRequest
	// has no SKU field, so it can't be sent even by a misbehaving client.
	payload := model.UpdateProductRequest{
		Name:     "Kaos Polos Updated",
		Price:    "24.5",
		Stock:    5,
		Category: "apparel",
		IsActive: true,
	}

	t.Run("admin can update", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			Return(sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true}), nil)

		req := newJSONRequest(t, http.MethodPut, "/products/"+id.String(), payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("admin can reactivate a soft-deleted product via is_active=true", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, arg repository.UpdateProductParams) (repository.Product, error) {
				assert.True(t, arg.IsActive)
				p := sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true})
				p.IsActive = true
				return p, nil
			})

		req := newJSONRequest(t, http.MethodPut, "/products/"+id.String(), payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		mockRepo.EXPECT().UpdateProduct(gomock.Any(), gomock.Any()).
			Return(repository.Product{}, pgx.ErrNoRows)

		req := newJSONRequest(t, http.MethodPut, "/products/"+id.String(), payload)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		id := uuid.New()

		req := newJSONRequest(t, http.MethodPut, "/products/"+id.String(), payload)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	t.Run("admin can delete", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		mockRepo.EXPECT().SoftDeleteProduct(gomock.Any(), gomock.Any()).
			Return(sampleRepoProduct(pgtype.UUID{Bytes: id, Valid: true}), nil)

		req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		router, mockRepo := newTestProductRouter(t)
		id := uuid.New()
		mockRepo.EXPECT().SoftDeleteProduct(gomock.Any(), gomock.Any()).
			Return(repository.Product{}, pgx.ErrNoRows)

		req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		router, _ := newTestProductRouter(t)
		id := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
		req.Header.Set("Authorization", "Bearer "+customerToken(t))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}
