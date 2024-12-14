package main

import (
	"log/slog"
	"net/http"

	"github.com/GregoryKogan/jwt-microservice/pkg/auth"
	"github.com/GregoryKogan/jwt-microservice/pkg/cache"
	"github.com/GregoryKogan/jwt-microservice/pkg/config"
	"github.com/GregoryKogan/jwt-microservice/pkg/logging"
	"github.com/GregoryKogan/jwt-microservice/pkg/ping"
	"github.com/spf13/viper"
)

func main() {
	config.Init()
	logging.Init()

	mux := http.NewServeMux()

	cache := cache.InitCacheConnection()

	// 1. Initialize repos
	authRepo := auth.NewAuthRepo(cache)

	// 2. Initialize services
	authService := auth.NewAuthService(authRepo)

	// 3. Initialize handlers
	authHandler := auth.NewAuthHandler(authService)
	pingHandler := ping.NewPingHandler()

	// Register routes
	mux.HandleFunc("/ping", pingHandler.Ping)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/refresh", authHandler.Refresh)
	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/authenticate", authHandler.Authenticate)

	port := viper.GetString("server.port")
	slog.Info("Starting server", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("Failed to start server", slog.Any("error", err))
	}
}
