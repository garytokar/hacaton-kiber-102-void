package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRole определяет уровень доступа пользователя в системе.
type UserRole string

const (
	RoleStudent UserRole = "student" // Ученик / обычный пользователь
	RoleTeacher UserRole = "teacher" // Преподаватель (доступ к классу)
	RoleExpert  UserRole = "expert"  // Эксперт-лингвист УФИЦ РАН
	RoleAdmin   UserRole = "admin"   // Администратор платформы
)

// School представляет учебное заведение (школу, лицей, гимназию).
type School struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
}

// Class представляет учебный класс внутри школы (например, "9 А").
type Class struct {
	ID          uuid.UUID `json:"id"`
	SchoolID    uuid.UUID `json:"school_id"`
	GradeNumber int       `json:"grade_number"` // 1-11
	Letter      string    `json:"letter"`       // "А", "Б", "В"
	CreatedAt   time.Time `json:"created_at"`
}

// User — единый аккаунт пользователя для Web, Mini App и Telegram-бота.
type User struct {
	ID             uuid.UUID  `json:"id"`
	TgID           *int64     `json:"tg_id,omitempty"`           // Telegram ID (для бота и TMA)
	TgUsername     *string    `json:"tg_username,omitempty"`     // @username
	Email          *string    `json:"email,omitempty"`           // Для учителей и экспертов
	HashedPassword *string    `json:"-"`                         // Хэш пароля (скрыт из json)
	FullName       string     `json:"full_name"`                 // Имя и фамилия
	Role           UserRole   `json:"role"`                      // Роль в системе
	SchoolID       *uuid.UUID `json:"school_id,omitempty"`       // Привязка к школе
	ClassID        *uuid.UUID `json:"class_id,omitempty"`        // Привязка к классу
	CurrentStreak  int        `json:"current_streak"`            // Текущая серия дней без перерыва
	MaxStreak      int        `json:"max_streak"`                // Максимальная серия дней
	LastActiveDate *time.Time `json:"last_active_date,omitempty"`// Дата последней активности
	TotalXP        int        `json:"total_xp"`                  // Общий накопленный опыт (рейтинг)
	CreatedAt      time.Time  `json:"created_at"`
}

// IsStreakActive проверяет, сохраняется ли стрик пользователя.
func (u *User) IsStreakActive() bool {
	if u.LastActiveDate == nil {
		return false
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	last := u.LastActiveDate.UTC().Truncate(24 * time.Hour)
	diff := today.Sub(last).Hours() / 24
	return diff <= 1
}

// UserRepository интерфейс хранилища пользователей
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByTgID(ctx context.Context, tgID int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateStreakAndXP(ctx context.Context, userID uuid.UUID, streak, maxStreak, addedXP int, activeDate time.Time) error
	GetTopUsers(ctx context.Context, limit int) ([]LeaderboardUserEntry, error)
}
