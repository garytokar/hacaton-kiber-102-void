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

// UserBadgeDTO представление достижения для API
type UserBadgeDTO struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IconURL     *string    `json:"icon_url,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	XPBonus     int        `json:"xp_bonus"`
	RewardXP    int        `json:"reward_xp"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
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
	Rank           int       `json:"rank"`
	SchoolID       uuid.UUID `json:"school_id"`
	SchoolName     string    `json:"school_name"`
	City           string    `json:"city"`
	TotalXP        int       `json:"total_xp"`
	ActiveStudents int       `json:"active_students"`
}

// LeaderboardClassEntry — рейтинг классов школ Республики Башкортостан.
type LeaderboardClassEntry struct {
	Rank           int       `json:"rank"`
	ClassID        uuid.UUID `json:"class_id"`
	SchoolID       uuid.UUID `json:"school_id"`
	SchoolName     string    `json:"school_name"`
	GradeNumber    int       `json:"grade_number"`
	Letter         string    `json:"letter"`
	ClassName      string    `json:"class_name"`
	TotalXP        int       `json:"total_xp"`
	ActiveStudents int       `json:"active_students"`
}

// LeaderboardVocabularyEntry — рейтинг по словарному запасу (раздел 4.5 ТЗ).
type LeaderboardVocabularyEntry struct {
	Rank           int       `json:"rank"`
	UserID         uuid.UUID `json:"user_id"`
	FullName       string    `json:"full_name"`
	TgUsername     *string   `json:"tg_username,omitempty"`
	SchoolName     *string   `json:"school_name,omitempty"`
	WordsCount     int       `json:"words_count"`
	VocabularySize int       `json:"vocabulary_size"`
	TotalXP        int       `json:"total_xp"`
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
	GetClassesLeaderboard(ctx context.Context, limit int) ([]LeaderboardClassEntry, error)
	GetClassStudents(ctx context.Context, classID uuid.UUID) ([]ClassStudentStats, error)
	CreateClass(ctx context.Context, c *Class) error
	GetClassByID(ctx context.Context, id uuid.UUID) (*Class, error)
	GetClassByInviteCode(ctx context.Context, code string) (*Class, error)
	UpdateClassInviteCode(ctx context.Context, classID uuid.UUID, code string) error
	AssignUserToClass(ctx context.Context, userID, classID, schoolID uuid.UUID) error
}

