package handler

import (
	"encoding/json"
	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/internal/service"
	"go-ecommerce-app/pkg/response"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type OrderHandler struct {
	srv       *service.OrderService
	secretKey string
}

func NewOrderHandler(srv *service.OrderService, secretKey string) *OrderHandler {
	return &OrderHandler{
		srv:       srv,
		secretKey: secretKey,
	}
}

func (h *OrderHandler) OrderRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		// No AuthMiddleware: Midtrans authenticates via signature_key instead (see HandleWebhook).
		r.Post("/webhook", h.Webhook)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(h.secretKey))
			r.Post("/checkout", h.Checkout)
			r.Get("/", h.ListOrders)
			r.Get("/{id}", h.GetOrder)
		})
	})
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("checkout", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	order, err := h.srv.Checkout(r.Context(), claims.ID)
	if err != nil {
		h.handleOrderError(w, "checkout", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"order": order,
		},
	}, http.StatusCreated, w)
}

func (h *OrderHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var notif model.MidtransNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		slog.Error("order webhook", "error", err)
		response.WriteErrorJSON("invalid payload", http.StatusBadRequest, w)
		return
	}

	if err := h.srv.HandleWebhook(r.Context(), notif); err != nil {
		h.handleOrderError(w, "order webhook", err)
		return
	}

	response.WriteJSON(response.JSONResponse{Success: true}, http.StatusOK, w)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("list orders", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	orders, err := h.srv.ListOrders(r.Context(), claims.ID)
	if err != nil {
		h.handleOrderError(w, "list order", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"orders": orders,
		},
	}, http.StatusOK, w)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("get order", "error", err)
		response.WriteErrorJSON("order id is invalid", http.StatusBadRequest, w)
		return
	}

	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("get order", "error", err)
		response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
		return
	}

	order, err := h.srv.GetOrder(r.Context(), claims.ID, id)
	if err != nil {
		h.handleOrderError(w, "get order", err)
		return
	}

	response.WriteJSON(response.JSONResponse{
		Success: true,
		Data: map[string]any{
			"order": order,
		},
	}, http.StatusOK, w)
}

func (h *OrderHandler) handleOrderError(w http.ResponseWriter, logCtx string, err error) {
	switch err {
	case service.ErrCartEmpty:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("cart is empty", http.StatusBadRequest, w)
	case service.ErrInsufficientStock:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("insufficient stock", http.StatusConflict, w)
	case service.ErrOrderNotFound:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("order not found", http.StatusNotFound, w)
	case service.ErrInvalidSignature:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("invalid signature", http.StatusForbidden, w)
	default:
		slog.Error(logCtx, "error", err)
		response.WriteErrorJSON("something went wrong", http.StatusInternalServerError, w)
	}
}
