package usecase_test

import (
	"context"
	"errors"
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
	topUsers      []domain.LeaderboardUserEntry
	periodCalled  string
	vocabUsers    []domain.LeaderboardVocabularyEntry
	allBadges     []domain.Badge
	userBadges    []domain.UserBadge
	awardedBadges []uuid.UUID
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
	return m.topUsers, nil
}
func (m *mockUserRepoForGamify) GetTopUsersByPeriod(ctx context.Context, period string, limit int) ([]domain.LeaderboardUserEntry, error) {
	m.periodCalled = period
	return m.topUsers, nil
}
func (m *mockUserRepoForGamify) GetVocabularyLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardVocabularyEntry, error) {
	return m.vocabUsers, nil
}
func (m *mockUserRepoForGamify) UpdatePreferredLangPair(ctx context.Context, userID uuid.UUID, langPair string) error {
	return nil
}
func (m *mockUserRepoForGamify) UpdateUser(ctx context.Context, u *domain.User) error {
	return nil
}
func (m *mockUserRepoForGamify) GetAllBadges(ctx context.Context) ([]domain.Badge, error) {
	return m.allBadges, nil
}
func (m *mockUserRepoForGamify) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	return m.userBadges, nil
}
func (m *mockUserRepoForGamify) AwardBadge(ctx context.Context, userID, badgeID uuid.UUID) error {
	m.awardedBadges = append(m.awardedBadges, badgeID)
	return nil
}
func (m *mockUserRepoForGamify) GetBadgeByCode(ctx context.Context, code string) (*domain.Badge, error) {
	for _, b := range m.allBadges {
		if b.Code == code {
			return &b, nil
		}
	}
	return nil, errors.New("badge not found")
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
func (m *mockSubRepoForGamify) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
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

func TestGamification_GetTopUsersByPeriod(t *testing.T) {
	userRepo := &mockUserRepoForGamify{
		topUsers: []domain.LeaderboardUserEntry{
			{Rank: 1, FullName: "Ильдар", TotalXP: 500},
		},
	}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, nil, nil)

	// Test "all" period
	entries, err := gamifyUC.GetTopUsersByPeriod(context.Background(), "all", 10)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(entries) != 1 || entries[0].FullName != "Ильдар" {
		t.Errorf("некорректный результат для period=all")
	}
	if userRepo.periodCalled != "" {
		t.Errorf("для period=all не должен вызываться GetTopUsersByPeriod")
	}

	// Test "week" period
	entries, err = gamifyUC.GetTopUsersByPeriod(context.Background(), "week", 10)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(entries) != 1 || userRepo.periodCalled != "week" {
		t.Errorf("ожидался вызов GetTopUsersByPeriod с 'week', вызвано с '%s'", userRepo.periodCalled)
	}
}

func TestGamification_GetVocabularyLeaderboard(t *testing.T) {
	userRepo := &mockUserRepoForGamify{
		vocabUsers: []domain.LeaderboardVocabularyEntry{
			{Rank: 1, FullName: "Айгуль", WordsCount: 30, VocabularySize: 30, TotalXP: 300},
		},
	}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, nil, nil)

	entries, err := gamifyUC.GetVocabularyLeaderboard(context.Background(), 10)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if len(entries) != 1 || entries[0].FullName != "Айгуль" {
		t.Errorf("некорректный результат словаря: %+v", entries)
	}
}

func TestGamification_GetUserAchievements_CustomBadges(t *testing.T) {
	b1 := domain.Badge{ID: uuid.New(), Code: "FIRST_STEP", Name: "Первый шаг"}
	b2 := domain.Badge{ID: uuid.New(), Code: "STREAK_7", Name: "7 дней"}
	userID := uuid.New()

	userRepo := &mockUserRepoForGamify{
		allBadges: []domain.Badge{b1, b2},
		userBadges: []domain.UserBadge{
			{UserID: userID, BadgeID: b1.ID, AwardedAt: time.Now()},
		},
	}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, nil, nil)

	achievements, err := gamifyUC.GetUserAchievements(context.Background(), userID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if len(achievements) != 2 {
		t.Fatalf("ожидалось 2 достижения, получено %d", len(achievements))
	}

	if achievements[0].Code != "FIRST_STEP" || !achievements[0].Unlocked {
		t.Errorf("FIRST_STEP должен быть разблокирован")
	}
	if achievements[1].Code != "STREAK_7" || achievements[1].Unlocked {
		t.Errorf("STREAK_7 должен быть заблокирован")
	}
}

func TestGamification_GetUserAchievements_DefaultFallback(t *testing.T) {
	userRepo := &mockUserRepoForGamify{
		allBadges: nil, // empty DB
	}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, nil, nil)

	userID := uuid.New()
	achievements, err := gamifyUC.GetUserAchievements(context.Background(), userID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if len(achievements) < 5 {
		t.Fatalf("ожидалось минимум 5 стандартных бейджей, получено %d", len(achievements))
	}
}

func TestGamification_AwardSubmissionXP_BadgesAwarded(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		XPCatchHoneypot:      30,
		XPDailyCompletion:    50,
	}

	bStep := domain.Badge{ID: uuid.New(), Code: "FIRST_STEP"}
	bStreak := domain.Badge{ID: uuid.New(), Code: "STREAK_7"}
	bHoney := domain.Badge{ID: uuid.New(), Code: "HONEYPOT_HUNTER"}
	bScholar := domain.Badge{ID: uuid.New(), Code: "SCHOLAR"}
	bVocab := domain.Badge{ID: uuid.New(), Code: "VOCAB_MASTER"}

	userRepo := &mockUserRepoForGamify{
		allBadges: []domain.Badge{bStep, bStreak, bHoney, bScholar, bVocab},
	}
	subRepo := &mockSubRepoForGamify{todayCount: 9} // with current submission, total = 10 -> triggers VOCAB_MASTER
	redisCache := &mockRedisCache{}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)

	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	user := &domain.User{
		ID:             uuid.New(),
		TotalXP:        490, // with +30 honeypot + 10 confirm, total will exceed 500 -> triggers SCHOLAR
		CurrentStreak:  6,   // with consecutive day, streak becomes 7 -> triggers STREAK_7
		MaxStreak:      6,
		LastActiveDate: &yesterday,
	}

	// Action: Confirmed (10 XP) + Honeypot caught (30 XP) -> TotalXP 530, Streak 7, Subs 10
	res, err := gamifyUC.AwardSubmissionXP(context.Background(), user, domain.ActionConfirmed, true)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if res.CurrentStreak != 7 {
		t.Errorf("ожидался стрик 7, получено %d", res.CurrentStreak)
	}
	if res.TotalXP < 500 {
		t.Errorf("ожидался TotalXP >= 500, получено %d", res.TotalXP)
	}

	// Verify all 5 badges were awarded
	if len(userRepo.awardedBadges) != 5 {
		t.Fatalf("ожидалось 5 начисленных бейджей, получено %d", len(userRepo.awardedBadges))
	}
}
