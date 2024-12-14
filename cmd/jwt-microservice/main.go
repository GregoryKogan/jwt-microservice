package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GregoryKogan/jwt-microservice/pkg/config"
	"github.com/GregoryKogan/jwt-microservice/pkg/logging"
	"github.com/spf13/viper"
)

func main() {
	config.Init()
	logging.Init()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	port := viper.GetString("server.port")
	slog.Info("Starting server", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("Failed to start server", slog.Any("error", err))
	}
}
