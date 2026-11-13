package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockLMSRepoForClassStats struct {
	students []domain.ClassStudentStats
}

func (m *mockLMSRepoForClassStats) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	return nil, nil
}
func (m *mockLMSRepoForClassStats) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return m.students, nil
}

func TestLMSHandler_GetClassStats_Success(t *testing.T) {
	classID := uuid.New()
	studentID := uuid.New()
	students := []domain.ClassStudentStats{
		{
			UserID:        studentID,
			FullName:      "Гульназ Баширова",
			TotalXP:       320,
			CurrentStreak: 5,
			SolvedToday:   5,
		},
	}

	lmsRepo := &mockLMSRepoForClassStats{students: students}
	lmsUC := usecase.NewLMSUseCase(lmsRepo)
	handler := v1.NewLMSHandler(lmsUC)

	r := chi.NewRouter()
	r.Get("/teacher/classes/{class_id}", handler.GetClassStats)

	req := httptest.NewRequest(http.MethodGet, "/teacher/classes/"+classID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d", rec.Code)
	}

	var resp []domain.ClassStudentStats
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if len(resp) != 1 || resp[0].FullName != "Гульназ Баширова" {
		t.Errorf("некорректные данные учеников класса")
	}
}

func TestLMSHandler_GetClassStats_InvalidUUID(t *testing.T) {
	lmsUC := usecase.NewLMSUseCase(&mockLMSRepoForClassStats{})
	handler := v1.NewLMSHandler(lmsUC)

	r := chi.NewRouter()
	r.Get("/teacher/classes/{class_id}", handler.GetClassStats)

	req := httptest.NewRequest(http.MethodGet, "/teacher/classes/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ожидался статус 400 Bad Request при невалидном UUID, получен %d", rec.Code)
	}
}
