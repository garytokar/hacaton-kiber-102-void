package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bashkir-backend/internal/usecase"
)

type LeaderboardHandler struct {
	gamifyUC *usecase.GamificationUseCase
	lmsUC    *usecase.LMSUseCase
}

func NewLeaderboardHandler(gamifyUC *usecase.GamificationUseCase, lmsUC *usecase.LMSUseCase) *LeaderboardHandler {
	return &LeaderboardHandler{gamifyUC: gamifyUC, lmsUC: lmsUC}
}

func (h *LeaderboardHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	entries, err := h.gamifyUC.GetTopUsers(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}

func (h *LeaderboardHandler) GetSchools(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	schools, err := h.lmsUC.GetSchoolsLeaderboard(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(schools)
}
