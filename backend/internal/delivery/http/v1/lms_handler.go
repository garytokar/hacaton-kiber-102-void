package v1

import (
	"encoding/json"
	"net/http"

	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LMSHandler struct {
	lmsUC *usecase.LMSUseCase
}

func NewLMSHandler(lmsUC *usecase.LMSUseCase) *LMSHandler {
	return &LMSHandler{lmsUC: lmsUC}
}

func (h *LMSHandler) GetClassStats(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "class_id")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid class_id"}`, http.StatusBadRequest)
		return
	}

	students, err := h.lmsUC.GetClassStudents(r.Context(), classID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(students)
}
