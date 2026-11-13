package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bashkir-backend/internal/delivery/http/middleware"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockLMSRepoForClassStats struct {
	students []domain.ClassStudentStats
	classes  map[uuid.UUID]*domain.Class
	codes    map[string]*domain.Class
}

func newMockLMSRepo(students []domain.ClassStudentStats) *mockLMSRepoForClassStats {
	return &mockLMSRepoForClassStats{
		students: students,
		classes:  make(map[uuid.UUID]*domain.Class),
		codes:    make(map[string]*domain.Class),
	}
}

func (m *mockLMSRepoForClassStats) GetSchoolsLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardSchoolEntry, error) {
	return nil, nil
}
func (m *mockLMSRepoForClassStats) GetClassStudents(ctx context.Context, classID uuid.UUID) ([]domain.ClassStudentStats, error) {
	return m.students, nil
}
func (m *mockLMSRepoForClassStats) GetClassesLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardClassEntry, error) {
	return nil, nil
}
func (m *mockLMSRepoForClassStats) CreateClass(ctx context.Context, c *domain.Class) error {
	if m.classes == nil {
		m.classes = make(map[uuid.UUID]*domain.Class)
	}
	if m.codes == nil {
		m.codes = make(map[string]*domain.Class)
	}
	m.classes[c.ID] = c
	if c.InviteCode != nil {
		m.codes[*c.InviteCode] = c
	}
	return nil
}
func (m *mockLMSRepoForClassStats) GetClassByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	if c, ok := m.classes[id]; ok {
		return c, nil
	}
	return nil, errors.New("class not found")
}
func (m *mockLMSRepoForClassStats) GetClassByInviteCode(ctx context.Context, code string) (*domain.Class, error) {
	if c, ok := m.codes[code]; ok {
		return c, nil
	}
	return nil, errors.New("class not found")
}
func (m *mockLMSRepoForClassStats) UpdateClassInviteCode(ctx context.Context, classID uuid.UUID, code string) error {
	if c, ok := m.classes[classID]; ok {
		c.InviteCode = &code
		if m.codes == nil {
			m.codes = make(map[string]*domain.Class)
		}
		m.codes[code] = c
		return nil
	}
	return errors.New("class not found")
}
func (m *mockLMSRepoForClassStats) AssignUserToClass(ctx context.Context, userID, classID, schoolID uuid.UUID) error {
	return nil
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

	lmsRepo := newMockLMSRepo(students)
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
	lmsUC := usecase.NewLMSUseCase(newMockLMSRepo(nil))
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

func TestLMSHandler_CreateClass_Success(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	body := map[string]any{
		"school_id":    schoolID,
		"grade_number": 9,
		"letter":       "А",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/teacher/classes", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.CreateClass(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201 Created, получен %d (тело: %s)", rec.Code, rec.Body.String())
	}

	var created domain.Class
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("ошибка декодирования ответа: %v", err)
	}

	if created.SchoolID != schoolID || created.GradeNumber != 9 || created.Letter != "А" {
		t.Errorf("некорректные параметры созданного класса: %+v", created)
	}
	if created.InviteCode == nil || *created.InviteCode == "" {
		t.Errorf("ожидался сгенерированный инвайт-код")
	}
}

func TestLMSHandler_CreateClass_Invalid(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	// Missing school_id
	body := map[string]any{
		"grade_number": 9,
		"letter":       "А",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/teacher/classes", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.CreateClass(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ожидался статус 400 Bad Request при отсутствии school_id, получен %d", rec.Code)
	}
}

func TestLMSHandler_GenerateInviteCode(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	class, err := lmsUC.CreateClass(context.Background(), schoolID, 11, "Б")
	if err != nil {
		t.Fatalf("не удалось создать класс: %v", err)
	}

	r := chi.NewRouter()
	r.Post("/classes/{id}/invite", handler.GenerateInviteCode)

	req := httptest.NewRequest(http.MethodPost, "/classes/"+class.ID.String()+"/invite", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d (тело: %s)", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if resp["invite_code"] == nil || resp["invite_code"] == "" {
		t.Errorf("ожидался непустой инвайт-код в ответе")
	}
}

func TestLMSHandler_JoinClass_Success(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	class, err := lmsUC.CreateClass(context.Background(), schoolID, 7, "В")
	if err != nil {
		t.Fatalf("ошибка создания класса: %v", err)
	}

	userID := uuid.New()
	body := map[string]any{
		"invite_code": *class.InviteCode,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/classes/join", bytes.NewReader(raw))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.JoinClass(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d (тело: %s)", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("ожидался success = true, получено: %v", resp["success"])
	}
}

func TestLMSHandler_JoinClass_Unauthorized(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	body := map[string]any{
		"invite_code": "INV-TEST-123",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/classes/join", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.JoinClass(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 Unauthorized для неавторизованного пользователя, получен %d", rec.Code)
	}
}

func TestLMSHandler_AddStudent_Success(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	class, err := lmsUC.CreateClass(context.Background(), schoolID, 8, "А")
	if err != nil {
		t.Fatalf("ошибка создания класса: %v", err)
	}

	studentID := uuid.New()
	body := map[string]any{
		"user_id": studentID,
	}
	raw, _ := json.Marshal(body)

	r := chi.NewRouter()
	r.Post("/classes/{id}/students", handler.AddStudent)

	req := httptest.NewRequest(http.MethodPost, "/classes/"+class.ID.String()+"/students", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d (тело: %s)", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("ожидался success = true")
	}
}

func TestLMSHandler_JoinClass_WithCodeAlias(t *testing.T) {
	repo := newMockLMSRepo(nil)
	lmsUC := usecase.NewLMSUseCase(repo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	class, err := lmsUC.CreateClass(context.Background(), schoolID, 9, "Б")
	if err != nil {
		t.Fatalf("ошибка создания класса: %v", err)
	}

	studentID := uuid.New()
	// Передаем "code" вместо "invite_code"
	body := map[string]any{
		"code": *class.InviteCode,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/classes/join", bytes.NewReader(raw))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, studentID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.JoinClass(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK при присоединении по алиасу 'code', получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("ожидался success = true, получено: %v", resp["success"])
	}
}

func TestLMSHandler_AddStudent_ByEmail_Success(t *testing.T) {
	repo := newMockLMSRepo(nil)
	userRepo := newMockUserRepo()
	lmsUC := usecase.NewLMSUseCase(repo, userRepo)
	handler := v1.NewLMSHandler(lmsUC)

	schoolID := uuid.New()
	class, err := lmsUC.CreateClass(context.Background(), schoolID, 10, "А")
	if err != nil {
		t.Fatalf("ошибка создания класса: %v", err)
	}

	// Учитель регистрирует нового ученика через email и имя
	body := map[string]any{
		"email":     "new_student@school10.ru",
		"full_name": "Азат Валеев",
		"password":  "MySecretPass123",
	}
	raw, _ := json.Marshal(body)

	r := chi.NewRouter()
	r.Post("/classes/{id}/students", handler.AddStudent)

	req := httptest.NewRequest(http.MethodPost, "/classes/"+class.ID.String()+"/students", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK при регистрации ученика учителем, получен %d (тело: %s)", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка json: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("ожидался success = true")
	}
	if resp["email"] != "new_student@school10.ru" {
		t.Errorf("ожидался email new_student@school10.ru, получен: %v", resp["email"])
	}
	if resp["full_name"] != "Азат Валеев" {
		t.Errorf("ожидался full_name Азат Валеев, получен: %v", resp["full_name"])
	}
}
