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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, tg_id, tg_username, email, hashed_password, full_name, role, school_id, class_id, current_streak, max_streak, total_xp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.TgID, u.TgUsername, u.Email, u.HashedPassword, u.FullName,
		string(u.Role), u.SchoolID, u.ClassID, u.CurrentStreak, u.MaxStreak,
		u.TotalXP, u.CreatedAt,
	)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
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
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE tg_id = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, tgID).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
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
		SELECT id, tg_id, tg_username, email, hashed_password, full_name, role, school_id, class_id, current_streak, max_streak, last_active_date, total_xp, created_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.TgID, &u.TgUsername, &u.Email, &u.HashedPassword, &u.FullName,
		&roleStr, &u.SchoolID, &u.ClassID, &u.CurrentStreak, &u.MaxStreak,
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

func (r *UserRepository) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
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
