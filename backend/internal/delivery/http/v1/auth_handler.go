package v1

import (
	"encoding/json"
	"net/http"

	"bashkir-backend/internal/delivery/http/middleware"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type AuthHandler struct {
	authUC   *usecase.AuthUseCase
	userRepo domain.UserRepository
}

func NewAuthHandler(authUC *usecase.AuthUseCase, userRepo domain.UserRepository) *AuthHandler {
	return &AuthHandler{authUC: authUC, userRepo: userRepo}
}

type registerReq struct {
	FullName string     `json:"full_name"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Role     string     `json:"role"`
	SchoolID *uuid.UUID `json:"school_id,omitempty"`
	ClassID  *uuid.UUID `json:"class_id,omitempty"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	role := domain.UserRole(req.Role)
	if role == "" {
		role = domain.RoleStudent
	}

	user, err := h.authUC.RegisterEmail(r.Context(), req.FullName, req.Email, req.Password, role, req.SchoolID, req.ClassID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	tokenResp, user, err := h.authUC.LoginEmail(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"token": tokenResp,
		"user":  user,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) TelegramWebAppAuth(w http.ResponseWriter, r *http.Request) {
	var req domain.AuthTelegramWebAppReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	tokenResp, user, err := h.authUC.AuthTelegramWebApp(r.Context(), req.InitData)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"token": tokenResp,
		"user":  user,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
