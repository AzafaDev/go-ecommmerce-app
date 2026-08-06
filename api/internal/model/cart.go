package model

import "github.com/google/uuid"

type AddCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type CartItemResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Price     string    `json:"price"`
	ImageURL  string    `json:"image_url"`
	IsActive  bool      `json:"is_active"`
	Quantity  int       `json:"quantity"`
	Subtotal  string    `json:"subtotal"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	Total      string             `json:"total"`
}
