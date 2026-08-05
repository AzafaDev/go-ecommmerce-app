package handler

import (
	"encoding/json"
	"fmt"
	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/response"
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
		validate:  validator.New(),
	}
}

func (h *ProductHandler) ProductRoutes(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
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
			response.WriteErrorJSON("SKU product is already existed", http.StatusForbidden, w)
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
	idStr := r.PathValue("id")
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
		if err == service.ErrSKUTaken {
			slog.Error("update product", "error", err)
			response.WriteErrorJSON("SKU product is already existed", http.StatusForbidden, w)
			return
		}

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
	idStr := r.PathValue("id")
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
	idStr := r.PathValue("id")
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

	existingProduct, err := h.srv.GetProductByID(r.Context(), id)
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

	listProductsRes, err := h.srv.ListProducts(r.Context(), search, category, page, limit)
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
