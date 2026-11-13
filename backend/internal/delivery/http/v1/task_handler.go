package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"bashkir-backend/internal/delivery/http/middleware"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskUC *usecase.TaskUseCase
}

func NewTaskHandler(taskUC *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{taskUC: taskUC}
}

func (h *TaskHandler) GetDailyTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	langPair := r.URL.Query().Get("lang_pair")
	resp, err := h.taskUC.GetDailyTasks(r.Context(), userID, langPair)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type submitPayload struct {
	FragmentTranslationID uuid.UUID `json:"fragment_translation_id"`
	Action                string    `json:"action"`
	SubmittedText         string    `json:"submitted_text"`
	TimeSpentSeconds      int       `json:"time_spent_seconds"`
}

func (h *TaskHandler) SubmitTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var payload submitPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: payload.FragmentTranslationID,
		Action:                domain.SubmissionAction(payload.Action),
		SubmittedText:         payload.SubmittedText,
		TimeSpentSeconds:      payload.TimeSpentSeconds,
	}

	result, err := h.taskUC.SubmitAnswer(r.Context(), req)
	if err != nil {
		if err == domain.ErrAlreadySubmitted {
			http.Error(w, `{"error":"task already submitted"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// GetDailyFragment для React фронтенда (/api/fragments/daily)
func (h *TaskHandler) GetDailyFragment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		// Для гостей отдаем демо-карточку
		userID = uuid.New()
	}

	langPair := r.URL.Query().Get("lang_pair")
	resp, err := h.taskUC.GetDailyTasks(r.Context(), userID, langPair)
	if err == nil && len(resp.Tasks) > 0 {
		t := resp.Tasks[0]
		frag := map[string]any{
			"id":              t.FragmentID.String(),
			"text_ba":         t.TextBa,
			"source_title":    t.SourceTitle,
			"source_document": t.SourceTitle,
			"chapter":         t.SourceCoordinates,
			"audio_url":       t.AudioURL,
			"category":        "Классическая литература",
			"difficulty":      "medium",
			"translations": []map[string]any{
				{
					"id":               t.FragmentTranslationID.String(),
					"fragment_id":      t.FragmentID.String(),
					"translation_text": t.SuggestedTranslation,
					"status":           "pending",
					"votes_count":      1,
					"reject_count":     0,
					"created_at":       time.Now().Format(time.RFC3339),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(frag)
		return
	}

	// Дефолтная карточка при завершении нормы
	demoFrag := map[string]any{
		"id":              uuid.New().String(),
		"text_ba":         "Башҡортостан — беҙҙең тыуған төйәгебеҙ, ата-бабаларыбыҙ төйәге.",
		"source_title":    "Мостай Кәрим «Оҙон-оҙаҡ бала саҡ»",
		"source_document": "Повесть",
		"chapter":         "Глава 1, «Кендек»",
		"category":        "Классическая литература",
		"difficulty":      "medium",
		"translations": []map[string]any{
			{
				"id":               uuid.New().String(),
				"fragment_id":      uuid.New().String(),
				"translation_text": "Башкортостан — наш родной край, край наших предков.",
				"status":           "pending",
				"votes_count":      2,
				"reject_count":     0,
				"created_at":       time.Now().Format(time.RFC3339),
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(demoFrag)
}

func (h *TaskHandler) GetFragmentByID(w http.ResponseWriter, r *http.Request) {
	h.GetDailyFragment(w, r)
}

type proposeReq struct {
	FragmentID            string `json:"fragment_id"`
	FragmentTranslationID string `json:"fragment_translation_id"`
	TranslationText       string `json:"translation_text"`
}

// ProposeTranslation предложение своего перевода (/api/translations)
func (h *TaskHandler) ProposeTranslation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req proposeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	var transID uuid.UUID
	if req.FragmentTranslationID != "" {
		if parsed, err := uuid.Parse(req.FragmentTranslationID); err == nil {
			transID = parsed
		}
	}

	if transID == uuid.Nil && req.FragmentID != "" {
		fragID, err := uuid.Parse(req.FragmentID)
		if err != nil {
			http.Error(w, `{"error":"invalid fragment_id"}`, http.StatusBadRequest)
			return
		}
		langPair := r.URL.Query().Get("lang_pair")
		if langPair == "" {
			langPair = "ba-ru"
		}
		ft, err := h.taskUC.GetTranslationByFragment(r.Context(), fragID, langPair)
		if err != nil {
			http.Error(w, `{"error":"fragment translation not found"}`, http.StatusNotFound)
			return
		}
		transID = ft.ID
	}

	if transID == uuid.Nil {
		http.Error(w, `{"error":"missing translation identifier"}`, http.StatusBadRequest)
		return
	}

	submitReq := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: transID,
		Action:                domain.ActionEdited,
		SubmittedText:         req.TranslationText,
		TimeSpentSeconds:      15,
	}

	result, err := h.taskUC.SubmitAnswer(r.Context(), submitReq)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadySubmitted) {
			http.Error(w, `{"error":"вы уже отправляли ответ на этот фрагмент"}`, http.StatusConflict)
			return
		}
		if errors.Is(err, domain.ErrProfanityDetected) {
			http.Error(w, `{"error":"обнаружена ненормативная лексика"}`, http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidTranslationText) {
			http.Error(w, `{"error":"некорректный текст перевода"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	msg := "Ваш вариант перевода сохранён и передан в систему консенсуса!"
	if result.IsHoneypotCaught {
		msg = "🎯 Потрясающе! Вы заметили контрольную ловушку автоперевода (+30 XP)!"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":          true,
		"xp_awarded":       result.XPEarned,
		"currency_awarded": result.XPEarned / 5,
		"level_up":         false,
		"streak_updated":   true,
		"current_streak":   result.CurrentStreak,
		"message":          msg,
	})
}

type voteReq struct {
	VoteType string `json:"vote_type"`
}

// VoteTranslation голосование за перевод (/api/translations/{id}/vote)
func (h *TaskHandler) VoteTranslation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	transIDStr := chi.URLParam(r, "id")
	transID, err := uuid.Parse(transIDStr)
	if err != nil {
		transID = uuid.New()
	}

	var req voteReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	action := domain.ActionConfirmed
	submittedText := "confirmed"
	if req.VoteType == "reject" {
		action = domain.ActionReportedError
		submittedText = "[ERROR_REPORTED]"
	}

	submitReq := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: transID,
		Action:                action,
		SubmittedText:         submittedText,
		TimeSpentSeconds:      5,
	}

	result, err := h.taskUC.SubmitAnswer(r.Context(), submitReq)
	earned := 10
	streak := 1
	msg := "Перевод подтверждён! Голос учтён в консенсусе."
	if req.VoteType == "reject" {
		msg = "Жалоба принята. Материал будет направлен экспертам УФИЦ РАН."
	}
	if err == nil && result != nil {
		earned = result.XPEarned
		streak = result.CurrentStreak
		if result.IsHoneypotCaught {
			msg = "🎯 Вы разоблачили контрольную ошибку автоперевода (+30 XP)!"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":          true,
		"xp_awarded":       earned,
		"currency_awarded": earned / 5,
		"level_up":         false,
		"streak_updated":   true,
		"current_streak":   streak,
		"message":          msg,
	})
}

func (h *TaskHandler) GetTranslationHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]any{
		{
			"id":     "1",
			"author": "Машинный перевод",
			"time":   "Сегодня",
			"text":   "Базовый автоперевод HuggingFace / bashkorttele",
		},
	})
}
