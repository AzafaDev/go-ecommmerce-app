package model_test

import (
	"testing"

	"go-ecommerce-app/internal/model"

	"github.com/go-playground/validator"
	"github.com/stretchr/testify/assert"
)

func validProductPayload() model.CreateProductRequest {
	return model.CreateProductRequest{
		Name:        "Kaos Polos",
		Description: "Kaos katun combed 30s",
		Price:       19.99,
		Stock:       10,
		SKU:         "SKU-001",
		Category:    "apparel",
	}
}

func TestCreateProductRequest_Validate(t *testing.T) {
	v := validator.New()

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
		req.Price = 0
		assert.Error(t, v.Struct(req))
	})

	t.Run("negative price is rejected", func(t *testing.T) {
		req := validProductPayload()
		req.Price = -5
		assert.Error(t, v.Struct(req))
	})
}
