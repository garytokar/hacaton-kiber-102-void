package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/delivery/http/middleware"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockLMSRepoForLB struct {
	schools []domain.LeaderboardSchoolEntry
	classes []domain.LeaderboardClassEntry
}

func (m *mockLMSRepoForLB) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	return m.schools, nil
}
func (m *mockLMSRepoForLB) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return nil, nil
}
func (m *mockLMSRepoForLB) GetClassesLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardClassEntry, error) {
	return m.classes, nil
}
func (m *mockLMSRepoForLB) CreateClass(ctx context.Context, c *domain.Class) error {
	return nil
}
func (m *mockLMSRepoForLB) GetClassByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	return nil, nil
}
func (m *mockLMSRepoForLB) GetClassByInviteCode(ctx context.Context, code string) (*domain.Class, error) {
	return nil, nil
}
func (m *mockLMSRepoForLB) UpdateClassInviteCode(ctx context.Context, classID uuid.UUID, code string) error {
	return nil
}
func (m *mockLMSRepoForLB) AssignUserToClass(ctx context.Context, userID, classID, schoolID uuid.UUID) error {
	return nil
}

func TestLeaderboardHandler_GetUsers(t *testing.T) {
	userRepo := newMockUserRepo()
	_ = userRepo.Create(context.Background(), &domain.User{
		ID:       uuid.New(),
		FullName: "Азамат",
		TotalXP:  400,
	})

	cfg := &config.Config{}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, &mockRedisCache{}, cfg)
	lmsUC := usecase.NewLMSUseCase(&mockLMSRepoForLB{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboards/users?limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var users []domain.LeaderboardUserEntry
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("ожидался 1 пользователь в рейтинге, получено %d", len(users))
	}
	if users[0].FullName != "Азамат" {
		t.Errorf("некорректное имя пользователя в топе")
	}
}

func TestLeaderboardHandler_GetUsers_Period(t *testing.T) {
	userRepo := newMockUserRepo()
	_ = userRepo.Create(context.Background(), &domain.User{
		ID:       uuid.New(),
		FullName: "Гульназ",
		TotalXP:  250,
	})

	cfg := &config.Config{}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, &mockRedisCache{}, cfg)
	lmsUC := usecase.NewLMSUseCase(&mockLMSRepoForLB{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboards/users?period=week&limit=10", nil)
	rec := httptest.NewRecorder()

	handler.GetUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var users []domain.LeaderboardUserEntry
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if len(users) != 1 || users[0].FullName != "Гульназ" {
		t.Errorf("некорректный пользователь в топе за период")
	}
}

func TestLeaderboardHandler_GetSchools(t *testing.T) {
	schools := []domain.LeaderboardSchoolEntry{
		{
			Rank:           1,
			SchoolID:       uuid.New(),
			SchoolName:     "Гимназия № 1",
			City:           "Уфа",
			TotalXP:        1500,
			ActiveStudents: 45,
		},
	}
	lmsRepo := &mockLMSRepoForLB{schools: schools}
	lmsUC := usecase.NewLMSUseCase(lmsRepo)
	gamifyUC := usecase.NewGamificationUseCase(newMockUserRepo(), nil, &mockRedisCache{}, &config.Config{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboards/schools", nil)
	rec := httptest.NewRecorder()

	handler.GetSchools(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var respSchools []domain.LeaderboardSchoolEntry
	if err := json.NewDecoder(rec.Body).Decode(&respSchools); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(respSchools) != 1 || respSchools[0].SchoolName != "Гимназия № 1" {
		t.Errorf("некорректные данные рейтинга школ")
	}
}

func TestLeaderboardHandler_GetClasses(t *testing.T) {
	classes := []domain.LeaderboardClassEntry{
		{
			Rank:           1,
			ClassID:        uuid.New(),
			SchoolID:       uuid.New(),
			SchoolName:     "БРГИ № 1",
			GradeNumber:    10,
			Letter:         "А",
			ClassName:      "10 А",
			TotalXP:        3200,
			ActiveStudents: 25,
		},
	}
	lmsRepo := &mockLMSRepoForLB{classes: classes}
	lmsUC := usecase.NewLMSUseCase(lmsRepo)
	gamifyUC := usecase.NewGamificationUseCase(newMockUserRepo(), nil, &mockRedisCache{}, &config.Config{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboards/classes", nil)
	rec := httptest.NewRecorder()

	handler.GetClasses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var respClasses []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&respClasses); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(respClasses) != 1 || respClasses[0]["class_name"] != "10 А" {
		t.Errorf("некорректные данные рейтинга классов: %+v", respClasses)
	}
}

func TestLeaderboardHandler_GetVocabulary(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.vocabEntries = []domain.LeaderboardVocabularyEntry{
		{
			Rank:           1,
			UserID:         uuid.New(),
			FullName:       "Ильдар Сабитов",
			WordsCount:     54,
			VocabularySize: 54,
			TotalXP:        680,
		},
	}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, &mockRedisCache{}, &config.Config{})
	lmsUC := usecase.NewLMSUseCase(&mockLMSRepoForLB{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboards/vocabulary", nil)
	rec := httptest.NewRecorder()

	handler.GetVocabulary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var resp []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(resp) != 1 || resp[0]["full_name"] != "Ильдар Сабитов" {
		t.Errorf("некорректные данные лидерборда словаря")
	}
}

func TestLeaderboardHandler_GetAchievements(t *testing.T) {
	userRepo := newMockUserRepo()
	badgeID := uuid.New()
	userRepo.allBadges = []domain.Badge{
		{
			ID:          badgeID,
			Code:        "FIRST_STEP",
			Name:        "Первый шаг",
			Description: "Первый подтвержденный перевод",
			XPBonus:     20,
		},
	}

	userID := uuid.New()
	userRepo.userBadges[userID] = []domain.UserBadge{
		{
			UserID:    userID,
			BadgeID:   badgeID,
			AwardedAt: time.Now(),
		},
	}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, nil, &mockRedisCache{}, &config.Config{})
	lmsUC := usecase.NewLMSUseCase(&mockLMSRepoForLB{})

	handler := v1.NewLeaderboardHandler(gamifyUC, lmsUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard/achievements", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetAchievements(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var achievements []domain.UserBadgeDTO
	if err := json.NewDecoder(rec.Body).Decode(&achievements); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(achievements) != 1 || !achievements[0].Unlocked {
		t.Errorf("ожидалось разблокированное достижение: %+v", achievements)
	}
}
