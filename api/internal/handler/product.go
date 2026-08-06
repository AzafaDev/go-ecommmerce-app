package handler

import (
	"encoding/json"
	"fmt"
	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/response"
	"go-ecommerce-app/pkg/validation"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type ProductHandler struct {
	srv       *service.ProductService
	secretKey string
	validate  *validator.Validate
}

func NewProductHandler(srv *service.ProductService, secretKey string) *ProductHandler {
	return &ProductHandler{
		srv:       srv,
		secretKey: secretKey,
		validate:  validation.New(),
	}
}

func (h *ProductHandler) ProductRoutes(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/categories", h.ListCategories)
		r.Get("/{id}", h.GetProductByID)
		r.Get("/", h.ListProducts)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(h.secretKey))
			r.Use(middleware.RequireRole("admin"))
			r.Post("/", h.CreateProduct)
			r.Put("/{id}", h.UpdateProduct)
			r.Delete("/{id}", h.DeleteProduct)
		})
	})

	r.Route("/admin/products", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(h.secretKey))
		r.Use(middleware.RequireRole("admin"))
		r.Get("/categories", h.AdminListCategories)
		r.Get("/{id}", h.AdminGetProductByID)
		r.Get("/", h.AdminListProducts)
	})
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("create product", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		slog.Error("create product", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	createdProduct, err := h.srv.CreateProduct(r.Context(), req)
	if err != nil {
		if err == service.ErrSKUTaken {
			slog.Error("create product", "error", err)
			response.WriteErrorJSON("SKU product is already existed", http.StatusConflict, w)
			return
		}

		slog.Error("create product", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"product": createdProduct,
		},
	}, http.StatusCreated, w)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		slog.Error("update product", "error", fmt.Errorf("id product is empty"))
		response.WriteErrorJSON("id product is empty", http.StatusBadRequest, w)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("update product", "error", fmt.Errorf("error in parsing string to be uuid"))
		response.WriteErrorJSON("id is invalid", http.StatusBadRequest, w)
		return
	}
	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("update product", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		slog.Error("update product", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	updatedProduct, err := h.srv.UpdateProduct(r.Context(), req, id)
	if err != nil {
		if err == service.ErrProductNotFound {
			slog.Error("update product", "error", err)
			response.WriteErrorJSON("product not found", http.StatusNotFound, w)
			return
		}

		slog.Error("update product", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"product": updatedProduct,
		},
	}, http.StatusOK, w)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		slog.Error("delete product", "error", fmt.Errorf("id product is empty"))
		response.WriteErrorJSON("id product is empty", http.StatusBadRequest, w)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("delete product", "error", fmt.Errorf("error in parsing string to be uuid"))
		response.WriteErrorJSON("id is invalid", http.StatusBadRequest, w)
		return
	}

	err = h.srv.DeleteProduct(r.Context(), id)
	if err != nil {
		if err == service.ErrProductNotFound {
			slog.Error("delete product", "error", err)
			response.WriteErrorJSON("product not found", http.StatusNotFound, w)
			return
		}

		slog.Error("delete product", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "deleted product successfully",
		},
	}, http.StatusOK, w)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	h.getProductByID(w, r, false)
}

func (h *ProductHandler) AdminGetProductByID(w http.ResponseWriter, r *http.Request) {
	h.getProductByID(w, r, true)
}

func (h *ProductHandler) getProductByID(w http.ResponseWriter, r *http.Request, includeInactive bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		slog.Error("get product by id", "error", fmt.Errorf("id product is empty"))
		response.WriteErrorJSON("id product is empty", http.StatusBadRequest, w)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("get product by id", "error", fmt.Errorf("error in parsing string to be uuid"))
		response.WriteErrorJSON("id is invalid", http.StatusBadRequest, w)
		return
	}

	existingProduct, err := h.srv.GetProductByID(r.Context(), id, includeInactive)
	if err != nil {
		if err == service.ErrProductNotFound {
			slog.Error("get product by id", "error", err)
			response.WriteErrorJSON("product not found", http.StatusNotFound, w)
			return
		}

		slog.Error("get product by id", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"product": existingProduct,
		},
	}, http.StatusOK, w)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	h.listProducts(w, r, false)
}

func (h *ProductHandler) AdminListProducts(w http.ResponseWriter, r *http.Request) {
	h.listProducts(w, r, true)
}

func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request, includeInactive bool) {
	search := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")

	page := 1
	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil {
			slog.Error("list products", "error", err)
			response.WriteErrorJSON("invalid page format", http.StatusBadRequest, w)
			return
		}
		page = parsedPage
	}

	limit := 10
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			slog.Error("list products", "error", err)
			response.WriteErrorJSON("invalid limit format", http.StatusBadRequest, w)
			return
		}
		limit = parsedLimit
	}

	listProductsRes, err := h.srv.ListProducts(r.Context(), search, category, page, limit, includeInactive)
	if err != nil {
		slog.Error("list products", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"products": listProductsRes.Data,
			"meta":     listProductsRes.Meta,
		},
	}, http.StatusOK, w)
}

func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	h.listCategories(w, r, false)
}

func (h *ProductHandler) AdminListCategories(w http.ResponseWriter, r *http.Request) {
	h.listCategories(w, r, true)
}

func (h *ProductHandler) listCategories(w http.ResponseWriter, r *http.Request, includeInactive bool) {
	categoriesRes, err := h.srv.ListCategories(r.Context(), includeInactive)
	if err != nil {
		slog.Error("list categories", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"categories": categoriesRes.Categories,
		},
	}, http.StatusOK, w)
}
