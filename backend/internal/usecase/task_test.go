package usecase_test

import (
	"context"
	"testing"

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

type mockSubRepoForTask struct {
	submitted       map[uuid.UUID]bool
	createdSubs     []*domain.UserSubmission
	todayCount      int
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
