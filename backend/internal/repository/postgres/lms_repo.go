package postgres

import (
	"context"

	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LMSRepository struct {
	pool *pgxpool.Pool
}

func NewLMSRepository(pool *pgxpool.Pool) *LMSRepository {
	return &LMSRepository{pool: pool}
}

func (r *LMSRepository) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	query := `
		SELECT 
			s.id, s.name, s.city, 
			COALESCE(SUM(u.total_xp), 0)::INT as total_xp,
			COUNT(DISTINCT u.id)::INT as active_students
		FROM schools s
		LEFT JOIN users u ON u.school_id = s.id
		GROUP BY s.id, s.name, s.city
		ORDER BY total_xp DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LeaderboardSchoolEntry
	rank := 1
	for rows.Next() {
		var entry domain.LeaderboardSchoolEntry
		entry.Rank = rank
		err := rows.Scan(
			&entry.SchoolID, &entry.SchoolName, &entry.City,
			&entry.TotalXP, &entry.ActiveStudents,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
		rank++
	}
	return result, nil
}

func (r *LMSRepository) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	query := `
		SELECT 
			u.id, u.full_name, u.tg_username, u.total_xp, u.current_streak,
			(SELECT COUNT(*) FROM user_submissions us WHERE us.user_id = u.id AND us.created_at >= CURRENT_DATE)::INT as solved_today
		FROM users u
		WHERE u.class_id = $1
		ORDER BY u.total_xp DESC
	`
	rows, err := r.pool.Query(ctx, query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ClassStudentStats
	for rows.Next() {
		var s domain.ClassStudentStats
		err := rows.Scan(&s.UserID, &s.FullName, &s.TgUsername, &s.TotalXP, &s.CurrentStreak, &s.SolvedToday)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}
