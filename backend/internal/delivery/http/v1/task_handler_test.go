package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/delivery/http/middleware"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockContentRepoForTaskH struct {
	tasks    []domain.TaskDTO
	ft       *domain.FragmentTranslation
	lastLang string
}

func (m *mockContentRepoForTaskH) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	m.lastLang = langPair
	return m.tasks, nil
}
func (m *mockContentRepoForTaskH) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	if m.ft != nil && m.ft.ID == id {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}
func (m *mockContentRepoForTaskH) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	return nil
}
func (m *mockContentRepoForTaskH) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	return nil, nil
}
func (m *mockContentRepoForTaskH) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForTaskH) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return nil, nil
}
func (m *mockContentRepoForTaskH) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForTaskH) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockContentRepoForTaskH) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	if m.ft != nil && m.ft.FragmentID == fragmentID && (langPair == "" || m.ft.LangPair == langPair) {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}

type mockSubRepoForTaskH struct {
	submitted map[uuid.UUID]bool
}

func (m *mockSubRepoForTaskH) HasUserSubmitted(ctx context.Context, userID, ftID uuid.UUID) (bool, error) {
	return m.submitted[ftID], nil
}
func (m *mockSubRepoForTaskH) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	if m.submitted == nil {
		m.submitted = make(map[uuid.UUID]bool)
	}
	m.submitted[sub.FragmentTranslationID] = true
	return nil
}
func (m *mockSubRepoForTaskH) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 1, nil
}
func (m *mockSubRepoForTaskH) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 1, nil
}
func (m *mockSubRepoForTaskH) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	return nil, 0, nil
}
func (m *mockSubRepoForTaskH) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	return nil
}

func setupTaskHandler(tasks []domain.TaskDTO, ft *domain.FragmentTranslation, user *domain.User) (*v1.TaskHandler, *mockSubRepoForTaskH, *mockContentRepoForTaskH) {
	contentRepo := &mockContentRepoForTaskH{tasks: tasks, ft: ft}
	subRepo := &mockSubRepoForTaskH{submitted: make(map[uuid.UUID]bool)}
	userRepo := newMockUserRepo()
	if user != nil {
		_ = userRepo.Create(context.Background(), user)
	}

	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		ConsensusMinVotes:    3,
	}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, &mockRedisCache{}, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, userRepo, gamifyUC, consensusUC)

	handler := v1.NewTaskHandler(taskUC)
	return handler, subRepo, contentRepo
}

func TestTaskHandler_GetDailyTasks_Authorized(t *testing.T) {
	userID := uuid.New()
	tasks := []domain.TaskDTO{
		{
			FragmentTranslationID: uuid.New(),
			TextBa:                "Башҡортостан — алтын бишек.",
			SuggestedTranslation:  "Башкортостан — золотая колыбель.",
		},
	}

	handler, _, _ := setupTaskHandler(tasks, nil, &domain.User{ID: userID})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/daily?lang_pair=ba-ru", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetDailyTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.DailyTasksResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if resp.Count != 1 || len(resp.Tasks) != 1 {
		t.Errorf("ожидалась 1 карточка задания")
	}
	if resp.Tasks[0].TextBa != "Башҡортостан — алтын бишек." {
		t.Errorf("некорректный текст задания")
	}
}

func TestTaskHandler_GetDailyTasks_FallbackToUserPreferredLangPair(t *testing.T) {
	userID := uuid.New()
	tasks := []domain.TaskDTO{
		{
			FragmentTranslationID: uuid.New(),
			TextBa:                "Алма ағасынан алыҫ төшмәй.",
			SuggestedTranslation:  "Яблоко от яблони далеко не падает.",
		},
	}

	user := &domain.User{
		ID:                userID,
		PreferredLangPair: "ba-kk",
	}

	handler, _, contentRepo := setupTaskHandler(tasks, nil, user)

	// Делаем запрос БЕЗ явного указания lang_pair в строке запроса
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/daily", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetDailyTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	if contentRepo.lastLang != "ba-kk" {
		t.Errorf("ожидался fallback на preferred_lang_pair пользователя 'ba-kk', получен: '%s'", contentRepo.lastLang)
	}
}

func TestTaskHandler_GetDailyTasks_Unauthorized(t *testing.T) {
	handler, _, _ := setupTaskHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/daily", nil)
	rec := httptest.NewRecorder()

	handler.GetDailyTasks(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 Unauthorized, получен %d", rec.Code)
	}
}

func TestTaskHandler_SubmitTask_Success(t *testing.T) {
	userID := uuid.New()
	ftID := uuid.New()
	user := &domain.User{ID: userID, TotalXP: 100, CurrentStreak: 2}

	ft := &domain.FragmentTranslation{
		ID:                     ftID,
		InitialAutoTranslation: "Башкортостан — наша родина.",
	}

	handler, _, _ := setupTaskHandler(nil, ft, user)

	payload := map[string]any{
		"fragment_translation_id": ftID.String(),
		"action":                  "confirmed",
		"submitted_text":          "confirmed",
		"time_spent_seconds":      8,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/submit", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.SubmitTask(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.SubmitResultDTO
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if resp.XPEarned != 10 {
		t.Errorf("ожидалось 10 XP за подтверждение, получено %d", resp.XPEarned)
	}
	if resp.TotalXP != 110 {
		t.Errorf("ожидалось 110 TotalXP, получено %d", resp.TotalXP)
	}
}

func TestTaskHandler_SubmitTask_AlreadySubmitted(t *testing.T) {
	userID := uuid.New()
	ftID := uuid.New()
	user := &domain.User{ID: userID}

	ft := &domain.FragmentTranslation{
		ID:                     ftID,
		InitialAutoTranslation: "Тест",
	}

	handler, subRepo, _ := setupTaskHandler(nil, ft, user)
	// Отмечаем, что этот фрагмент уже был решен
	subRepo.submitted[ftID] = true

	payload := map[string]any{
		"fragment_translation_id": ftID.String(),
		"action":                  "confirmed",
		"submitted_text":          "confirmed",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/submit", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.SubmitTask(rec, req)

	// Должен вернуть 409 Conflict
	if rec.Code != http.StatusConflict {
		t.Fatalf("ожидался статус 409 Conflict при повторном сабмите, получен %d", rec.Code)
	}
}

func TestTaskHandler_ProposeTranslation_WithFragmentTranslationID(t *testing.T) {
	ftID := uuid.New()
	fragID := uuid.New()
	ft := &domain.FragmentTranslation{
		ID:                     ftID,
		FragmentID:             fragID,
		LangPair:               "ba-ru",
		InitialAutoTranslation: "Тест",
		Status:                 domain.StatusPending,
	}
	userID := uuid.New()
	user := &domain.User{ID: userID, PreferredLangPair: "ba-ru"}
	handler, subRepo, _ := setupTaskHandler(nil, ft, user)

	payload := map[string]any{
		"fragment_translation_id": ftID.String(),
		"translation_text":        "Мой новый литературный перевод",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/translations", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ProposeTranslation(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	if !subRepo.submitted[ftID] {
		t.Errorf("ответ не был зафиксирован в submissionRepo!")
	}
}

func TestTaskHandler_ProposeTranslation_WithFragmentIDLookup(t *testing.T) {
	ftID := uuid.New()
	fragID := uuid.New()
	ft := &domain.FragmentTranslation{
		ID:                     ftID,
		FragmentID:             fragID,
		LangPair:               "ba-ru",
		InitialAutoTranslation: "Тест",
		Status:                 domain.StatusPending,
	}
	userID := uuid.New()
	user := &domain.User{ID: userID, PreferredLangPair: "ba-ru"}
	handler, subRepo, _ := setupTaskHandler(nil, ft, user)

	// Передаем ТОЛЬКО fragment_id, как в React UI
	payload := map[string]any{
		"fragment_id":      fragID.String(),
		"translation_text": "Перевод с поиском по fragment_id",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/translations?lang_pair=ba-ru", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ProposeTranslation(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK при поиске по fragment_id, получен %d: %s", rec.Code, rec.Body.String())
	}

	if !subRepo.submitted[ftID] {
		t.Errorf("ответ не был зафиксирован в submissionRepo при поиске по fragment_id!")
	}
}

func TestTaskHandler_ProposeTranslation_NotFound(t *testing.T) {
	userID := uuid.New()
	user := &domain.User{ID: userID, PreferredLangPair: "ba-ru"}
	handler, _, _ := setupTaskHandler(nil, nil, user)

	payload := map[string]any{
		"fragment_id":      uuid.New().String(),
		"translation_text": "Перевод несуществующего фрагмента",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/translations", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ProposeTranslation(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404 Not Found, получен %d: %s", rec.Code, rec.Body.String())
	}
}
