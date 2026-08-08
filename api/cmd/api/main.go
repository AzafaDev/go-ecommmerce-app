package main

import (
	"context"
	"errors"
	"fmt"
	"go-ecommerce-app/internal/config"
	"go-ecommerce-app/internal/email"
	"go-ecommerce-app/internal/handler"
	"go-ecommerce-app/internal/payment"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

func main() {
	var wg sync.WaitGroup
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

	cartService := service.NewCartService(repo)
	cartHandler := handler.NewCartHandler(cartService, cfg.JWTSecret)

	midtransClient := payment.NewMidtransClient(cfg.MidtransServerKey, cfg.MidtransEnv)
	orderService := service.NewOrderService(repo, dbPool, midtransClient)
	orderHandler := handler.NewOrderHandler(orderService, cfg.JWTSecret)

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
		cartHandler.CartRoutes(r)
		orderHandler.OrderRoutes(r)
	})

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	fmt.Println("server is running at port:", cfg.Port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	wg.Add(1)
	go runOrderSweeper(ctx, &wg, orderService, cfg.OrderSweepInterval, cfg.OrderSweepThreshold)

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

	sweeperDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(sweeperDone)
	}()

	select {
	case <-sweeperDone:
		slog.Info("sweeper stopped cleanly")
	case <-time.After(5 * time.Second):
		slog.Warn("sweeper did not stop before timeout, exiting anyway")
	}

	slog.Info("server is already stopped")
}

// runOrderSweeper stops on ctx cancellation, same signal as the HTTP server.
func runOrderSweeper(ctx context.Context, wg *sync.WaitGroup, orderService *service.OrderService, interval, threshold time.Duration) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			swept, err := orderService.SweepExpiredOrders(ctx, threshold)
			if err != nil {
				slog.Error("order sweep", "error", err, "swept", swept)
				continue
			}
			if swept > 0 {
				slog.Info("order sweep", "swept", swept)
			}
		}
	}
}
