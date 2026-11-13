package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskDTO — карточка задания для клиента (Web, TMA или Telegram-чат).
type TaskDTO struct {
	FragmentTranslationID uuid.UUID `json:"fragment_translation_id"`
	FragmentID            uuid.UUID `json:"fragment_id"`
	TextBa                string    `json:"text_ba"`
	LangPair              string    `json:"lang_pair"`
	SuggestedTranslation  string    `json:"suggested_translation"`
	SourceTitle           string    `json:"source_title"`
	SourceCoordinates     string    `json:"source_coordinates"`
	AudioURL              *string   `json:"audio_url,omitempty"`
	AudioStartMS          *int      `json:"audio_start_ms,omitempty"`
	AudioEndMS            *int      `json:"audio_end_ms,omitempty"`
}

// DailyTasksResponse — 5 заданий на день.
type DailyTasksResponse struct {
	Date  string    `json:"date"`
	Count int       `json:"count"`
	Tasks []TaskDTO `json:"tasks"`
}

// SubmitAnswerReq — входной запрос на проверку/исправление карточки.
type SubmitAnswerReq struct {
	UserID                uuid.UUID        `json:"user_id"`
	FragmentTranslationID uuid.UUID        `json:"fragment_translation_id"`
	Action                SubmissionAction `json:"action"` // "confirmed" или "edited"
	SubmittedText         string           `json:"submitted_text"`
	TimeSpentSeconds      int              `json:"time_spent_seconds"`
}

// SubmitResultDTO — ответ после сабмита: начисленный опыт, стрик и проверка ловушки.
type SubmitResultDTO struct {
	XPEarned         int  `json:"xp_earned"`
	TotalXP          int  `json:"total_xp"`
	CurrentStreak    int  `json:"current_streak"`
	IsHoneypotCaught bool `json:"is_honeypot_caught"`
	IsDailyCompleted bool `json:"is_daily_completed"`
}

// AuthTelegramWebAppReq — запрос аутентификации из Telegram Mini App.
type AuthTelegramWebAppReq struct {
	InitData string `json:"init_data"`
}

// AuthTokenResponse — JWT токен доступа.
type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
