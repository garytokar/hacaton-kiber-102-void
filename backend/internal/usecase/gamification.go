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
	if uc.redisRepo != nil {
		_ = uc.redisRepo.IncrementUserXP(ctx, user.ID, earned)
	}

	// 7. Проверка и выдача достижений (бейджей)
	totalSubs := todayCount + 1
	if allSubs, err := uc.subRepo.CountUserSubmissions(ctx, user.ID); err == nil && allSubs > 0 {
		totalSubs = allSubs + 1
	}
	uc.checkAndAwardBadges(ctx, user.ID, streak, newTotalXP, isHoneypotCaught, totalSubs)

	return &RewardResult{
		XPEarned:         earned,
		TotalXP:          newTotalXP,
		CurrentStreak:    streak,
		IsDailyCompleted: isDailyCompleted,
	}, nil
}

func (uc *GamificationUseCase) checkAndAwardBadges(ctx context.Context, userID uuid.UUID, streak, totalXP int, isHoneypotCaught bool, totalSubs int) {
	if uc.userRepo == nil {
		return
	}

	// 1. FIRST_STEP: Первый шаг
	if b, err := uc.userRepo.GetBadgeByCode(ctx, "FIRST_STEP"); err == nil && b != nil {
		_ = uc.userRepo.AwardBadge(ctx, userID, b.ID)
	}

	// 2. STREAK_7: 7 дней стрика
	if streak >= 7 {
		if b, err := uc.userRepo.GetBadgeByCode(ctx, "STREAK_7"); err == nil && b != nil {
			_ = uc.userRepo.AwardBadge(ctx, userID, b.ID)
		}
	}

	// 3. HONEYPOT_HUNTER: Распознал проверочную ошибку
	if isHoneypotCaught {
		if b, err := uc.userRepo.GetBadgeByCode(ctx, "HONEYPOT_HUNTER"); err == nil && b != nil {
			_ = uc.userRepo.AwardBadge(ctx, userID, b.ID)
		}
	}

	// 4. SCHOLAR: Набрал 500+ XP
	if totalXP >= 500 {
		if b, err := uc.userRepo.GetBadgeByCode(ctx, "SCHOLAR"); err == nil && b != nil {
			_ = uc.userRepo.AwardBadge(ctx, userID, b.ID)
		}
	}

	// 5. VOCAB_MASTER: Проверил 10+ переводов
	if totalSubs >= 10 {
		if b, err := uc.userRepo.GetBadgeByCode(ctx, "VOCAB_MASTER"); err == nil && b != nil {
			_ = uc.userRepo.AwardBadge(ctx, userID, b.ID)
		}
	}
}

func (uc *GamificationUseCase) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	return uc.userRepo.GetTopUsers(ctx, limit)
}

func (uc *GamificationUseCase) GetTopUsersByPeriod(ctx context.Context, period string, limit int) ([]domain.LeaderboardUserEntry, error) {
	if period == "" || period == "all" || period == "all-time" {
		return uc.userRepo.GetTopUsers(ctx, limit)
	}
	return uc.userRepo.GetTopUsersByPeriod(ctx, period, limit)
}

func (uc *GamificationUseCase) GetVocabularyLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardVocabularyEntry, error) {
	return uc.userRepo.GetVocabularyLeaderboard(ctx, limit)
}

func (uc *GamificationUseCase) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]domain.UserBadgeDTO, error) {
	allBadges, err := uc.userRepo.GetAllBadges(ctx)
	if err != nil || len(allBadges) == 0 {
		// Стандартный набор бейджей платформы при отсутствии в БД
		return []domain.UserBadgeDTO{
			{
				Code:        "FIRST_STEP",
				Name:        "Беренсе аҙым (Первый шаг)",
				Description: "Подтвердите свой первый перевод в системе",
				Icon:        "🌱",
				XPBonus:     20,
				RewardXP:    20,
				Unlocked:    userID != uuid.Nil,
			},
			{
				Code:        "STREAK_7",
				Name:        "Ете көн (Семидневный хранитель)",
				Description: "Удерживайте стрик 7 дней подряд без перерыва",
				Icon:        "🔥",
				XPBonus:     100,
				RewardXP:    100,
				Unlocked:    false,
			},
			{
				Code:        "HONEYPOT_HUNTER",
				Name:        "Ҡырағай күҙ (Зоркий сокол)",
				Description: "Разоблачите скрытую проверочную ошибку автоперевода",
				Icon:        "🎯",
				XPBonus:     150,
				RewardXP:    150,
				Unlocked:    false,
			},
			{
				Code:        "SCHOLAR",
				Name:        "Тел ғалимы (Лингвист)",
				Description: "Наберите свыше 500 очков опыта",
				Icon:        "👑",
				XPBonus:     200,
				RewardXP:    200,
				Unlocked:    false,
			},
			{
				Code:        "VOCAB_MASTER",
				Name:        "Һүҙ байлығы (Мастер словаря)",
				Description: "Проверьте 10 или более переводов в системе",
				Icon:        "📚",
				XPBonus:     75,
				RewardXP:    75,
				Unlocked:    false,
			},
		}, nil
	}

	var userBadges []domain.UserBadge
	if userID != uuid.Nil {
		userBadges, _ = uc.userRepo.GetUserBadges(ctx, userID)
	}

	awardedMap := make(map[uuid.UUID]time.Time)
	for _, ub := range userBadges {
		awardedMap[ub.BadgeID] = ub.AwardedAt
	}

	icons := map[string]string{
		"FIRST_STEP":      "🌱",
		"STREAK_7":        "🔥",
		"HONEYPOT_HUNTER": "🎯",
		"SCHOLAR":         "👑",
		"VOCAB_MASTER":    "📚",
	}

	var result []domain.UserBadgeDTO
	for _, b := range allBadges {
		icon := icons[b.Code]
		if icon == "" {
			icon = "🎖️"
		}
		dto := domain.UserBadgeDTO{
			ID:          b.ID,
			Code:        b.Code,
			Name:        b.Name,
			Description: b.Description,
			IconURL:     b.IconURL,
			Icon:        icon,
			XPBonus:     b.XPBonus,
			RewardXP:    b.XPBonus,
		}
		if at, ok := awardedMap[b.ID]; ok {
			dto.Unlocked = true
			dto.UnlockedAt = &at
		}
		result = append(result, dto)
	}
	return result, nil
}

