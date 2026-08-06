package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=2000"`
	Price       string `json:"price" validate:"required,price"`
	Stock       int    `json:"stock" validate:"gte=0"`
	SKU         string `json:"sku" validate:"required,min=3,max=50"`
	Category    string `json:"category" validate:"required,max=100"`
	ImageURL    string `json:"image_url" validate:"omitempty,url"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=2000"`
	Price       string `json:"price" validate:"required,price"`
	Stock       int    `json:"stock" validate:"gte=0"`
	Category    string `json:"category" validate:"required,max=100"`
	IsActive    bool   `json:"is_active"`
	ImageURL    string `json:"image_url" validate:"omitempty,url"`
}

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       string    `json:"price"`
	Stock       int       `json:"stock"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListProductsResponse struct {
	Data []ProductResponse       `json:"data"`
	Meta MetaListProductResponse `json:"meta"`
}

type MetaListProductResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type CategoriesResponse struct {
	Categories []string `json:"categories"`
}
