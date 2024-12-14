package config

import (
	"log/slog"

	"github.com/GregoryKogan/jwt-microservice/pkg/secrets"
	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Error("Config file not found", slog.Any("error", err))
			panic(err)
		} else {
			slog.Error("Error reading config file", slog.Any("error", err))
			panic(err)
		}
	}

	slog.Info("Config file successfully read")

	if err := secrets.LoadSecretsIntoViper(); err != nil {
		slog.Error("Failed to load secrets", slog.Any("error", err))
		panic(err)
	}
}
