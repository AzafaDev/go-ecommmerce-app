package handler

import (
	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/internal/service"

	"github.com/go-chi/chi"
)

type ProductHandler struct {
	srv       *service.ProductService
	secretKey string
}

func NewProductHandler(srv *service.ProductService, secretKey string) *ProductHandler {
	return &ProductHandler{
		srv:       srv,
		secretKey: secretKey,
	}
}

func (h *ProductHandler) ProductRoutes(r chi.Router) {
	r.Route("/product", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(h.secretKey))
			r.Use(middleware.RequireRole("admin"))
		})
	})
}
