package v1_test

import (
	"bytes"
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
	"bashkir-backend/internal/pkg/jwt"
	"bashkir-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockContentRepoForAdmin struct {
	items        []domain.ExpertReviewItem
	translations map[uuid.UUID]*domain.FragmentTranslation
	fragments    map[uuid.UUID]*domain.Fragment
	rejectedIDs  map[uuid.UUID]bool
}

func (m *mockContentRepoForAdmin) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	return nil, nil
}
func (m *mockContentRepoForAdmin) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	ft, ok := m.translations[id]
	if !ok {
		return nil, domain.ErrFragmentNotFound
	}
	return ft, nil
}
func (m *mockContentRepoForAdmin) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	return nil
}
func (m *mockContentRepoForAdmin) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	f, ok := m.fragments[id]
	if !ok {
		return nil, domain.ErrFragmentNotFound
	}
	return f, nil
}
func (m *mockContentRepoForAdmin) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForAdmin) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return m.items, nil
}
func (m *mockContentRepoForAdmin) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	ft, ok := m.translations[id]
	if !ok {
		return nil, domain.ErrFragmentNotFound
	}
	frag := m.fragments[ft.FragmentID]
	text := approvedText
	if text == "" {
		text = ft.InitialAutoTranslation
	}
	ft.Status = domain.StatusConfirmed
	ft.ConfirmedTranslation = &text

	return &domain.ConfirmedTranslation{
		ID:             uuid.New(),
		FragmentID:     ft.FragmentID,
		TextBa:         frag.TextBa,
		LangPair:       ft.LangPair,
		Translation:    text,
		SourceCoord:    "Тестовый источник",
		ConfirmedCount: ft.ConfirmedCount + 1,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}
func (m *mockContentRepoForAdmin) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	ft, ok := m.translations[id]
	if !ok {
		return domain.ErrFragmentNotFound
	}
	ft.Status = domain.StatusRejected
	if m.rejectedIDs == nil {
		m.rejectedIDs = make(map[uuid.UUID]bool)
	}
	m.rejectedIDs[id] = true
	return nil
}

func setupAdminRouter(repo *mockContentRepoForAdmin, secret string) *chi.Mux {
	cfg := &config.Config{ConsensusMinVotes: 3}
	consensusUC := usecase.NewConsensusUseCase(repo, nil, cfg)
	handler := v1.NewAdminHandler(consensusUC)

	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.Auth(secret))
		r.Use(middleware.RequireRoles("admin", "expert"))
		r.Get("/review", handler.GetReviewQueue)
		r.Post("/review/{id}/approve", handler.ApproveTranslation)
		r.Post("/review/{id}/reject", handler.RejectTranslation)
	})
	return r
}

func TestAdminHandler_GetReviewQueue_Authorization(t *testing.T) {
	secret := "admin-test-secret-12345"
	transID := uuid.New()
	fragID := uuid.New()

	repo := &mockContentRepoForAdmin{
		items: []domain.ExpertReviewItem{
			{
				ID:                     transID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				TextBa:                 "Ағиҙел буйлап аҡ пароход йөҙә.",
				SourceTitle:            "Рассказы о родном крае",
				InitialAutoTranslation: "По реке Белой плывет белый пароход.",
				Status:                 "expert_review",
				ConfirmedCount:         2,
				VotesCount:             4,
			},
		},
		translations: map[uuid.UUID]*domain.FragmentTranslation{
			transID: {
				ID:                     transID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				InitialAutoTranslation: "По реке Белой плывет белый пароход.",
				Status:                 domain.StatusExpertReview,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {
				ID:     fragID,
				TextBa: "Ағиҙел буйлап аҡ пароход йөҙә.",
			},
		},
	}

	router := setupAdminRouter(repo, secret)

	// 1. Без токена -> 401 Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	recUnauth := httptest.NewRecorder()
	router.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 для неавторизованного запроса, получен %d", recUnauth.Code)
	}

	// 2. С ролью student -> 403 Forbidden
	studentToken, _, _ := jwt.GenerateToken(uuid.New(), "student", secret, time.Hour)
	reqStudent := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	reqStudent.Header.Set("Authorization", "Bearer "+studentToken)
	recStudent := httptest.NewRecorder()
	router.ServeHTTP(recStudent, reqStudent)
	if recStudent.Code != http.StatusForbidden {
		t.Fatalf("ожидался статус 403 для роли student, получен %d", recStudent.Code)
	}

	// 3. С ролью teacher -> 403 Forbidden
	teacherToken, _, _ := jwt.GenerateToken(uuid.New(), "teacher", secret, time.Hour)
	reqTeacher := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	reqTeacher.Header.Set("Authorization", "Bearer "+teacherToken)
	recTeacher := httptest.NewRecorder()
	router.ServeHTTP(recTeacher, reqTeacher)
	if recTeacher.Code != http.StatusForbidden {
		t.Fatalf("ожидался статус 403 для роли teacher, получен %d", recTeacher.Code)
	}

	// 4. С ролью expert -> 200 OK
	expertToken, _, _ := jwt.GenerateToken(uuid.New(), "expert", secret, time.Hour)
	reqExpert := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	reqExpert.Header.Set("Authorization", "Bearer "+expertToken)
	recExpert := httptest.NewRecorder()
	router.ServeHTTP(recExpert, reqExpert)
	if recExpert.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK для эксперта УФИЦ РАН, получен %d: %s", recExpert.Code, recExpert.Body.String())
	}

	var items []domain.ExpertReviewItem
	if err := json.NewDecoder(recExpert.Body).Decode(&items); err != nil {
		t.Fatalf("ошибка декодирования json: %v", err)
	}
	if len(items) != 1 || items[0].TextBa != "Ағиҙел буйлап аҡ пароход йөҙә." {
		t.Errorf("некорректные данные в очереди на экспертизу")
	}

	// 5. С ролью admin -> 200 OK
	adminToken, _, _ := jwt.GenerateToken(uuid.New(), "admin", secret, time.Hour)
	reqAdmin := httptest.NewRequest(http.MethodGet, "/admin/review", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	recAdmin := httptest.NewRecorder()
	router.ServeHTTP(recAdmin, reqAdmin)
	if recAdmin.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK для администратора, получен %d", recAdmin.Code)
	}
}

func TestAdminHandler_ApproveTranslation_Success(t *testing.T) {
	secret := "admin-test-secret-12345"
	transID := uuid.New()
	fragID := uuid.New()

	repo := &mockContentRepoForAdmin{
		translations: map[uuid.UUID]*domain.FragmentTranslation{
			transID: {
				ID:                     transID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				InitialAutoTranslation: "Неточный перевод",
				Status:                 domain.StatusExpertReview,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {
				ID:     fragID,
				TextBa: "Китап — белем шишмәһе.",
			},
		},
	}

	router := setupAdminRouter(repo, secret)
	expertToken, _, _ := jwt.GenerateToken(uuid.New(), "expert", secret, time.Hour)

	// Утверждаем перевод с уточнением текста экспертом УФИЦ РАН
	body, _ := json.Marshal(map[string]string{
		"approved_text": "Книга — родник знаний.",
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/review/"+transID.String()+"/approve", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+expertToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK при одобрении перевода, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success              bool                        `json:"success"`
		Message              string                      `json:"message"`
		ConfirmedTranslation *domain.ConfirmedTranslation `json:"confirmed_translation"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if !resp.Success {
		t.Errorf("ожидался success: true")
	}
	if resp.ConfirmedTranslation == nil || resp.ConfirmedTranslation.Translation != "Книга — родник знаний." {
		t.Errorf("некорректный утвержденный перевод в ответе: %+v", resp.ConfirmedTranslation)
	}
}

func TestAdminHandler_RejectTranslation_Success(t *testing.T) {
	secret := "admin-test-secret-12345"
	transID := uuid.New()
	fragID := uuid.New()

	repo := &mockContentRepoForAdmin{
		translations: map[uuid.UUID]*domain.FragmentTranslation{
			transID: {
				ID:                     transID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				InitialAutoTranslation: "Спам или нечитаемый текст",
				Status:                 domain.StatusExpertReview,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {ID: fragID, TextBa: "Тестовый фрагмент"},
		},
		rejectedIDs: make(map[uuid.UUID]bool),
	}

	router := setupAdminRouter(repo, secret)
	expertToken, _, _ := jwt.GenerateToken(uuid.New(), "expert", secret, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/admin/review/"+transID.String()+"/reject", nil)
	req.Header.Set("Authorization", "Bearer "+expertToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK при отклонении перевода, получен %d: %s", rec.Code, rec.Body.String())
	}

	if !repo.rejectedIDs[transID] {
		t.Errorf("перевод должен быть помечен как отклоненный в репозитории")
	}
}

func TestAdminHandler_ApproveTranslation_NotFound(t *testing.T) {
	secret := "admin-test-secret-12345"
	nonExistentID := uuid.New()

	repo := &mockContentRepoForAdmin{
		translations: make(map[uuid.UUID]*domain.FragmentTranslation),
		fragments:    make(map[uuid.UUID]*domain.Fragment),
	}

	router := setupAdminRouter(repo, secret)
	expertToken, _, _ := jwt.GenerateToken(uuid.New(), "expert", secret, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/admin/review/"+nonExistentID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+expertToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404 StatusNotFound для несуществующего перевода, получен %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_RejectTranslation_NotFound(t *testing.T) {
	secret := "admin-test-secret-12345"
	nonExistentID := uuid.New()

	repo := &mockContentRepoForAdmin{
		translations: make(map[uuid.UUID]*domain.FragmentTranslation),
		fragments:    make(map[uuid.UUID]*domain.Fragment),
	}

	router := setupAdminRouter(repo, secret)
	expertToken, _, _ := jwt.GenerateToken(uuid.New(), "expert", secret, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/admin/review/"+nonExistentID.String()+"/reject", nil)
	req.Header.Set("Authorization", "Bearer "+expertToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404 StatusNotFound для несуществующего перевода, получен %d: %s", rec.Code, rec.Body.String())
	}
}


func (m *mockContentRepoForAdmin) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	for _, ft := range m.translations {
		if ft.FragmentID == fragmentID && (langPair == "" || ft.LangPair == langPair) {
			return ft, nil
		}
	}
	return nil, domain.ErrFragmentNotFound
}
