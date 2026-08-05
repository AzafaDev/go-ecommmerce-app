package main

import (
	"context"
	"fmt"
	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/repository"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: go run ./cmd/seed-admin <email>")
		os.Exit(1)
	}

	email := strings.ToLower(strings.TrimSpace(os.Args[1]))

	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}

	dbPool, err := config.NewDatabase(cfg.DatabaseURL, ctx)
	if err != nil {
		fmt.Println("database error:", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	repo := repository.New(dbPool)

	user, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("user not found:", err)
		os.Exit(1)
	}

	if user.Role == "admin" {
		fmt.Printf("%s is already an admin\n", email)
		return
	}

	updatedUser, err := repo.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
		Role: "admin",
		ID:   user.ID,
	})
	if err != nil {
		fmt.Println("failed to update role:", err)
		os.Exit(1)
	}

	fmt.Printf("%s (%s) is now role=%s\n", updatedUser.Email, updatedUser.ID, updatedUser.Role)
}
