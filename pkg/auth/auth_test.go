package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GregoryKogan/jwt-microservice/pkg/auth"
	"github.com/GregoryKogan/jwt-microservice/pkg/cache"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	mockCache cache.MockCache
	service   auth.AuthService
	handler   auth.AuthHandler
}

func (s *AuthTestSuite) SetupSuite() {
	viper.Set("secrets.jwt_key", "test_secret_key")
	viper.Set("auth.issuer", "test-jwt-microservice")
	viper.Set("auth.access_lifetime", 15*time.Minute)
	viper.Set("auth.refresh_lifetime", 720*time.Hour)
	viper.Set("auth.auto_logout", 24*time.Hour)
	viper.Set("logging.level", "debug")

	s.mockCache = cache.NewMockCache()
}

func (s *AuthTestSuite) TearDownSuite() {
	s.mockCache.Cleanup()
}

func (s *AuthTestSuite) SetupTest() {
	repo := auth.NewAuthRepo(s.mockCache.Cache())
	s.service = auth.NewAuthService(repo)
	s.handler = auth.NewAuthHandler(s.service)
}

func (s *AuthTestSuite) TearDownTest() {
	s.mockCache.Flush()
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) TestLoginFlow() {
	// Test login
	loginReq := map[string]interface{}{
		"user_id": uint(1),
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.Contains(resp, "access")
	s.Contains(resp, "refresh")

	// Test authentication with received token
	authReq := httptest.NewRequest(http.MethodGet, "/authenticate", nil)
	authReq.Header.Set("Authorization", "Bearer "+resp["access"])
	w = httptest.NewRecorder()

	s.handler.Authenticate(w, authReq)
	s.Equal(http.StatusOK, w.Code)

	var claims map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &claims)
	s.Equal(float64(1), claims["user_id"])
}

func (s *AuthTestSuite) TestRefreshFlow() {
	// First login to get tokens
	loginResp, _ := s.service.Login(1)

	// Test refresh
	refreshReq := map[string]string{
		"refresh": loginResp.Refresh,
	}
	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	s.handler.Refresh(w, req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.Contains(resp, "access")
	s.Contains(resp, "refresh")
	s.NotEqual(loginResp.Access, resp["access"])
}

func (s *AuthTestSuite) TestLogoutFlow() {
	// First login to get tokens
	loginResp, _ := s.service.Login(1)

	// Test logout
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Access)
	w := httptest.NewRecorder()

	s.handler.Logout(w, req)
	s.Equal(http.StatusOK, w.Code)

	// Verify token is invalidated
	authReq := httptest.NewRequest(http.MethodGet, "/authenticate", nil)
	authReq.Header.Set("Authorization", "Bearer "+loginResp.Access)
	w = httptest.NewRecorder()

	s.handler.Authenticate(w, authReq)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *AuthTestSuite) TestInvalidMethods() {
	tests := []struct {
		name     string
		path     string
		method   string
		handler  func(w http.ResponseWriter, r *http.Request)
		wantCode int
	}{
		{"Login with GET", "/login", http.MethodGet, s.handler.Login, http.StatusMethodNotAllowed},
		{"Refresh with GET", "/refresh", http.MethodGet, s.handler.Refresh, http.StatusMethodNotAllowed},
		{"Logout with GET", "/logout", http.MethodGet, s.handler.Logout, http.StatusMethodNotAllowed},
		{"Authenticate with POST", "/authenticate", http.MethodPost, s.handler.Authenticate, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func (s *AuthTestSuite) TestInvalidTokens() {
	// Test with invalid access token
	req := httptest.NewRequest(http.MethodGet, "/authenticate", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	s.handler.Authenticate(w, req)
	s.Equal(http.StatusBadRequest, w.Code)

	// Test with invalid refresh token
	refreshReq := map[string]string{
		"refresh": "invalid_token",
	}
	body, _ := json.Marshal(refreshReq)
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	w = httptest.NewRecorder()

	s.handler.Refresh(w, req)
	s.Equal(http.StatusInternalServerError, w.Code)
}
