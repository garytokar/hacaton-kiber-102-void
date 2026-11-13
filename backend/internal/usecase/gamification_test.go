package usecase_test

import (
	"context"
	"testing"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

// mockUserRepoForGamify — мок хранилища пользователей
type mockUserRepoForGamify struct {
	updatedStreak int
	updatedXP     int
	updatedMax    int
	lastDate      time.Time
}

func (m *mockUserRepoForGamify) Create(ctx context.Context, u *domain.User) error { return nil }
func (m *mockUserRepoForGamify) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForGamify) FindByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForGamify) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForGamify) UpdateStreakAndXP(ctx context.Context, userID uuid.UUID, streak, maxStreak, addedXP int, activeDate time.Time) error {
	m.updatedStreak = streak
	m.updatedMax = maxStreak
	m.updatedXP += addedXP
	m.lastDate = activeDate
	return nil
}
func (m *mockUserRepoForGamify) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	return nil, nil
}

// mockSubRepoForGamify — мок хранилища ответов
type mockSubRepoForGamify struct {
	todayCount int
}

func (m *mockSubRepoForGamify) HasUserSubmitted(ctx context.Context, userID, ftID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockSubRepoForGamify) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	return nil
}
func (m *mockSubRepoForGamify) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.todayCount, nil
}
func (m *mockSubRepoForGamify) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	return nil, 0, nil
}
func (m *mockSubRepoForGamify) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	return nil
}

// mockRedisCache — мок Redis кэша
type mockRedisCache struct {
	incrementedXP int
}

func (m *mockRedisCache) IncrementUserXP(ctx context.Context, userID uuid.UUID, xp int) error {
	m.incrementedXP += xp
	return nil
}

func TestGamification_AwardSubmissionXP_Confirm(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		XPCatchHoneypot:      30,
		XPDailyCompletion:    50,
	}
	userRepo := &mockUserRepoForGamify{}
	subRepo := &mockSubRepoForGamify{todayCount: 1}
	redisCache := &mockRedisCache{}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)

	user := &domain.User{
		ID:            uuid.New(),
		TotalXP:       100,
		CurrentStreak: 0,
		MaxStreak:     0,
	}

	res, err := gamifyUC.AwardSubmissionXP(context.Background(), user, domain.ActionConfirmed, false)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if res.XPEarned != 10 {
		t.Errorf("ожидалось 10 XP, получено %d", res.XPEarned)
	}
	if res.TotalXP != 110 {
		t.Errorf("ожидалось 110 TotalXP, получено %d", res.TotalXP)
	}
	if res.CurrentStreak != 1 {
		t.Errorf("первый стрик должен быть 1, получено %d", res.CurrentStreak)
	}
	if redisCache.incrementedXP != 10 {
		t.Errorf("в Redis должно быть начислено 10 XP, получено %d", redisCache.incrementedXP)
	}
}

func TestGamification_AwardSubmissionXP_HoneypotCaughtAndDailyBonus(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		XPCatchHoneypot:      30,
		XPDailyCompletion:    50,
	}
	userRepo := &mockUserRepoForGamify{}
	// todayCount = 4, значит этот сабмит будет 5-м -> срабатывает бонус закрытия дня!
	subRepo := &mockSubRepoForGamify{todayCount: 4}
	redisCache := &mockRedisCache{}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)

	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	user := &domain.User{
		ID:             uuid.New(),
		TotalXP:        200,
		CurrentStreak:  3,
		MaxStreak:      5,
		LastActiveDate: &yesterday,
	}

	// Действие: отредактировал перевод (25 XP) + поймал honeypot (30 XP) + закрыл 5 заданий (50 XP) = 105 XP
	res, err := gamifyUC.AwardSubmissionXP(context.Background(), user, domain.ActionEdited, true)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	expectedXP := 25 + 30 + 50
	if res.XPEarned != expectedXP {
		t.Errorf("ожидалось %d XP, получено %d", expectedXP, res.XPEarned)
	}
	if !res.IsDailyCompleted {
		t.Errorf("должен быть флаг завершения дневной нормы")
	}
	// Вчера был 3-й день стрика, сегодня непрерывный 4-й день
	if res.CurrentStreak != 4 {
		t.Errorf("ожидался стрик 4, получено %d", res.CurrentStreak)
	}
}

func TestGamification_StreakReset(t *testing.T) {
	cfg := &config.Config{XPConfirmTranslation: 10}
	userRepo := &mockUserRepoForGamify{}
	subRepo := &mockSubRepoForGamify{todayCount: 0}
	redisCache := &mockRedisCache{}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)

	// Пользователь не заходил 3 дня
	threeDaysAgo := time.Now().UTC().Add(-72 * time.Hour)
	user := &domain.User{
		ID:             uuid.New(),
		TotalXP:        300,
		CurrentStreak:  15,
		MaxStreak:      15,
		LastActiveDate: &threeDaysAgo,
	}

	res, err := gamifyUC.AwardSubmissionXP(context.Background(), user, domain.ActionConfirmed, false)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	// Стрик должен сгореть до 1
	if res.CurrentStreak != 1 {
		t.Errorf("стрик должен был сброситься на 1, получено %d", res.CurrentStreak)
	}
}
