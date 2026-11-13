package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SourceType тип первоисточника материала.
type SourceType string

const (
	SourceTextNews        SourceType = "text_news"
	SourceTextLiterature  SourceType = "text_literature"
	SourceAudioBroadcast  SourceType = "audio_broadcast"
	SourceAudioBook       SourceType = "audio_book"
	SourceHFDataset       SourceType = "hf_dataset"
)

// TranslationStatus статус консенсуса и согласованности перевода.
type TranslationStatus string

const (
	StatusPending      TranslationStatus = "pending"       // В процессе сбора ответов пользователей
	StatusConfirmed    TranslationStatus = "confirmed"     // Консенсус достигнут (3+ согласных голоса)
	StatusExpertReview TranslationStatus = "expert_review" // Спорный случай, передан экспертам УФИЦ РАН
	StatusRejected     TranslationStatus = "rejected"      // Отклонен экспертом (спам/некорректный)
)

// ExpertReviewItem спорный или эскалированный случай перевода для экспертизы УФИЦ РАН
type ExpertReviewItem struct {
	ID                     uuid.UUID `json:"id"`
	FragmentID             uuid.UUID `json:"fragment_id"`
	LangPair               string    `json:"lang_pair"`
	TextBa                 string    `json:"text_ba"`
	SourceTitle            string    `json:"source_title"`
	InitialAutoTranslation string    `json:"initial_auto_translation"`
	SuggestedTranslation   string    `json:"suggested_translation"`
	Status                 string    `json:"status"`
	ConfirmedCount         int       `json:"confirmed_count"`
	VotesCount             int       `json:"votes_count"`
	CreatedAt              time.Time `json:"created_at"`
}

// Source описывает оригинальный документ или аудиофайл.
type Source struct {
	ID            uuid.UUID      `json:"id"`
	Title         string         `json:"title"`
	SourceType    SourceType     `json:"source_type"`
	ExternalURL   *string        `json:"external_url,omitempty"`
	HFDatasetName *string        `json:"hf_dataset_name,omitempty"`
	Metadata      map[string]any `json:"metadata"`
	CreatedAt     time.Time      `json:"created_at"`
}

// Fragment — атомарная часть текста (предложение) с точными координатами источника.
type Fragment struct {
	ID            uuid.UUID `json:"id"`
	SourceID      uuid.UUID `json:"source_id"`
	TextBa        string    `json:"text_ba"`                 // Текст на башкирском языке
	AudioURL      *string   `json:"audio_url,omitempty"`     // Ссылка на аудиофрагмент
	AudioStartMS  *int      `json:"audio_start_ms,omitempty"`// Таймкод начала в миллисекундах
	AudioEndMS    *int      `json:"audio_end_ms,omitempty"`  // Таймкод окончания в миллисекундах
	ChapterInfo   *string   `json:"chapter_info,omitempty"`  // Произведение/глава/страница
	ParagraphIdx  *int      `json:"paragraph_idx,omitempty"` // Номер абзаца
	SentenceIdx   *int      `json:"sentence_idx,omitempty"`  // Номер предложения
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// FragmentTranslation — языковая пара и состояние проверки автоперевода.
type FragmentTranslation struct {
	ID                     uuid.UUID         `json:"id"`
	FragmentID             uuid.UUID         `json:"fragment_id"`
	LangPair               string            `json:"lang_pair"`                 // Например, "ba-ru", "ba-kk"
	InitialAutoTranslation string            `json:"initial_auto_translation"`  // Базовый машинный перевод
	HasHoneypot            bool              `json:"has_honeypot"`              // Является ли проверочной ловушкой
	HoneypotVariant        *string           `json:"honeypot_variant,omitempty"`// Искаженный автоперевод для античита
	ConfirmedTranslation   *string           `json:"confirmed_translation,omitempty"` // Финальный утвержденный вариант
	ConsensusScore         float64           `json:"consensus_score"`           // Метрика согласованности (0.0 - 1.0)
	ConfirmedCount         int               `json:"confirmed_count"`           // Число подтвердивших пользователей
	Status                 TranslationStatus `json:"status"`                    // pending, confirmed, expert_review, rejected
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
}

// ContentRepository интерфейс хранилища материалов и фрагментов
type ContentRepository interface {
	GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]TaskDTO, error)
	GetTranslationByID(ctx context.Context, id uuid.UUID) (*FragmentTranslation, error)
	GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*FragmentTranslation, error)
	UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status TranslationStatus, confirmedText string, score float64, count int) error
	GetFragmentByID(ctx context.Context, id uuid.UUID) (*Fragment, error)
	GetConfirmedTranslations(ctx context.Context, langPair string) ([]ConfirmedTranslation, error)
	GetReviewQueue(ctx context.Context) ([]ExpertReviewItem, error)
	ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*ConfirmedTranslation, error)
	RejectExpertReview(ctx context.Context, id uuid.UUID) error
}

