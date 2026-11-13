package postgres

import (
	"context"
	"errors"
	"time"

	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentRepository struct {
	pool *pgxpool.Pool
}

func NewContentRepository(pool *pgxpool.Pool) *ContentRepository {
	return &ContentRepository{pool: pool}
}

// GetDailyFragmentsForUser выбирает 5 заданий для пользователя на текущий день
func (r *ContentRepository) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	query := `
		SELECT 
			ft.id, f.id, f.text_ba, ft.lang_pair,
			CASE 
				WHEN ft.has_honeypot = TRUE AND ft.honeypot_variant IS NOT NULL THEN ft.honeypot_variant
				ELSE ft.initial_auto_translation
			END AS suggested_translation,
			s.title,
			COALESCE(f.chapter_info, 'Раздел 1') AS coordinates,
			f.audio_url, f.audio_start_ms, f.audio_end_ms
		FROM fragment_translations ft
		JOIN fragments f ON ft.fragment_id = f.id
		JOIN sources s ON f.source_id = s.id
		WHERE ft.lang_pair = $1
		  AND ft.status = 'pending'
		  AND f.is_active = TRUE
		  AND NOT EXISTS (
			  SELECT 1 FROM user_submissions us 
			  WHERE us.fragment_translation_id = ft.id AND us.user_id = $2
		  )
		ORDER BY ft.confirmed_count ASC, RANDOM()
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, langPair, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.TaskDTO
	for rows.Next() {
		var t domain.TaskDTO
		err := rows.Scan(
			&t.FragmentTranslationID,
			&t.FragmentID,
			&t.TextBa,
			&t.LangPair,
			&t.SuggestedTranslation,
			&t.SourceTitle,
			&t.SourceCoordinates,
			&t.AudioURL,
			&t.AudioStartMS,
			&t.AudioEndMS,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *ContentRepository) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	query := `
		SELECT id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, confirmed_translation, consensus_score, confirmed_count, status, created_at, updated_at
		FROM fragment_translations
		WHERE id = $1
	`
	var ft domain.FragmentTranslation
	var statusStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ft.ID, &ft.FragmentID, &ft.LangPair, &ft.InitialAutoTranslation,
		&ft.HasHoneypot, &ft.HoneypotVariant, &ft.ConfirmedTranslation,
		&ft.ConsensusScore, &ft.ConfirmedCount, &statusStr, &ft.CreatedAt, &ft.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFragmentNotFound
		}
		return nil, err
	}
	ft.Status = domain.TranslationStatus(statusStr)
	return &ft, nil
}

func (r *ContentRepository) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	query := `
		UPDATE fragment_translations
		SET status = $1, confirmed_translation = $2, consensus_score = $3, confirmed_count = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.pool.Exec(ctx, query, string(status), confirmedText, score, count, time.Now(), id)
	return err
}

func (r *ContentRepository) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	query := `
		SELECT id, source_id, text_ba, audio_url, audio_start_ms, audio_end_ms, chapter_info, paragraph_idx, sentence_idx, is_active, created_at
		FROM fragments
		WHERE id = $1
	`
	var f domain.Fragment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.SourceID, &f.TextBa, &f.AudioURL, &f.AudioStartMS, &f.AudioEndMS,
		&f.ChapterInfo, &f.ParagraphIdx, &f.SentenceIdx, &f.IsActive, &f.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetConfirmedTranslations экспорт базы по разделу 5 ТЗ
func (r *ContentRepository) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	query := `
		SELECT id, fragment_id, text_ba, lang_pair, translation, source_coord, confirmed_count, status, created_at, updated_at
		FROM confirmed_translations
		WHERE lang_pair = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, langPair)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ConfirmedTranslation
	for rows.Next() {
		var item domain.ConfirmedTranslation
		err := rows.Scan(
			&item.ID, &item.FragmentID, &item.TextBa, &item.LangPair,
			&item.Translation, &item.SourceCoord, &item.ConfirmedCount,
			&item.Status, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
