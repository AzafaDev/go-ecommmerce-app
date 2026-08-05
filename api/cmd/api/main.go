package main

import (
	"context"
	"fmt"
	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/email"
	"go-ecommerce-app/internal/handler"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	dbPool, err := config.NewDatabase(cfg.DatabaseURL, ctx)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis", "error", err)
		os.Exit(1)
	}

	senderEmail := email.NewResendClient(cfg.ResendAPIKey, cfg.ResendFromAddress, cfg.FrontendURL)

	repo := repository.New(dbPool)

	userService := service.NewUserService(repo, cfg, senderEmail)
	userHandler := handler.NewUserHandler(userService, cfg.JWTSecret, cfg.RefreshTokenExpiry, rdb, cfg.Env)

	productService := service.NewProductService(repo)
	productHandler := handler.NewProductHandler(productService, cfg.JWTSecret)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{cfg.FrontendURL},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		}))
		userHandler.UserRoutes(r)
		productHandler.ProductRoutes(r)
	})

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	fmt.Println("server is running at port:", cfg.Port)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	stop()
	slog.Info("starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server is running out of time", "error", err)
	} else {
		slog.Info("server stopeed successfully")
	}

	slog.Info("server is already stopped")
}
