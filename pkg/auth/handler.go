package auth

import (
	"encoding/json"
	"io"
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
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	type loginRequest struct {
		UserID uint `json:"user_id"`
	}

	var req loginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	tokenPair, err := h.service.Login(req.UserID)
	if err != nil {
		http.Error(w, "failed to login", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandlerImpl) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type refreshRequest struct {
		RefreshToken string `json:"refresh"`
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	tokenPair, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		http.Error(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandlerImpl) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "missing authorization header", http.StatusBadRequest)
		return
	}

	accessToken = accessToken[len("Bearer "):]

	if err := h.service.Logout(accessToken); err != nil {
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlerImpl) Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "missing authorization header", http.StatusBadRequest)
		return
	}

	accessToken = accessToken[len("Bearer "):]

	claims, err := h.service.Authenticate(accessToken)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(claims); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
