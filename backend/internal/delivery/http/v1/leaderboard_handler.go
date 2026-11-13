package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bashkir-backend/internal/delivery/http/middleware"
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

	period := r.URL.Query().Get("period")
	if period == "" {
		period = r.URL.Query().Get("timeframe")
	}
	if period == "" {
		period = r.URL.Query().Get("range")
	}

	entries, err := h.gamifyUC.GetTopUsersByPeriod(r.Context(), period, limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	formatted := make([]map[string]any, 0)
	for _, entry := range entries {
		uname := entry.FullName
		if entry.TgUsername != nil && *entry.TgUsername != "" {
			uname = "@" + *entry.TgUsername
		}
		sName := "Республика Башкортостан"
		if entry.SchoolName != nil && *entry.SchoolName != "" {
			sName = *entry.SchoolName
		}
		level := entry.TotalXP/100 + 1
		title := "Искатель"
		if entry.TotalXP > 300 {
			title = "Хранитель языка"
		}

		formatted = append(formatted, map[string]any{
			"rank":            entry.Rank,
			"user_id":         entry.UserID.String(),
			"username":        uname,
			"full_name":       entry.FullName,
			"avatar":          "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80",
			"level":           level,
			"level_title":     title,
			"xp":              entry.TotalXP,
			"total_xp":        entry.TotalXP,
			"streak":          entry.CurrentStreak,
			"current_streak":  entry.CurrentStreak,
			"confirmed_count": entry.TotalXP / 15,
			"school_name":     sName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(formatted)
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

	formatted := make([]map[string]any, 0)
	for _, s := range schools {
		formatted = append(formatted, map[string]any{
			"rank":            s.Rank,
			"school_id":       s.SchoolID.String(),
			"school_name":     s.SchoolName,
			"name":            s.SchoolName,
			"city":            s.City,
			"active_students": s.ActiveStudents,
			"students_count":  s.ActiveStudents,
			"total_xp":        s.TotalXP,
			"total_confirmed": s.TotalXP / 15,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(formatted)
}

func (h *LeaderboardHandler) GetClasses(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	classes, err := h.lmsUC.GetClassesLeaderboard(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	formatted := make([]map[string]any, 0)
	for _, c := range classes {
		formatted = append(formatted, map[string]any{
			"rank":            c.Rank,
			"class_id":        c.ClassID.String(),
			"school_id":       c.SchoolID.String(),
			"school_name":     c.SchoolName,
			"grade_number":    c.GradeNumber,
			"letter":          c.Letter,
			"class_name":      c.ClassName,
			"name":            c.ClassName + " (" + c.SchoolName + ")",
			"total_xp":        c.TotalXP,
			"active_students": c.ActiveStudents,
			"students_count":  c.ActiveStudents,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(formatted)
}

func (h *LeaderboardHandler) GetVocabulary(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	entries, err := h.gamifyUC.GetVocabularyLeaderboard(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	formatted := make([]map[string]any, 0)
	for _, entry := range entries {
		uname := entry.FullName
		if entry.TgUsername != nil && *entry.TgUsername != "" {
			uname = "@" + *entry.TgUsername
		}
		sName := "Республика Башкортостан"
		if entry.SchoolName != nil && *entry.SchoolName != "" {
			sName = *entry.SchoolName
		}

		formatted = append(formatted, map[string]any{
			"rank":            entry.Rank,
			"user_id":         entry.UserID.String(),
			"username":        uname,
			"full_name":       entry.FullName,
			"words_count":     entry.WordsCount,
			"vocabulary_size": entry.VocabularySize,
			"total_xp":        entry.TotalXP,
			"school_name":     sName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(formatted)
}

func (h *LeaderboardHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	teams := []map[string]any{
		{"rank": 1, "team_id": 1, "name": "Бөркөттәр (Орлы Урала)", "emblem": "🦅", "members_count": 14, "total_xp": 2450, "total_confirmed": 120},
		{"rank": 2, "team_id": 2, "name": "Толпар (Крылатые кони)", "emblem": "🐎", "members_count": 11, "total_xp": 1890, "total_confirmed": 94},
		{"rank": 3, "team_id": 3, "name": "Ағиҙел дуҫтары", "emblem": "🌊", "members_count": 9, "total_xp": 1240, "total_confirmed": 65},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(teams)
}

func (h *LeaderboardHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	achievements, err := h.gamifyUC.GetUserAchievements(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(achievements)
}

