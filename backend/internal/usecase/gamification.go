package usecase

import (
	"context"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"

	"github.com/google/uuid"
)

type LeaderboardCache interface {
	IncrementUserXP(ctx context.Context, userID uuid.UUID, xp int) error
}

type GamificationUseCase struct {
	userRepo  domain.UserRepository
	subRepo   domain.SubmissionRepository
	redisRepo LeaderboardCache
	cfg       *config.Config
}

func NewGamificationUseCase(
	userRepo domain.UserRepository,
	subRepo domain.SubmissionRepository,
	redisRepo LeaderboardCache,
	cfg *config.Config,
) *GamificationUseCase {
	return &GamificationUseCase{
		userRepo:   userRepo,
		subRepo:    subRepo,
		redisRepo:  redisRepo,
		cfg:        cfg,
	}
}

type RewardResult struct {
	XPEarned         int
	TotalXP          int
	CurrentStreak    int
	IsDailyCompleted bool
}

// AwardSubmissionXP начисляет опыт за сабмит, обновляет стрик и лидерборды
func (uc *GamificationUseCase) AwardSubmissionXP(
	ctx context.Context,
	user *domain.User,
	action domain.SubmissionAction,
	isHoneypotCaught bool,
) (*RewardResult, error) {
	earned := 0

	// 1. Базовый опыт за действие
	switch action {
	case domain.ActionConfirmed:
		earned += uc.cfg.XPConfirmTranslation
	case domain.ActionEdited:
		earned += uc.cfg.XPEditTranslation
	case domain.ActionReportedError:
		earned += 5
	}

	// 2. Бонус за разоблачение honeypot-ловушки
	if isHoneypotCaught {
		earned += uc.cfg.XPCatchHoneypot
	}

	// 3. Расчет серии дней (Streak)
	now := time.Now().UTC()
	streak := user.CurrentStreak
	maxStreak := user.MaxStreak

	if user.LastActiveDate == nil || streak == 0 {
		streak = 1
	} else {
		today := now.Truncate(24 * time.Hour)
		lastDate := user.LastActiveDate.UTC().Truncate(24 * time.Hour)
		diffDays := int(today.Sub(lastDate).Hours() / 24)

		if diffDays == 1 {
			streak++
		} else if diffDays > 1 {
			streak = 1 // Стрик сгорел
		}
	}

	if streak > maxStreak {
		maxStreak = streak
	}

	// 4. Проверка закрытия 5 заданий за сегодня
	todayCount, _ := uc.subRepo.CountTodaySubmissions(ctx, user.ID)
	// Учитываем текущий ответ (+1)
	isDailyCompleted := (todayCount + 1) == 5
	if isDailyCompleted {
		earned += uc.cfg.XPDailyCompletion
	}

	newTotalXP := user.TotalXP + earned

	// 5. Сохранение в Postgres
	err := uc.userRepo.UpdateStreakAndXP(ctx, user.ID, streak, maxStreak, earned, now)
	if err != nil {
		return nil, err
	}
	user.TotalXP = newTotalXP

	// 6. Обновление в Redis ZSET (для мгновенного отображения в топе)
	_ = uc.redisRepo.IncrementUserXP(ctx, user.ID, earned)

	return &RewardResult{
		XPEarned:         earned,
		TotalXP:          newTotalXP,
		CurrentStreak:    streak,
		IsDailyCompleted: isDailyCompleted,
	}, nil
}

func (uc *GamificationUseCase) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	return uc.userRepo.GetTopUsers(ctx, limit)
}
