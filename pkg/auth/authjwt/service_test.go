package authjwt_test

import (
	"testing"

	"github.com/GregoryKogan/jwt-microservice/pkg/auth/authjwt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupTestJWTService(_ *testing.T) authjwt.JWTService {
	viper.Set("auth.access_lifetime", "15m")
	viper.Set("auth.refresh_lifetime", "720h")
	viper.Set("auth.auto_logout", "24h")
	viper.Set("secrets.jwt_key", "test_key")

	return authjwt.NewJWTService()
}

func TestNewAccessToken(t *testing.T) {
	service := setupTestJWTService(t)

	tokenString, err := service.NewAccessToken(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestNewRefreshToken(t *testing.T) {
	service := setupTestJWTService(t)

	tokenString, err := service.NewRefreshToken(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestParseToken(t *testing.T) {
	service := setupTestJWTService(t)

	accessTokenString, err := service.NewAccessToken(1)
	assert.NoError(t, err)

	refreshTokenString, err := service.NewRefreshToken(1)
	assert.NoError(t, err)

	accessTokenClaims, err := service.ParseToken(accessTokenString)
	assert.NoError(t, err)

	refreshTokenClaims, err := service.ParseToken(refreshTokenString)
	assert.NoError(t, err)

	assert.Equal(t, uint(1), accessTokenClaims.UserID)
	assert.Equal(t, uint(1), refreshTokenClaims.UserID)

	assert.Equal(t, "access", accessTokenClaims.Type)
	assert.Equal(t, "refresh", refreshTokenClaims.Type)

	assert.NotEmpty(t, accessTokenClaims.UID)
	assert.NotEmpty(t, refreshTokenClaims.UID)
	assert.NotEqual(t, accessTokenClaims.UID, refreshTokenClaims.UID)
}

func TestParseToken_Invalid(t *testing.T) {
	service := setupTestJWTService(t)

	_, err := service.ParseToken("invalid token")
	assert.Error(t, err)
}
