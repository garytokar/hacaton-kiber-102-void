package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/delivery/http/middleware"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/hasher"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

// mockUserRepo — in-memory реализация domain.UserRepository для unit-тестов
type mockUserRepo struct {
	mu          sync.RWMutex
	usersByID   map[uuid.UUID]*domain.User
	usersByMail map[string]*domain.User
	usersByTg   map[int64]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:   make(map[uuid.UUID]*domain.User),
		usersByMail: make(map[string]*domain.User),
		usersByTg:   make(map[int64]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	m.usersByID[u.ID] = u
	if u.Email != nil {
		m.usersByMail[*u.Email] = u
	}
	if u.TgID != nil {
		m.usersByTg[*u.TgID] = u
	}
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usersByTg[tgID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usersByMail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) UpdateStreakAndXP(ctx context.Context, userID uuid.UUID, streak, maxStreak, addedXP int, activeDate time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usersByID[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.CurrentStreak = streak
	u.MaxStreak = maxStreak
	u.TotalXP += addedXP
	u.LastActiveDate = &activeDate
	return nil
}

func (m *mockUserRepo) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []domain.LeaderboardUserEntry
	for _, u := range m.usersByID {
		list = append(list, domain.LeaderboardUserEntry{
			UserID:        u.ID,
			FullName:      u.FullName,
			TgUsername:    u.TgUsername,
			TotalXP:       u.TotalXP,
			CurrentStreak: u.CurrentStreak,
		})
	}
	return list, nil
}

type mockRedisCache struct {
	incrementedXP int
}

func (m *mockRedisCache) IncrementUserXP(ctx context.Context, userID uuid.UUID, xp int) error {
	m.incrementedXP += xp
	return nil
}

// setupTestHandler вспомогательная функция для инициализации хендлера и моков
func setupTestHandler() (*v1.AuthHandler, *mockUserRepo, *config.Config) {
	cfg := &config.Config{
		JWTSecret:        "test-jwt-secret-key-12345",
		JWTTTL:           time.Hour,
		TelegramBotToken: "mock_token",
	}
	repo := newMockUserRepo()
	authUC := usecase.NewAuthUseCase(repo, cfg)
	handler := v1.NewAuthHandler(authUC, repo)
	return handler, repo, cfg
}

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, _, _ := setupTestHandler()

	payload := map[string]string{
		"full_name": "Салават Батыр",
		"email":     "salawat@example.com",
		"password":  "strongpass123",
		"role":      "student",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201 Created, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.User
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if resp.FullName != "Салават Батыр" {
		t.Errorf("ожидалось имя 'Салават Батыр', получено '%s'", resp.FullName)
	}
	if resp.Email == nil || *resp.Email != "salawat@example.com" {
		t.Errorf("некорректный email в ответе")
	}
	if resp.ID == uuid.Nil {
		t.Errorf("ID пользователя не должен быть пустым")
	}
}

func TestAuthHandler_Register_AlreadyExists(t *testing.T) {
	handler, repo, _ := setupTestHandler()

	// Предварительно создаем пользователя
	email := "duplicate@example.com"
	_ = repo.Create(context.Background(), &domain.User{
		ID:       uuid.New(),
		Email:    &email,
		FullName: "Первый Пользователь",
	})

	payload := map[string]string{
		"full_name": "Второй Пользователь",
		"email":     email,
		"password":  "password123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ожидался статус 400 Bad Request при дубликате email, получен %d", rec.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, repo, _ := setupTestHandler()

	email := "teacher@example.com"
	rawPassword := "teacherpass123"
	hashed, _ := hasher.HashPassword(rawPassword)

	_ = repo.Create(context.Background(), &domain.User{
		ID:             uuid.New(),
		Email:          &email,
		HashedPassword: &hashed,
		FullName:       "Учитель Башкирского",
		Role:           domain.RoleTeacher,
	})

	payload := map[string]string{
		"email":    email,
		"password": rawPassword,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token domain.AuthTokenResponse `json:"token"`
		User  domain.User              `json:"user"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга json: %v", err)
	}

	if resp.Token.AccessToken == "" {
		t.Errorf("access_token не должен быть пустым")
	}
	if resp.User.Role != domain.RoleTeacher {
		t.Errorf("ожидалась роль teacher, получена %s", resp.User.Role)
	}
}

func TestAuthHandler_Login_InvalidPassword(t *testing.T) {
	handler, repo, _ := setupTestHandler()

	email := "student@example.com"
	hashed, _ := hasher.HashPassword("correctpassword")

	_ = repo.Create(context.Background(), &domain.User{
		ID:             uuid.New(),
		Email:          &email,
		HashedPassword: &hashed,
		FullName:       "Ученик",
	})

	payload := map[string]string{
		"email":    email,
		"password": "wrongpassword!",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 Unauthorized при неверном пароле, получен %d", rec.Code)
	}
}

func TestAuthHandler_TelegramWebAppAuth_DevBypass(t *testing.T) {
	handler, _, _ := setupTestHandler()

	payload := map[string]string{
		"init_data": "dev_test",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram-webapp", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.TelegramWebAppAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token domain.AuthTokenResponse `json:"token"`
		User  domain.User              `json:"user"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка парсинга ответа: %v", err)
	}

	if resp.Token.AccessToken == "" {
		t.Errorf("access_token пуст")
	}
	if resp.User.TgID == nil || *resp.User.TgID != 999999999 {
		t.Errorf("ожидался тестовый tg_id 999999999")
	}
}

func TestAuthHandler_Me_Authorized(t *testing.T) {
	handler, repo, _ := setupTestHandler()

	userID := uuid.New()
	email := "me@example.com"
	_ = repo.Create(context.Background(), &domain.User{
		ID:       userID,
		Email:    &email,
		FullName: "Текущий Пользователь",
		TotalXP:  150,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	// Внедряем userID в контекст, эмулируя работу JWT middleware
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 OK, получен %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.User
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != userID {
		t.Errorf("ожидался ID %s, получен %s", userID, resp.ID)
	}
	if resp.TotalXP != 150 {
		t.Errorf("ожидалось 150 XP, получено %d", resp.TotalXP)
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 Unauthorized без токена, получен %d", rec.Code)
	}
}
