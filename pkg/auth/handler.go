package auth

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Authenticate(w http.ResponseWriter, r *http.Request)
}

type AuthHandlerImpl struct {
	service AuthService
}

func NewAuthHandler(service AuthService) AuthHandler {
	return &AuthHandlerImpl{
		service: service,
	}
}

func (h *AuthHandlerImpl) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for login", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read login request body", "error", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	type loginRequest struct {
		UserID uint `json:"user_id"`
	}

	var req loginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("Failed to unmarshal login request", "error", err)
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	slog.Info("Processing login request", "user_id", req.UserID)
	tokenPair, err := h.service.Login(req.UserID)
	if err != nil {
		slog.Error("Failed to login user", "error", err, "user_id", req.UserID)
		http.Error(w, "failed to login", http.StatusInternalServerError)
		return
	}
	slog.Info("Login successful", "user_id", req.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandlerImpl) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for refresh", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type refreshRequest struct {
		RefreshToken string `json:"refresh"`
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode refresh request", "error", err)
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	slog.Info("Processing refresh token request")
	tokenPair, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh token", "error", err)
		http.Error(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}
	slog.Info("Token refresh successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandlerImpl) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("Invalid method for logout", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		slog.Warn("Missing authorization header in logout request")
		http.Error(w, "missing authorization header", http.StatusBadRequest)
		return
	}

	accessToken = accessToken[len("Bearer "):]
	slog.Info("Processing logout request")

	if err := h.service.Logout(accessToken); err != nil {
		slog.Error("Failed to logout", "error", err)
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}
	slog.Info("Logout successful")

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlerImpl) Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("Invalid method for authenticate", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		slog.Warn("Missing authorization header in authenticate request")
		http.Error(w, "missing authorization header", http.StatusBadRequest)
		return
	}

	accessToken = accessToken[len("Bearer "):]
	slog.Info("Processing authentication request")

	claims, err := h.service.Authenticate(accessToken)
	if err != nil {
		slog.Error("Authentication failed", "error", err)
		http.Error(w, "failed to authenticate", http.StatusBadRequest)
		return
	}
	slog.Info("Authentication successful", "user_id", claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(claims); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
