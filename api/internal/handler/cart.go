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

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type CartHandler struct {
	srv       *service.CartService
	secretKey string
	validate  *validator.Validate
}

func NewCartHandler(srv *service.CartService, secretKey string) *CartHandler {
	return &CartHandler{
		srv:       srv,
		secretKey: secretKey,
		validate:  validation.New(),
	}
}

func (h *CartHandler) CartRoutes(r chi.Router) {
	r.Route("/cart", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(h.secretKey))
		r.Get("/", h.GetCart)
		r.Delete("/", h.ClearCart)
		r.Route("/items", func(r chi.Router) {
			r.Post("/", h.AddItem)
			r.Patch("/{id}", h.UpdateItemQuantity)
			r.Delete("/{id}", h.RemoveItem)
		})
	})
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("get cart", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	cart, err := h.srv.GetCart(r.Context(), claims.ID)
	if err != nil {
		slog.Error("get cart", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"cart": cart,
		},
	}, http.StatusOK, w)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("add cart item", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	var req model.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("add cart item", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		slog.Error("add cart item", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	item, err := h.srv.AddItem(r.Context(), claims.ID, req)
	if err != nil {
		h.handleCartError(w, "add cart item", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"item": item,
		},
	}, http.StatusOK, w)
}

func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("update cart item", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	productID, ok := h.parseProductID(w, r, "update cart item")
	if !ok {
		return
	}

	var req model.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("update cart item", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		slog.Error("update cart item", "error", err)
		response.WriteErrorJSON("invalid payload request", http.StatusBadRequest, w)
		return
	}

	item, err := h.srv.UpdateItemQuantity(r.Context(), claims.ID, productID, req.Quantity)
	if err != nil {
		h.handleCartError(w, "update cart item", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"item": item,
		},
	}, http.StatusOK, w)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("remove cart item", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	productID, ok := h.parseProductID(w, r, "remove cart item")
	if !ok {
		return
	}

	if err := h.srv.RemoveItem(r.Context(), claims.ID, productID); err != nil {
		h.handleCartError(w, "remove cart item", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "removed cart item successfully",
		},
	}, http.StatusOK, w)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("clear cart", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	if err := h.srv.ClearCart(r.Context(), claims.ID); err != nil {
		slog.Error("clear cart", "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]string{
			"message": "cleared cart successfully",
		},
	}, http.StatusOK, w)
}

func (h *CartHandler) parseProductID(w http.ResponseWriter, r *http.Request, logCtx string) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		slog.Error(logCtx, "error", fmt.Errorf("product id is empty"))
		response.WriteErrorJSON("product id is empty", http.StatusBadRequest, w)
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error(logCtx, "error", fmt.Errorf("error in parsing string to be uuid"))
		response.WriteErrorJSON("id is invalid", http.StatusBadRequest, w)
		return uuid.UUID{}, false
	}
	return id, true
}

func (h *CartHandler) handleCartError(w http.ResponseWriter, logCtx string, err error) {
	switch err {
	case service.ErrProductNotFound:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("product not found", http.StatusNotFound, w)
	case service.ErrCartItemNotFound:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("cart item not found", http.StatusNotFound, w)
	case service.ErrInsufficientStock:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("insufficient stock", http.StatusConflict, w)
	default:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
	}
}
