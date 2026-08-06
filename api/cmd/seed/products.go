package main

import (
	"context"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func seedProducts(ctx context.Context, repo *repository.Queries) error {
	for _, sp := range seedProductList {
		// products.sku is only uniquely indexed among active rows, so a
		// soft-deleted (inactive) seed product wouldn't be caught by a
		// unique-violation retry — check explicitly instead, active or not.
		if _, err := repo.AdminGetProductBySKU(ctx, sp.sku); err == nil {
			fmt.Printf("  skip product %-24s (already exists)\n", sp.sku)
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check product %s: %w", sp.sku, err)
		}

		price, err := service.StringToNumeric(sp.price)
		if err != nil {
			return fmt.Errorf("parse price for %s: %w", sp.sku, err)
		}

		imageURL := fmt.Sprintf("https://picsum.photos/seed/%s/600/600", strings.ToLower(sp.sku))

		created, err := repo.CreateProduct(ctx, repository.CreateProductParams{
			Name:        sp.name,
			Description: pgtype.Text{String: sp.desc, Valid: true},
			Price:       price,
			Stock:       sp.stock,
			Sku:         sp.sku,
			Category:    sp.category,
			ImageUrl:    pgtype.Text{String: imageURL, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("create product %s: %w", sp.sku, err)
		}

		status := "active"
		if sp.inactive {
			if _, err := repo.SoftDeleteProduct(ctx, created.ID); err != nil {
				return fmt.Errorf("soft-delete product %s: %w", sp.sku, err)
			}
			status = "inactive"
		}

		fmt.Printf("  created      %-24s stock=%-3d %s\n", sp.sku, sp.stock, status)
	}

	return nil
}
