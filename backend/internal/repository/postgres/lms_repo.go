package postgres

import (
	"context"
	"fmt"
	"time"

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

func (r *LMSRepository) GetClassesLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardClassEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT 
			c.id, c.school_id, s.name as school_name, c.grade_number, c.letter,
			COALESCE(SUM(u.total_xp), 0)::INT as total_xp,
			COUNT(DISTINCT u.id)::INT as active_students
		FROM classes c
		JOIN schools s ON c.school_id = s.id
		LEFT JOIN users u ON u.class_id = c.id
		GROUP BY c.id, c.school_id, s.name, c.grade_number, c.letter
		ORDER BY total_xp DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LeaderboardClassEntry
	rank := 1
	for rows.Next() {
		var entry domain.LeaderboardClassEntry
		entry.Rank = rank
		err := rows.Scan(
			&entry.ClassID, &entry.SchoolID, &entry.SchoolName,
			&entry.GradeNumber, &entry.Letter,
			&entry.TotalXP, &entry.ActiveStudents,
		)
		if err != nil {
			return nil, err
		}
		entry.ClassName = fmt.Sprintf("%d %s", entry.GradeNumber, entry.Letter)
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

func (r *LMSRepository) CreateClass(ctx context.Context, c *domain.Class) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO classes (id, school_id, grade_number, letter, invite_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.SchoolID, c.GradeNumber, c.Letter, c.InviteCode, c.CreatedAt)
	return err
}

func (r *LMSRepository) GetClassByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	query := `
		SELECT id, school_id, grade_number, letter, invite_code, created_at
		FROM classes
		WHERE id = $1
	`
	var c domain.Class
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.SchoolID, &c.GradeNumber, &c.Letter, &c.InviteCode, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *LMSRepository) GetClassByInviteCode(ctx context.Context, code string) (*domain.Class, error) {
	query := `
		SELECT id, school_id, grade_number, letter, invite_code, created_at
		FROM classes
		WHERE UPPER(invite_code) = UPPER($1)
	`
	var c domain.Class
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&c.ID, &c.SchoolID, &c.GradeNumber, &c.Letter, &c.InviteCode, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *LMSRepository) UpdateClassInviteCode(ctx context.Context, classID uuid.UUID, code string) error {
	query := `UPDATE classes SET invite_code = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, code, classID)
	return err
}

func (r *LMSRepository) AssignUserToClass(ctx context.Context, userID, classID, schoolID uuid.UUID) error {
	query := `UPDATE users SET class_id = $1, school_id = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, classID, schoolID, userID)
	return err
}

