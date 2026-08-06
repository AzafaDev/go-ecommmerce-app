package model_test

import (
	"testing"

	"go-ecommerce-app/internal/model"
	"go-ecommerce-app/pkg/validation"

	"github.com/stretchr/testify/assert"
)

func validProductPayload() model.CreateProductRequest {
	return model.CreateProductRequest{
		Name:        "Kaos Polos",
		Description: "Kaos katun combed 30s",
		Price:       "19.99",
		Stock:       10,
		SKU:         "SKU-001",
		Category:    "apparel",
	}
}

func TestCreateProductRequest_Validate(t *testing.T) {
	v := validation.New()

	t.Run("valid payload passes", func(t *testing.T) {
		assert.NoError(t, v.Struct(validProductPayload()))
	})

	t.Run("stock zero is a legitimate new-product state", func(t *testing.T) {
		// Regression guard: "required" on an int field rejects the zero
		// value, which wrongly blocked creating a product with no stock yet.
		req := validProductPayload()
		req.Stock = 0
		assert.NoError(t, v.Struct(req))
	})

	t.Run("negative stock is rejected", func(t *testing.T) {
		req := validProductPayload()
		req.Stock = -1
		assert.Error(t, v.Struct(req))
	})

	t.Run("zero price is rejected", func(t *testing.T) {
		// The DB CHECK constraint requires price > 0; validation should
		// reject it up front with a clean error instead of a raw DB error.
		req := validProductPayload()
		req.Price = "0.00"
		assert.Error(t, v.Struct(req))
	})

	t.Run("negative price is rejected", func(t *testing.T) {
		req := validProductPayload()
		req.Price = "-5"
		assert.Error(t, v.Struct(req))
	})

	t.Run("price with more than 2 decimal digits is rejected", func(t *testing.T) {
		// Regression guard for the hardening gap: %.2f used to silently
		// round "19.995" down to "20.00" with no error back to the client.
		req := validProductPayload()
		req.Price = "19.995"
		assert.Error(t, v.Struct(req))
	})

	t.Run("non-numeric price is rejected", func(t *testing.T) {
		req := validProductPayload()
		req.Price = "abc"
		assert.Error(t, v.Struct(req))
	})

	t.Run("empty category is rejected", func(t *testing.T) {
		// Category is sourced from a dropdown on the client, so an empty
		// value should never reach the API in normal use — but the server
		// must not trust that and should reject it independently.
		req := validProductPayload()
		req.Category = ""
		assert.Error(t, v.Struct(req))
	})

	t.Run("empty image_url is allowed", func(t *testing.T) {
		req := validProductPayload()
		req.ImageURL = ""
		assert.NoError(t, v.Struct(req))
	})

	t.Run("invalid image_url is rejected", func(t *testing.T) {
		req := validProductPayload()
		req.ImageURL = "not-a-url"
		assert.Error(t, v.Struct(req))
	})

	t.Run("valid image_url passes", func(t *testing.T) {
		req := validProductPayload()
		req.ImageURL = "https://cdn.example.com/kaos.jpg"
		assert.NoError(t, v.Struct(req))
	})
}

func validUpdateProductPayload() model.UpdateProductRequest {
	return model.UpdateProductRequest{
		Name:     "Kaos Polos Updated",
		Price:    "24.5",
		Stock:    5,
		Category: "apparel",
		IsActive: true,
	}
}

func TestUpdateProductRequest_Validate(t *testing.T) {
	v := validation.New()

	t.Run("valid payload passes", func(t *testing.T) {
		assert.NoError(t, v.Struct(validUpdateProductPayload()))
	})

	t.Run("empty category is rejected", func(t *testing.T) {
		req := validUpdateProductPayload()
		req.Category = ""
		assert.Error(t, v.Struct(req))
	})

	t.Run("empty image_url is allowed (means: remove image)", func(t *testing.T) {
		req := validUpdateProductPayload()
		req.ImageURL = ""
		assert.NoError(t, v.Struct(req))
	})
}
