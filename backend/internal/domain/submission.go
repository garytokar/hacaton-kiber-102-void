package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubmissionAction тип действия пользователя над предложенным автопереводом.
type SubmissionAction string

const (
	ActionConfirmed     SubmissionAction = "confirmed"      // Пользователь подтвердил корректность
	ActionEdited        SubmissionAction = "edited"         // Пользователь ввел свое исправление
	ActionReportedError SubmissionAction = "reported_error" // Пользователь пожаловался на нечитаемость
)

// UserSubmission — ответ конкретного пользователя на карточку задания.
type UserSubmission struct {
	ID                    uuid.UUID        `json:"id"`
	UserID                uuid.UUID        `json:"user_id"`
	FragmentTranslationID uuid.UUID        `json:"fragment_translation_id"`
	Action                SubmissionAction `json:"action"`
	SubmittedText         string           `json:"submitted_text"`
	TimeSpentSeconds      int              `json:"time_spent_seconds"`
	DetectedHoneypot      bool             `json:"detected_honeypot"` // Разоблачил ли проверочную ошибку
	XPAwarded             int              `json:"xp_awarded"`        // Начисленный опыт
	CreatedAt             time.Time        `json:"created_at"`
}

// ConfirmedTranslation — формат накопленной базы переводов по разделу 5 общего ТЗ.
type ConfirmedTranslation struct {
	ID             uuid.UUID `json:"id"`
	FragmentID     uuid.UUID `json:"fragment_id"`
	TextBa         string    `json:"text_ba"`
	LangPair       string    `json:"lang_pair"`
	Translation    string    `json:"translation"`
	SourceCoord    string    `json:"source"`
	ConfirmedCount int       `json:"confirmed"`
	Status         string    `json:"status"` // "confirmed" / "expert_review"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SubmissionGroup агрегированная группа одинаковых ответов
type SubmissionGroup struct {
	Text  string
	Count int
}

// SubmissionRepository интерфейс хранилища ответов
type SubmissionRepository interface {
	HasUserSubmitted(ctx context.Context, userID, fragmentTranslationID uuid.UUID) (bool, error)
	CreateSubmission(ctx context.Context, sub *UserSubmission) error
	CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error)
	CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error)
	GetSubmissionsGroupedByText(ctx context.Context, fragmentTranslationID uuid.UUID) ([]SubmissionGroup, int, error)
	SaveConfirmedTranslation(ctx context.Context, ct *ConfirmedTranslation) error
}
