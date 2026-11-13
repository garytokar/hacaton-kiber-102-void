package postgres

import (
	"context"
	"errors"
	"fmt"
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

func (r *ContentRepository) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	query := `
		SELECT 
			ft.id, ft.fragment_id, ft.lang_pair, f.text_ba, s.title,
			ft.initial_auto_translation,
			COALESCE(ft.confirmed_translation, ft.initial_auto_translation) as suggested_translation,
			ft.status, ft.confirmed_count,
			COALESCE((SELECT COUNT(*) FROM user_submissions us WHERE us.fragment_translation_id = ft.id), 0)::INT as votes_count,
			ft.created_at
		FROM fragment_translations ft
		JOIN fragments f ON ft.fragment_id = f.id
		JOIN sources s ON f.source_id = s.id
		WHERE ft.status = 'expert_review'
		ORDER BY ft.updated_at DESC, ft.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ExpertReviewItem
	for rows.Next() {
		var item domain.ExpertReviewItem
		err := rows.Scan(
			&item.ID, &item.FragmentID, &item.LangPair, &item.TextBa, &item.SourceTitle,
			&item.InitialAutoTranslation, &item.SuggestedTranslation,
			&item.Status, &item.ConfirmedCount, &item.VotesCount, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *ContentRepository) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	ft, err := r.GetTranslationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	frag, err := r.GetFragmentByID(ctx, ft.FragmentID)
	if err != nil {
		return nil, err
	}

	finalText := approvedText
	if finalText == "" {
		if ft.ConfirmedTranslation != nil && *ft.ConfirmedTranslation != "" {
			finalText = *ft.ConfirmedTranslation
		} else {
			finalText = ft.InitialAutoTranslation
		}
	}

	newCount := ft.ConfirmedCount + 1
	updateQuery := `
		UPDATE fragment_translations
		SET status = 'confirmed', confirmed_translation = $1, consensus_score = 1.0, confirmed_count = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err = r.pool.Exec(ctx, updateQuery, finalText, newCount, id)
	if err != nil {
		return nil, err
	}

	coord := "Источник материалов"
	if frag.ChapterInfo != nil && *frag.ChapterInfo != "" {
		coord = *frag.ChapterInfo
	}
	if frag.SentenceIdx != nil {
		coord = fmt.Sprintf("%s, предл. %d", coord, *frag.SentenceIdx)
	}

	ct := &domain.ConfirmedTranslation{
		ID:             uuid.New(),
		FragmentID:     ft.FragmentID,
		TextBa:         frag.TextBa,
		LangPair:       ft.LangPair,
		Translation:    finalText,
		SourceCoord:    coord,
		ConfirmedCount: newCount,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	insertQuery := `
		INSERT INTO confirmed_translations (id, fragment_id, text_ba, lang_pair, translation, source_coord, confirmed_count, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, _ = r.pool.Exec(ctx, insertQuery,
		ct.ID, ct.FragmentID, ct.TextBa, ct.LangPair, ct.Translation,
		ct.SourceCoord, ct.ConfirmedCount, ct.Status, ct.CreatedAt, ct.UpdatedAt,
	)

	return ct, nil
}

func (r *ContentRepository) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE fragment_translations
		SET status = 'rejected', updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFragmentNotFound
	}
	return nil
}



func (r *ContentRepository) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	query := `
		SELECT id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, 
		       confirmed_translation, consensus_score, confirmed_count, status, created_at, updated_at
		FROM fragment_translations
		WHERE fragment_id = $1 AND lang_pair = $2
		LIMIT 1
	`
	var ft domain.FragmentTranslation
	var statusStr string
	err := r.pool.QueryRow(ctx, query, fragmentID, langPair).Scan(
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
