package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminHandler struct {
	consensusUC *usecase.ConsensusUseCase
}

func NewAdminHandler(consensusUC *usecase.ConsensusUseCase) *AdminHandler {
	return &AdminHandler{consensusUC: consensusUC}
}

// GetReviewQueue спорные случаи на экспертизу в УФИЦ РАН
func (h *AdminHandler) GetReviewQueue(w http.ResponseWriter, r *http.Request) {
	items, err := h.consensusUC.GetReviewQueue(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []domain.ExpertReviewItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h *AdminHandler) ApproveTranslation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = chi.URLParam(r, "review_id")
	}
	if idStr == "" {
		idStr = chi.URLParam(r, "translation_id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid translation id"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		ApprovedText string `json:"approved_text"`
		Translation  string `json:"translation"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}

	text := payload.ApprovedText
	if text == "" {
		text = payload.Translation
	}

	confirmed, err := h.consensusUC.ApproveTranslation(r.Context(), id, text)
	if err != nil {
		if errors.Is(err, domain.ErrFragmentNotFound) {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":               true,
		"message":               "Перевод утвержден экспертом УФИЦ РАН и занесен в эталонный корпус!",
		"confirmed_translation": confirmed,
	})
}

func (h *AdminHandler) RejectTranslation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = chi.URLParam(r, "review_id")
	}
	if idStr == "" {
		idStr = chi.URLParam(r, "translation_id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid translation id"}`, http.StatusBadRequest)
		return
	}

	err = h.consensusUC.RejectTranslation(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrFragmentNotFound) {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Перевод отклонен экспертом. Замечание зафиксировано.",
	})
}


func (h *AdminHandler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"pending_review": 4,
		"total_approved": 86,
		"accuracy_rate":  94.2,
		"active_experts": 3,
	})
}

func (h *AdminHandler) GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]any{
		{
			"id":          uuid.New().String(),
			"email":       "teacher@school1.ru",
			"username":    "teacher_zuhra",
			"full_name":   "Зухра Рашитовна",
			"avatar":      "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=100&auto=format&fit=crop&q=80",
			"role":        "teacher",
			"school_name": "Гимназия №1 им. Р. Гарипова",
			"created_at":  time.Now().Format(time.RFC3339),
		},
		{
			"id":          uuid.New().String(),
			"email":       "expert@ufic-ran.ru",
			"username":    "expert_ufic",
			"full_name":   "Азат Салаватович (УФИЦ РАН)",
			"avatar":      "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&auto=format&fit=crop&q=80",
			"role":        "expert",
			"school_name": "УФИЦ РАН",
			"created_at":  time.Now().Format(time.RFC3339),
		},
	})
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":   chi.URLParam(r, "id"),
		"role": role,
	})
}

// Dictionary Handler
func (h *AdminHandler) GetDictionary(w http.ResponseWriter, r *http.Request) {
	dict := []map[string]any{
		{
			"id":          1,
			"word":        "Ҡояш",
			"translation": "Солнце",
			"context":     "Ҡояш тау артына байыны",
			"is_favorite": true,
			"difficulty":  "easy",
			"created_at":  time.Now().Format(time.RFC3339),
		},
		{
			"id":          2,
			"word":        "Төйәк",
			"translation": "Край, обитель, родной дом",
			"context":     "Башҡортостан — беҙҙең тыуған төйәгебеҙ",
			"is_favorite": true,
			"difficulty":  "medium",
			"created_at":  time.Now().Format(time.RFC3339),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dict)
}

func (h *AdminHandler) AddDictionaryWord(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	_ = json.NewDecoder(r.Body).Decode(&payload)
	payload["id"] = time.Now().Unix()
	payload["is_favorite"] = true
	payload["created_at"] = time.Now().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *AdminHandler) ToggleFavoriteWord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":          chi.URLParam(r, "id"),
		"is_favorite": true,
	})
}

func (h *AdminHandler) DeleteDictionaryWord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// Quests
func (h *AdminHandler) GetDailyQuests(w http.ResponseWriter, r *http.Request) {
	quests := []map[string]any{
		{
			"id":              1,
			"code":            "DAILY_5",
			"name":            "Дневная норма Хранителя",
			"description":     "Проверьте 5 фрагментов текста за день",
			"target":          5,
			"reward_xp":       50,
			"reward_currency": 10,
			"progress":        3,
			"completed":       false,
			"claimed":         false,
		},
		{
			"id":              2,
			"code":            "EDIT_1",
			"name":            "Острый глаз",
			"description":     "Предложите хотя бы одно исправление перевода",
			"target":          1,
			"reward_xp":       25,
			"reward_currency": 5,
			"progress":        1,
			"completed":       true,
			"claimed":         false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quests)
}

func (h *AdminHandler) ClaimQuest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":          true,
		"xp_awarded":       50,
		"currency_awarded": 10,
		"message":          "Награда за квест успешно получена!",
	})
}

func (h *AdminHandler) GetWeeklyExpedition(w http.ResponseWriter, r *http.Request) {
	exp := map[string]any{
		"id":                  1,
		"title":               "Экспедиция по эпосу «Урал-батыр»",
		"description":         "Общая миссия школ республики: собрать 1000 подтверждённых строф башкирского эпоса",
		"target_count":        1000,
		"current_count":       342,
		"reward_xp":           250,
		"reward_currency":     50,
		"progress_percentage": 34,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(exp)
}
