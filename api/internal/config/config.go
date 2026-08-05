package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiry           time.Duration
	RefreshTokenExpiry  time.Duration
	MidtransServerKey   string
	MidtransEnv         string
	MidtransMerchantID  string
	MidtransClientKey   string
	OrderSweepInterval  time.Duration
	OrderSweepThreshold time.Duration
	ResendAPIKey        string
	FrontendURL         string
	ResendFromAddress   string
	RedisAddr           string
	RedisPassword       string
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	orderSweepInterval, err := time.ParseDuration(getEnv("ORDER_SWEEP_INTERVAL", "5m"))
	if err != nil {
		return nil, fmt.Errorf("error parsing ORDER_SWEEP_INTERVAL")
	}

	orderSweepThreshold, err := time.ParseDuration(getEnv("ORDER_SWEEP_THRESHOLD", "30m"))
	if err != nil {
		return nil, fmt.Errorf("error parsing ORDER_SWEEP_THRESHOLD")
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("error parsing JWT_EXPIRY")
	}

	refreshTokenExpiry, err := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("error parsing REFRESH_TOKEN_EXPIRY")
	}

	cfg := Config{
		Port:                getEnv("PORT", "8080"),
		JWTExpiry:           jwtExpiry,
		RefreshTokenExpiry:  refreshTokenExpiry,
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		MidtransServerKey:   os.Getenv("MIDTRANS_SERVER_KEY"),
		MidtransEnv:         os.Getenv("MIDTRANS_ENV"),
		MidtransMerchantID:  os.Getenv("MIDTRANS_MERCHANT_ID"),
		MidtransClientKey:   os.Getenv("MIDTRANS_CLIENT_KEY"),
		ResendAPIKey:        os.Getenv("RESEND_API_KEY"),
		ResendFromAddress:   os.Getenv("RESEND_FROM_ADDRESS"),
		OrderSweepInterval:  orderSweepInterval,
		OrderSweepThreshold: orderSweepThreshold,
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", "myredispassword"),
	}

	var missing []string

	for index, value := range map[string]string{
		"DATABASE_URL":         cfg.DatabaseURL,
		"JWT_SECRET":           cfg.JWTSecret,
		"MIDTRANS_SERVER_KEY":  cfg.MidtransServerKey,
		"MIDTRANS_ENV":         cfg.MidtransEnv,
		"MIDTRANS_MERCHANT_ID": cfg.MidtransMerchantID,
		"MIDTRANS_CLIENT_KEY":  cfg.MidtransClientKey,
		"RESEND_API_KEY":       cfg.ResendAPIKey,
		"RESEND_FROM_ADDRESS":  cfg.ResendFromAddress,
	} {
		if value == "" {
			missing = append(missing, index)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing .env value: %v", missing)
	}

	return &cfg, nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
