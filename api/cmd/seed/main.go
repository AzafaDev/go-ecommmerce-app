package main

import (
	"context"
	"fmt"
	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/repository"
	"os"
)

func main() {
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

	fmt.Println("Seeding users...")
	if err := seedUsers(ctx, repo); err != nil {
		fmt.Println("seed users failed:", err)
		os.Exit(1)
	}

	fmt.Println("Seeding products...")
	if err := seedProducts(ctx, repo); err != nil {
		fmt.Println("seed products failed:", err)
		os.Exit(1)
	}

	fmt.Println("\nDemo credentials (password is the same for all):")
	fmt.Printf("  password: %s\n", seedPassword)
	for _, su := range seedUserList {
		role := "customer"
		if su.isAdmin {
			role = "admin"
		}
		fmt.Printf("  %-32s role=%s\n", su.email, role)
	}
}
