package main

import (
	"context"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/pkg/security"

	"github.com/jackc/pgx/v5"
)

func seedUsers(ctx context.Context, repo *repository.Queries) error {
	hashedPassword, err := security.HashPassword(seedPassword)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	for _, su := range seedUserList {
		if _, err := repo.GetUserByEmail(ctx, su.email); err == nil {
			fmt.Printf("  skip user  %-32s (already exists)\n", su.email)
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check user %s: %w", su.email, err)
		}

		created, err := repo.CreateUser(ctx, repository.CreateUserParams{
			FullName:     su.fullName,
			Email:        su.email,
			PasswordHash: hashedPassword,
		})
		if err != nil {
			return fmt.Errorf("create user %s: %w", su.email, err)
		}

		if _, err := repo.SetUserVerified(ctx, created.ID); err != nil {
			return fmt.Errorf("verify user %s: %w", su.email, err)
		}

		role := "customer"
		if su.isAdmin {
			role = "admin"
			if _, err := repo.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
				Role: role,
				ID:   created.ID,
			}); err != nil {
				return fmt.Errorf("promote user %s: %w", su.email, err)
			}
		}

		fmt.Printf("  created    %-32s (role=%s)\n", su.email, role)
	}

	return nil
}
