package usecase_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockContentRepoForTask struct {
	tasks []domain.TaskDTO
	ft    *domain.FragmentTranslation
}

func (m *mockContentRepoForTask) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	return m.tasks, nil
}
func (m *mockContentRepoForTask) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	if m.ft != nil && m.ft.ID == id {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}
func (m *mockContentRepoForTask) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	return nil
}
func (m *mockContentRepoForTask) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	return nil, nil
}
func (m *mockContentRepoForTask) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForTask) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return nil, nil
}
func (m *mockContentRepoForTask) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForTask) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockContentRepoForTask) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	if m.ft != nil && m.ft.FragmentID == fragmentID && (langPair == "" || m.ft.LangPair == langPair) {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}



type mockSubRepoForTask struct {
	submitted   map[uuid.UUID]bool
	createdSubs []*domain.UserSubmission
	todayCount  int
}

func (m *mockSubRepoForTask) HasUserSubmitted(ctx context.Context, userID, ftID uuid.UUID) (bool, error) {
	return m.submitted[ftID], nil
}
func (m *mockSubRepoForTask) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	m.createdSubs = append(m.createdSubs, sub)
	return nil
}
func (m *mockSubRepoForTask) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.todayCount, nil
}
func (m *mockSubRepoForTask) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.todayCount, nil
}
func (m *mockSubRepoForTask) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	return nil, 0, nil
}
func (m *mockSubRepoForTask) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	return nil
}

func TestTaskUseCase_GetDailyTasks(t *testing.T) {
	fragID := uuid.New()
	contentRepo := &mockContentRepoForTask{
		tasks: []domain.TaskDTO{
			{
				FragmentTranslationID: uuid.New(),
				FragmentID:            fragID,
				TextBa:                "Башҡорт теле — бай тел.",
				LangPair:              "ba-ru",
				SuggestedTranslation:  "Башкирский язык — богатый язык.",
			},
		},
	}
	taskUC := usecase.NewTaskUseCase(contentRepo, &mockSubRepoForTask{}, &mockUserRepoForGamify{}, nil, nil)

	resp, err := taskUC.GetDailyTasks(context.Background(), uuid.New(), "ba-ru")
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if resp.Count != 1 || len(resp.Tasks) != 1 {
		t.Errorf("ожидалось 1 задание, получено %d", resp.Count)
	}
	if resp.Tasks[0].TextBa != "Башҡорт теле — бай тел." {
		t.Errorf("некорректный текст задания")
	}
}

func TestTaskUseCase_SubmitAnswer_SuccessAndHoneypot(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		XPCatchHoneypot:      30,
	}

	ftID := uuid.New()
	userID := uuid.New()
	wrongTrap := "Искаженный перевод"

	contentRepo := &mockContentRepoForTask{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			HasHoneypot:            true,
			HoneypotVariant:        &wrongTrap,
			InitialAutoTranslation: "Правильный перевод",
		},
	}

	subRepo := &mockSubRepoForTask{
		submitted: make(map[uuid.UUID]bool),
	}

	userRepo := &mockUserRepoForGamify{}
	redisCache := &mockRedisCache{}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)

	// Добавим фейкового пользователя в userRepo
	fakeUser := &domain.User{ID: userID, TotalXP: 50, CurrentStreak: 1}
	mockRepoWithUser := &mockUserRepoWithFind{user: fakeUser}

	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, mockRepoWithUser, gamifyUC, consensusUC)

	// Пользователь исправил ловушку!
	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: ftID,
		Action:                domain.ActionEdited,
		SubmittedText:         "Мой правильный перевод",
		TimeSpentSeconds:      10,
	}

	res, err := taskUC.SubmitAnswer(context.Background(), req)
	if err != nil {
		t.Fatalf("ошибка сабмита: %v", err)
	}

	if !res.IsHoneypotCaught {
		t.Errorf("ловушка должна была быть распознана как пойманная")
	}
	// Опыт: 25 за исправление + 30 за ловушку = 55 XP
	if res.XPEarned != 55 {
		t.Errorf("ожидалось 55 XP, получено %d", res.XPEarned)
	}

	if len(subRepo.createdSubs) != 1 {
		t.Fatalf("сабмит не был записан в репозиторий")
	}
	if subRepo.createdSubs[0].SubmittedText != "Мой правильный перевод" {
		t.Errorf("некорректный текст сабмита")
	}
}

type mockUserRepoWithFind struct {
	mockUserRepoForGamify
	user *domain.User
}

func (m *mockUserRepoWithFind) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.user != nil && m.user.ID == id {
		return m.user, nil
	}
	return nil, domain.ErrUserNotFound
}

func TestTaskUseCase_SubmitAnswer_ProfanityBlocked(t *testing.T) {
	ftID := uuid.New()
	userID := uuid.New()

	contentRepo := &mockContentRepoForTask{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			InitialAutoTranslation: "Дружба — богатство жизни.",
		},
	}
	subRepo := &mockSubRepoForTask{submitted: make(map[uuid.UUID]bool)}
	userRepo := &mockUserRepoWithFind{user: &domain.User{ID: userID}}
	cfg := &config.Config{XPEditTranslation: 25}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, &mockRedisCache{}, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, userRepo, gamifyUC, consensusUC)

	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: ftID,
		Action:                domain.ActionEdited,
		SubmittedText:         "хуй",
	}

	_, err := taskUC.SubmitAnswer(context.Background(), req)
	if err != domain.ErrProfanityDetected {
		t.Fatalf("ожидалась ошибка ErrProfanityDetected, получена: %v", err)
	}

	if len(subRepo.createdSubs) != 0 {
		t.Errorf("нецензурный сабмит ни в коем случае не должен сохраняться в БД!")
	}
}

func TestTaskUseCase_SubmitAnswer_ReportError(t *testing.T) {
	ftID := uuid.New()
	userID := uuid.New()

	contentRepo := &mockContentRepoForTask{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			InitialAutoTranslation: "Неправильный перевод",
		},
	}
	subRepo := &mockSubRepoForTask{submitted: make(map[uuid.UUID]bool)}
	userRepo := &mockUserRepoWithFind{user: &domain.User{ID: userID, TotalXP: 10}}
	cfg := &config.Config{ConsensusMinVotes: 3}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, &mockRedisCache{}, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, userRepo, gamifyUC, consensusUC)

	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: ftID,
		Action:                domain.ActionReportedError,
		SubmittedText:         "",
	}

	result, err := taskUC.SubmitAnswer(context.Background(), req)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if result.XPEarned != 5 {
		t.Errorf("ожидалось 5 XP за жалобу на ошибку, получено %d", result.XPEarned)
	}
	if len(subRepo.createdSubs) != 1 {
		t.Fatalf("сабмит не был сохранен")
	}
	if subRepo.createdSubs[0].SubmittedText != "[ERROR_REPORTED]" {
		t.Errorf("ожидался текст [ERROR_REPORTED], получен: %s", subRepo.createdSubs[0].SubmittedText)
	}
}

// panickingSubRepoForConsensus вызывает панику при попытке агрегации голосов консенсуса
type panickingSubRepoForConsensus struct {
	mockSubRepoForTask
	panicCh chan struct{}
	once    sync.Once
}

func (m *panickingSubRepoForConsensus) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	if m.panicCh != nil {
		m.once.Do(func() {
			close(m.panicCh)
		})
	}
	panic("critical database failure inside consensus routine")
}

// TestTaskUseCase_SubmitAnswer_ConsensusPanicRecovery проверяет, что паника в фоне консенсуса перехватывается и не роняет процесс
func TestTaskUseCase_SubmitAnswer_ConsensusPanicRecovery(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		ConsensusMinVotes:    3,
	}

	ftID := uuid.New()
	userID := uuid.New()

	contentRepo := &mockContentRepoForTask{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			InitialAutoTranslation: "Башкортостан — прекрасный край",
			Status:                 domain.StatusPending,
		},
	}

	panicTriggered := make(chan struct{})
	panickingSubRepo := &panickingSubRepoForConsensus{
		mockSubRepoForTask: mockSubRepoForTask{
			submitted: make(map[uuid.UUID]bool),
		},
		panicCh: panicTriggered,
	}

	userRepo := &mockUserRepoWithFind{user: &domain.User{ID: userID, TotalXP: 20}}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, panickingSubRepo, &mockRedisCache{}, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, panickingSubRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, panickingSubRepo, userRepo, gamifyUC, consensusUC)

	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: ftID,
		Action:                domain.ActionConfirmed,
		SubmittedText:         "confirmed",
	}

	res, err := taskUC.SubmitAnswer(context.Background(), req)
	if err != nil {
		t.Fatalf("SubmitAnswer должен завершиться успешно, несмотря на сбой консенсуса: %v", err)
	}

	if res.XPEarned != 10 {
		t.Errorf("ожидалось 10 XP, получено %d", res.XPEarned)
	}

	// Убедимся, что фоновая горутина запустилась, вызвала панику и была безопасно перехвачена
	select {
	case <-panicTriggered:
		// Паника произошла внутри горутины и была перехвачена recover
	case <-time.After(2 * time.Second):
		t.Fatal("таймаут ожидания паники в фоновой горутине консенсуса")
	}

	// Небольшая пауза для гарантированного завершения defer recover
	time.Sleep(50 * time.Millisecond)
}

// TestTaskUseCase_SubmitAnswer_NilConsensusSafe проверяет безопасность при неинициализированном consensusUC
func TestTaskUseCase_SubmitAnswer_NilConsensusSafe(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
	}
	ftID := uuid.New()
	userID := uuid.New()

	contentRepo := &mockContentRepoForTask{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			InitialAutoTranslation: "Башкортостан",
		},
	}
	subRepo := &mockSubRepoForTask{submitted: make(map[uuid.UUID]bool)}
	userRepo := &mockUserRepoWithFind{user: &domain.User{ID: userID}}
	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, &mockRedisCache{}, cfg)

	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, userRepo, gamifyUC, nil)

	req := domain.SubmitAnswerReq{
		UserID:                userID,
		FragmentTranslationID: ftID,
		Action:                domain.ActionConfirmed,
	}

	res, err := taskUC.SubmitAnswer(context.Background(), req)
	if err != nil {
		t.Fatalf("неожиданная ошибка при nil consensus: %v", err)
	}
	if res == nil {
		t.Fatal("ожидался непустой результат")
	}
}

// Потокобезопасные моки для тестирования конкурентных запросов
type mockConcurrentContentRepo struct {
	mu sync.Mutex
	ft *domain.FragmentTranslation
}

func (m *mockConcurrentContentRepo) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	return nil, nil
}
func (m *mockConcurrentContentRepo) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ft != nil && m.ft.ID == id {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}
func (m *mockConcurrentContentRepo) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	return nil
}
func (m *mockConcurrentContentRepo) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	return nil, nil
}
func (m *mockConcurrentContentRepo) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockConcurrentContentRepo) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return nil, nil
}
func (m *mockConcurrentContentRepo) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockConcurrentContentRepo) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockConcurrentContentRepo) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ft != nil && m.ft.FragmentID == fragmentID && (langPair == "" || m.ft.LangPair == langPair) {
		return m.ft, nil
	}
	return nil, domain.ErrFragmentNotFound
}


type mockConcurrentSubRepo struct {
	mu             sync.Mutex
	submitted      map[string]bool
	createdSubs    []*domain.UserSubmission
	panicConsensus bool
}

func (m *mockConcurrentSubRepo) HasUserSubmitted(ctx context.Context, userID, ftID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", userID, ftID)
	return m.submitted[key], nil
}

func (m *mockConcurrentSubRepo) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createdSubs = append(m.createdSubs, sub)
	key := fmt.Sprintf("%s:%s", sub.UserID, sub.FragmentTranslationID)
	m.submitted[key] = true
	return nil
}

func (m *mockConcurrentSubRepo) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockConcurrentSubRepo) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockConcurrentSubRepo) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	if m.panicConsensus {
		panic("database deadlock during concurrent consensus grouping")
	}
	return nil, 0, nil
}

func (m *mockConcurrentSubRepo) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	return nil
}

type mockConcurrentUserRepo struct {
	mu sync.Mutex
}

func (m *mockConcurrentUserRepo) Create(ctx context.Context, u *domain.User) error { return nil }
func (m *mockConcurrentUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return &domain.User{ID: id, TotalXP: 100, CurrentStreak: 1}, nil
}
func (m *mockConcurrentUserRepo) FindByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) UpdateStreakAndXP(ctx context.Context, userID uuid.UUID, streak, maxStreak, addedXP int, activeDate time.Time) error {
	return nil
}
func (m *mockConcurrentUserRepo) GetTopUsers(ctx context.Context, limit int) ([]domain.LeaderboardUserEntry, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) GetTopUsersByPeriod(ctx context.Context, period string, limit int) ([]domain.LeaderboardUserEntry, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) GetVocabularyLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardVocabularyEntry, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) UpdatePreferredLangPair(ctx context.Context, userID uuid.UUID, langPair string) error {
	return nil
}
func (m *mockConcurrentUserRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	return nil
}
func (m *mockConcurrentUserRepo) GetAllBadges(ctx context.Context) ([]domain.Badge, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	return nil, nil
}
func (m *mockConcurrentUserRepo) AwardBadge(ctx context.Context, userID, badgeID uuid.UUID) error {
	return nil
}
func (m *mockConcurrentUserRepo) GetBadgeByCode(ctx context.Context, code string) (*domain.Badge, error) {
	return nil, nil
}


type mockConcurrentRedisCache struct {
	mu sync.Mutex
	xp map[uuid.UUID]int
}

func (m *mockConcurrentRedisCache) IncrementUserXP(ctx context.Context, userID uuid.UUID, xp int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.xp == nil {
		m.xp = make(map[uuid.UUID]int)
	}
	m.xp[userID] += xp
	return nil
}

// TestTaskUseCase_SubmitAnswer_ConcurrentSubmissionsAndPanics проверяет стабильность при множественных параллельных отправках с паниками в консенсусе
func TestTaskUseCase_SubmitAnswer_ConcurrentSubmissionsAndPanics(t *testing.T) {
	cfg := &config.Config{
		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		ConsensusMinVotes:    3,
	}

	ftID := uuid.New()
	contentRepo := &mockConcurrentContentRepo{
		ft: &domain.FragmentTranslation{
			ID:                     ftID,
			InitialAutoTranslation: "Сөт — файҙалы эсемлек.",
			Status:                 domain.StatusPending,
		},
	}

	subRepo := &mockConcurrentSubRepo{
		submitted:      make(map[string]bool),
		panicConsensus: true, // В фоне консенсус паникует
	}

	userRepo := &mockConcurrentUserRepo{}
	redisCache := &mockConcurrentRedisCache{
		xp: make(map[uuid.UUID]int),
	}

	gamifyUC := usecase.NewGamificationUseCase(userRepo, subRepo, redisCache, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, subRepo, userRepo, gamifyUC, consensusUC)

	const numGoroutines = 40
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			uid := uuid.New()
			req := domain.SubmitAnswerReq{
				UserID:                uid,
				FragmentTranslationID: ftID,
				Action:                domain.ActionConfirmed,
				SubmittedText:         "confirmed",
				TimeSpentSeconds:      5,
			}

			res, err := taskUC.SubmitAnswer(context.Background(), req)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d failed: %w", idx, err)
				return
			}
			if res.XPEarned != 10 {
				errCh <- fmt.Errorf("goroutine %d: expected 10 XP, got %d", idx, res.XPEarned)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("ошибка при конкурентной отправке: %v", err)
	}

	// Даем время фоновым горутинам консенсуса отработать и безопасно восстановиться после паники
	time.Sleep(100 * time.Millisecond)

	subRepo.mu.Lock()
	defer subRepo.mu.Unlock()
	if len(subRepo.createdSubs) != numGoroutines {
		t.Errorf("ожидалось %d сохраненных сабмитов, получено %d", numGoroutines, len(subRepo.createdSubs))
	}
}
