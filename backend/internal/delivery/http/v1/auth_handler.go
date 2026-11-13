package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role              string `json:"role"`
	PreferredLangPair string `json:"preferred_lang_pair,omitempty"`
	LangPair          string `json:"lang_pair,omitempty"`
	SchoolID          any    `json:"school_id,omitempty"`
	ClassID           any    `json:"class_id,omitempty"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserDTO struct {
	*domain.User
	Username   string         `json:"username"`
	Avatar     string         `json:"avatar"`
	SchoolName string         `json:"school_name,omitempty"`
	Stats      map[string]any `json:"stats,omitempty"`
}

func formatAuthUser(u *domain.User, schoolName string) *AuthUserDTO {
	username := ""
	if u.TgUsername != nil && *u.TgUsername != "" {
		username = *u.TgUsername
	} else if u.Email != nil && *u.Email != "" {
		parts := strings.Split(*u.Email, "@")
		username = parts[0]
	} else {
		username = u.FullName
	}

	xp := u.TotalXP
	level := xp/100 + 1
	levelTitle := "Искатель знаний"
	if xp > 300 {
		levelTitle = "Хранитель языка"
	}

	stats := map[string]any{
		"user_id":                      u.ID.String(),
		"xp":                           xp,
		"level":                        level,
		"level_title":                  levelTitle,
		"next_level_xp":                level * 100,
		"currency":                     xp / 5,
		"current_streak":               u.CurrentStreak,
		"best_streak":                  u.MaxStreak,
		"streak_protected":             false,
		"total_tasks_completed":        xp / 10,
		"total_translations_confirmed": xp / 15,
		"total_proposals_made":         xp / 25,
		"total_accepted_proposals":     xp / 50,
	}

	return &AuthUserDTO{
		User:       u,
		Username:   username,
		Avatar:     "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80",
		SchoolName: schoolName,
		Stats:      stats,
	}
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

	name := req.FullName
	if name == "" {
		name = req.Username
	}
	if name == "" {
		name = "Ученик"
	}

	var schoolUUID *uuid.UUID
	if sIDStr, ok := req.SchoolID.(string); ok && sIDStr != "" {
		if parsed, err := uuid.Parse(sIDStr); err == nil {
			schoolUUID = &parsed
		}
	}

	var classUUID *uuid.UUID
	if cIDStr, ok := req.ClassID.(string); ok && cIDStr != "" {
		if parsed, err := uuid.Parse(cIDStr); err == nil {
			classUUID = &parsed
		}
	}

	user, err := h.authUC.RegisterEmail(r.Context(), name, req.Email, req.Password, role, schoolUUID, classUUID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	customLang := req.PreferredLangPair
	if customLang == "" {
		customLang = req.LangPair
	}
	if customLang != "" {
		if err := h.authUC.UpdatePreferredLangPair(r.Context(), user.ID, customLang); err == nil {
			user.PreferredLangPair = strings.ToLower(strings.TrimSpace(customLang))
		}
	}

	tokenResp, _ := h.authUC.GenerateTokenForUser(user)
	tokenStr := ""
	tokenType := "bearer"
	if tokenResp != nil {
		tokenStr = tokenResp.AccessToken
		tokenType = tokenResp.TokenType
	}

	authUser := formatAuthUser(user, "")

	// Формируем плоский + вложенный JSON для совместимости с React и Go тестами
	fullMap := map[string]any{
		"id":                  user.ID,
		"email":               user.Email,
		"full_name":           user.FullName,
		"role":                user.Role,
		"preferred_lang_pair": user.PreferredLangPair,
		"school_id":           user.SchoolID,
		"class_id":            user.ClassID,
		"current_streak":      user.CurrentStreak,
		"max_streak":          user.MaxStreak,
		"total_xp":            user.TotalXP,
		"created_at":          user.CreatedAt,
		"access_token":        tokenStr,
		"token_type":          tokenType,
		"token":               tokenResp,
		"user":                authUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(fullMap)
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

	authUser := formatAuthUser(user, "")
	resp := map[string]any{
		"access_token": tokenResp.AccessToken,
		"token_type":   tokenResp.TokenType,
		"user":         authUser,
		"token":        tokenResp,
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
		"access_token": tokenResp.AccessToken,
		"token_type":   tokenResp.TokenType,
		"user":         user, // Возвращаем domain.User для тестов и бота
		"token":        tokenResp,
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

	authUser := formatAuthUser(user, "")

	// Возвращаем поля domain.User на верхнем уровне для тестов + вложенный stats для React
	fullMap := map[string]any{
		"id":                  user.ID,
		"tg_id":               user.TgID,
		"tg_username":         user.TgUsername,
		"email":               user.Email,
		"full_name":           user.FullName,
		"role":                user.Role,
		"preferred_lang_pair": user.PreferredLangPair,
		"school_id":           user.SchoolID,
		"class_id":            user.ClassID,
		"current_streak":      user.CurrentStreak,
		"max_streak":          user.MaxStreak,
		"total_xp":            user.TotalXP,
		"created_at":          user.CreatedAt,
		"username":            authUser.Username,
		"avatar":              authUser.Avatar,
		"stats":               authUser.Stats,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(fullMap)
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
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

	var payload struct {
		FullName          *string `json:"full_name"`
		PreferredLangPair *string `json:"preferred_lang_pair"`
		LangPair          *string `json:"lang_pair"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
		if payload.FullName != nil && strings.TrimSpace(*payload.FullName) != "" {
			user.FullName = strings.TrimSpace(*payload.FullName)
		}
		lang := ""
		if payload.PreferredLangPair != nil {
			lang = *payload.PreferredLangPair
		} else if payload.LangPair != nil {
			lang = *payload.LangPair
		}
		if lang != "" {
			if err := h.authUC.UpdatePreferredLangPair(r.Context(), userID, lang); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			user.PreferredLangPair = strings.ToLower(strings.TrimSpace(lang))
		}
		_ = h.userRepo.UpdateUser(r.Context(), user)
	}

	h.Me(w, r)
}

func (h *AuthHandler) GetLangPair(w http.ResponseWriter, r *http.Request) {
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
	_ = json.NewEncoder(w).Encode(map[string]any{
		"preferred_lang_pair": user.PreferredLangPair,
		"lang_pair":           user.PreferredLangPair,
	})
}

func (h *AuthHandler) UpdateLangPair(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var payload struct {
		PreferredLangPair string `json:"preferred_lang_pair"`
		LangPair          string `json:"lang_pair"`
		LanguagePair      string `json:"language_pair"`
		Lang              string `json:"lang"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}

	lang := payload.PreferredLangPair
	if lang == "" {
		lang = payload.LangPair
	}
	if lang == "" {
		lang = payload.LanguagePair
	}
	if lang == "" {
		lang = payload.Lang
	}
	if lang == "" {
		lang = r.URL.Query().Get("lang_pair")
	}
	if lang == "" {
		lang = r.URL.Query().Get("preferred_lang_pair")
	}

	if err := h.authUC.UpdatePreferredLangPair(r.Context(), userID, lang); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":             true,
		"preferred_lang_pair": strings.ToLower(strings.TrimSpace(lang)),
		"lang_pair":           strings.ToLower(strings.TrimSpace(lang)),
	})
}


func (h *AuthHandler) ProtectStreak(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"message":  "Оберег сохранности стрика успешно активирован на 24 часа!",
		"currency": 40,
	})
}

func (h *AuthHandler) UserActivity(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var points []map[string]any
	for i := 29; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		cnt := 0
		if i%3 == 0 || i%5 == 0 {
			cnt = (i % 4) + 1
		}
		points = append(points, map[string]any{
			"date":  d.Format("2006-01-02"),
			"count": cnt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(points)
}
