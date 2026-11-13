package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Badge — достижение, выдаваемое за активность.
type Badge struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     *string   `json:"icon_url,omitempty"`
	XPBonus     int       `json:"xp_bonus"`
}

// UserBadge — связь пользователя и полученного им бейджа.
type UserBadge struct {
	UserID    uuid.UUID `json:"user_id"`
	BadgeID   uuid.UUID `json:"badge_id"`
	AwardedAt time.Time `json:"awarded_at"`
}

// LeaderboardUserEntry — запись в индивидуальном рейтинге.
type LeaderboardUserEntry struct {
	Rank          int       `json:"rank"`
	UserID        uuid.UUID `json:"user_id"`
	FullName      string    `json:"full_name"`
	TgUsername    *string   `json:"tg_username,omitempty"`
	SchoolName    *string   `json:"school_name,omitempty"`
	TotalXP       int       `json:"total_xp"`
	CurrentStreak int       `json:"current_streak"`
}

// LeaderboardSchoolEntry — запись в рейтинге школ.
type LeaderboardSchoolEntry struct {
	Rank            int       `json:"rank"`
	SchoolID        uuid.UUID `json:"school_id"`
	SchoolName      string    `json:"school_name"`
	City            string    `json:"city"`
	TotalXP         int       `json:"total_xp"`
	ActiveStudents  int       `json:"active_students"`
}

// ClassStudentStats статистика ученика в сводке для учителя
type ClassStudentStats struct {
	UserID        uuid.UUID `json:"user_id"`
	FullName      string    `json:"full_name"`
	TgUsername    *string   `json:"tg_username,omitempty"`
	TotalXP       int       `json:"total_xp"`
	CurrentStreak int       `json:"current_streak"`
	SolvedToday   int       `json:"solved_today"`
}

// LMSRepository интерфейс хранилища школьной и классной аналитики
type LMSRepository interface {
	GetSchoolsLeaderboard(ctx context.Context, limit int) ([]LeaderboardSchoolEntry, error)
	GetClassStudents(ctx context.Context, classID uuid.UUID) ([]ClassStudentStats, error)
}
