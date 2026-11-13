package v1

import (
	"encoding/json"
	"net/http"

	"bashkir-backend/internal/delivery/http/middleware"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

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
	if langPair == "" {
		langPair = "ba-ru"
	}

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
