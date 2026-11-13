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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, tg_id, tg_username, email, hashed_password, full_name, role, preferred_lang_pair, school_id, class_id, current_streak, max_streak, total_xp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.PreferredLangPair == "" {
		u.PreferredLangPair = "ba-ru"
	}

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.TgID, u.TgUsername, u.Email, u.HashedPassword, u.FullName,
		string(u.Role), u.PreferredLangPair, u.SchoolID, u.ClassID, u.CurrentStreak, u.MaxStreak,
		u.TotalXP, u.CreatedAt,
	)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, COALESCE(preferred_lang_pair, 'ba-ru'), school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.PreferredLangPair, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
		&u.LastActiveDate, &u.TotalXP, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepository) FindByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	query := `
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, COALESCE(preferred_lang_pair, 'ba-ru'), school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE tg_id = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, tgID).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.PreferredLangPair, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
		&u.LastActiveDate, &u.TotalXP, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, COALESCE(preferred_lang_pair, 'ba-ru'), school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.PreferredLangPair, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
		&u.LastActiveDate, &u.TotalXP, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepository) UpdateStreakAndXP(ctx context.Context, userID uuid.UUID, streak, maxStreak, addedXP int, activeDate time.Time) error {
	query := `
		UPDATE users
		SET current_streak = $1, max_streak = $2, total_xp = total_xp + $3, last_active_date = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, streak, maxStreak, addedXP, activeDate, userID)
	return err
}

func (r *UserRepository) UpdatePreferredLangPair(ctx context.Context, userID uuid.UUID, langPair string) error {
	query := `UPDATE users SET preferred_lang_pair = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, langPair, userID)
	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users 
		SET full_name = $1, preferred_lang_pair = $2, school_id = $3, class_id = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, u.FullName, u.PreferredLangPair, u.SchoolID, u.ClassID, u.ID)
	return err
}

func (r *UserRepository) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	return r.GetTopUsersByPeriod(ctx, "all", limit)
}

func (r *UserRepository) GetTopUsersByPeriod(ctx context.Context, period string, limit int) ([]domain.LeaderboardUserEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if period == "" || period == "all" || period == "all-time" {
		query := `
			SELECT u.id, u.full_name, u.tg_username, s.name, u.total_xp, u.current_streak
			FROM users u
			LEFT JOIN schools s ON u.school_id = s.id
			ORDER BY u.total_xp DESC
			LIMIT $1
		`
		rows, err := r.pool.Query(ctx, query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result []domain.LeaderboardUserEntry
		rank := 1
		for rows.Next() {
			var entry domain.LeaderboardUserEntry
			entry.Rank = rank
			err := rows.Scan(
				&entry.UserID, &entry.FullName, &entry.TgUsername, &entry.SchoolName,
				&entry.TotalXP, &entry.CurrentStreak,
			)
			if err != nil {
				return nil, err
			}
			result = append(result, entry)
			rank++
		}
		return result, nil
	}

	var intervalCond string
	switch period {
	case "day":
		intervalCond = "us.created_at >= CURRENT_DATE"
	case "week":
		intervalCond = "us.created_at >= NOW() - INTERVAL '7 days'"
	case "month":
		intervalCond = "us.created_at >= NOW() - INTERVAL '30 days'"
	default:
		return r.GetTopUsers(ctx, limit)
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.full_name, u.tg_username, s.name,
		       COALESCE(SUM(us.xp_awarded), 0)::INT as period_xp,
		       u.current_streak
		FROM users u
		LEFT JOIN schools s ON u.school_id = s.id
		JOIN user_submissions us ON u.id = us.user_id AND %s
		GROUP BY u.id, u.full_name, u.tg_username, s.name, u.current_streak
		ORDER BY period_xp DESC
		LIMIT $1
	`, intervalCond)

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LeaderboardUserEntry
	rank := 1
	for rows.Next() {
		var entry domain.LeaderboardUserEntry
		entry.Rank = rank
		err := rows.Scan(
			&entry.UserID, &entry.FullName, &entry.TgUsername, &entry.SchoolName,
			&entry.TotalXP, &entry.CurrentStreak,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
		rank++
	}
	return result, nil
}

func (r *UserRepository) GetVocabularyLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardVocabularyEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT u.id, u.full_name, u.tg_username, s.name, u.total_xp,
		       COUNT(DISTINCT us.fragment_translation_id)::INT as words_count
		FROM users u
		LEFT JOIN schools s ON u.school_id = s.id
		LEFT JOIN user_submissions us ON u.id = us.user_id
		GROUP BY u.id, u.full_name, u.tg_username, s.name, u.total_xp
		ORDER BY words_count DESC, u.total_xp DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LeaderboardVocabularyEntry
	rank := 1
	for rows.Next() {
		var entry domain.LeaderboardVocabularyEntry
		entry.Rank = rank
		err := rows.Scan(
			&entry.UserID, &entry.FullName, &entry.TgUsername, &entry.SchoolName,
			&entry.TotalXP, &entry.WordsCount,
		)
		if err != nil {
			return nil, err
		}
		entry.VocabularySize = entry.WordsCount
		result = append(result, entry)
		rank++
	}
	return result, nil
}

// Badges
func (r *UserRepository) GetAllBadges(ctx context.Context) ([]domain.Badge, error) {
	query := `
		SELECT id, code, name, description, icon_url, xp_bonus
		FROM badges
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []domain.Badge
	for rows.Next() {
		var b domain.Badge
		err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Description, &b.IconURL, &b.XPBonus)
		if err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}
	return badges, nil
}

func (r *UserRepository) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	query := `
		SELECT user_id, badge_id, awarded_at
		FROM user_badges
		WHERE user_id = $1
		ORDER BY awarded_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ubs []domain.UserBadge
	for rows.Next() {
		var ub domain.UserBadge
		err := rows.Scan(&ub.UserID, &ub.BadgeID, &ub.AwardedAt)
		if err != nil {
			return nil, err
		}
		ubs = append(ubs, ub)
	}
	return ubs, nil
}

func (r *UserRepository) AwardBadge(ctx context.Context, userID, badgeID uuid.UUID) error {
	query := `
		INSERT INTO user_badges (user_id, badge_id, awarded_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, badge_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, badgeID)
	return err
}

func (r *UserRepository) GetBadgeByCode(ctx context.Context, code string) (*domain.Badge, error) {
	query := `
		SELECT id, code, name, description, icon_url, xp_bonus
		FROM badges
		WHERE code = $1
	`
	var b domain.Badge
	err := r.pool.QueryRow(ctx, query, code).Scan(&b.ID, &b.Code, &b.Name, &b.Description, &b.IconURL, &b.XPBonus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

