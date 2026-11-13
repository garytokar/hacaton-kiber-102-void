package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bashkir-backend/internal/config"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockLMSRepoForLB struct {
	schools []domain.LeaderboardSchoolEntry
}

func (m *mockLMSRepoForLB) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	return m.schools, nil
}
func (m *mockLMSRepoForLB) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return nil, nil
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
