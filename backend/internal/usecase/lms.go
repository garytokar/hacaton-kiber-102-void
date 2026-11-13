package usecase

import (
	"context"

	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
)

type LMSUseCase struct {
	lmsRepo domain.LMSRepository
}

func NewLMSUseCase(lmsRepo domain.LMSRepository) *LMSUseCase {
	return &LMSUseCase{lmsRepo: lmsRepo}
}

func (uc *LMSUseCase) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.lmsRepo.GetSchoolsLeaderboard(ctx, limit)
}

func (uc *LMSUseCase) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return uc.lmsRepo.GetClassStudents(ctx, classID)
}
