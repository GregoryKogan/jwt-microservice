package secrets

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func LoadSecretsIntoViper() error {
	files, err := os.ReadDir("/run/secrets")
	if err != nil {
		slog.Error("Failed to read secrets directory", slog.Any("error", err))
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		secretName := file.Name()
		value, err := readSecret(secretName)
		if err != nil {
			return err
		}
		viper.Set("secrets."+secretName, value)
		slog.Debug("Loaded secret", slog.String("name", secretName))
	}

	slog.Info("All secrets loaded into viper")
	return nil
}

func readSecret(secret string) (string, error) {
	buffer, err := os.ReadFile("/run/secrets/" + secret)
	if err != nil {
		slog.Error("failed to read secret", slog.Any("secret", secret), slog.Any("error", err))
		return "", err
	}
	return string(buffer), nil
}
