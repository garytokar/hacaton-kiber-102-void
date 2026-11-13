package postgres

import (
	"context"
	"time"

	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepository struct {
	pool *pgxpool.Pool
}

func NewSubmissionRepository(pool *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{pool: pool}
}

func (r *SubmissionRepository) HasUserSubmitted(ctx context.Context, userID, fragmentTranslationID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM user_submissions WHERE user_id = $1 AND fragment_translation_id = $2`
	var exists int
	err := r.pool.QueryRow(ctx, query, userID, fragmentTranslationID).Scan(&exists)
	if err != nil {
		return false, nil // no rows
	}
	return true, nil
}

func (r *SubmissionRepository) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	query := `
		INSERT INTO user_submissions (id, user_id, fragment_translation_id, action, submitted_text, time_spent_seconds, detected_honeypot, xp_awarded, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		sub.ID, sub.UserID, sub.FragmentTranslationID, string(sub.Action),
		sub.SubmittedText, sub.TimeSpentSeconds, sub.DetectedHoneypot,
		sub.XPAwarded, sub.CreatedAt,
	)
	return err
}

func (r *SubmissionRepository) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_submissions
		WHERE user_id = $1 AND created_at >= CURRENT_DATE
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_submissions
		WHERE user_id = $1
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) GetSubmissionsGroupedByText(ctx context.Context, fragmentTranslationID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	query := `
		SELECT LOWER(TRIM(submitted_text)) as clean_text, COUNT(*) as cnt
		FROM user_submissions
		WHERE fragment_translation_id = $1
		GROUP BY clean_text
		ORDER BY cnt DESC
	`
	rows, err := r.pool.Query(ctx, query, fragmentTranslationID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []domain.SubmissionGroup
	total := 0
	for rows.Next() {
		var g domain.SubmissionGroup
		err := rows.Scan(&g.Text, &g.Count)
		if err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
		total += g.Count
	}
	return groups, total, nil
}

func (r *SubmissionRepository) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	query := `
		INSERT INTO confirmed_translations (id, fragment_id, text_ba, lang_pair, translation, source_coord, confirmed_count, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE 
		SET translation = EXCLUDED.translation, confirmed_count = EXCLUDED.confirmed_count, updated_at = NOW()
	`
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	now := time.Now()
	ct.CreatedAt = now
	ct.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		ct.ID, ct.FragmentID, ct.TextBa, ct.LangPair, ct.Translation,
		ct.SourceCoord, ct.ConfirmedCount, ct.Status, ct.CreatedAt, ct.UpdatedAt,
	)
	return err
}
