package usecase_test

import (
	"context"
	"strings"
	"testing"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
)

type mockContentRepoForConsensus struct {
	trans      map[uuid.UUID]*domain.FragmentTranslation
	fragments  map[uuid.UUID]*domain.Fragment
	lastStatus domain.TranslationStatus
}

func (m *mockContentRepoForConsensus) GetDailyFragmentsForUser(ctx context.Context, userID uuid.UUID, langPair string, limit int) ([]domain.TaskDTO, error) {
	return nil, nil
}
func (m *mockContentRepoForConsensus) GetTranslationByID(ctx context.Context, id uuid.UUID) (*domain.FragmentTranslation, error) {
	t, ok := m.trans[id]
	if !ok {
		return nil, domain.ErrFragmentNotFound
	}
	return t, nil
}
func (m *mockContentRepoForConsensus) UpdateTranslationStatus(ctx context.Context, id uuid.UUID, status domain.TranslationStatus, confirmedText string, score float64, count int) error {
	m.lastStatus = status
	if t, ok := m.trans[id]; ok {
		t.Status = status
	}
	return nil
}
func (m *mockContentRepoForConsensus) GetFragmentByID(ctx context.Context, id uuid.UUID) (*domain.Fragment, error) {
	f, ok := m.fragments[id]
	if !ok {
		return nil, domain.ErrFragmentNotFound
	}
	return f, nil
}
func (m *mockContentRepoForConsensus) GetConfirmedTranslations(ctx context.Context, langPair string) ([]domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForConsensus) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return nil, nil
}
func (m *mockContentRepoForConsensus) ApproveExpertReview(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return nil, nil
}
func (m *mockContentRepoForConsensus) RejectExpertReview(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockContentRepoForConsensus) GetTranslationByFragmentAndLang(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	for _, t := range m.trans {
		if t.FragmentID == fragmentID && (langPair == "" || t.LangPair == langPair) {
			return t, nil
		}
	}
	return nil, domain.ErrFragmentNotFound
}



type mockSubRepoForConsensus struct {
	groups         []domain.SubmissionGroup
	totalVotes     int
	savedConfirmed *domain.ConfirmedTranslation
}

func (m *mockSubRepoForConsensus) HasUserSubmitted(ctx context.Context, userID, ftID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockSubRepoForConsensus) CreateSubmission(ctx context.Context, sub *domain.UserSubmission) error {
	return nil
}
func (m *mockSubRepoForConsensus) CountTodaySubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockSubRepoForConsensus) CountUserSubmissions(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockSubRepoForConsensus) GetSubmissionsGroupedByText(ctx context.Context, ftID uuid.UUID) ([]domain.SubmissionGroup, int, error) {
	return m.groups, m.totalVotes, nil
}
func (m *mockSubRepoForConsensus) SaveConfirmedTranslation(ctx context.Context, ct *domain.ConfirmedTranslation) error {
	m.savedConfirmed = ct
	return nil
}

func TestConsensus_SuccessMajority(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()
	fragID := uuid.New()

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:                     ftID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				InitialAutoTranslation: "Башкортостан — наш родной край.",
				Status:                 domain.StatusPending,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {
				ID:          fragID,
				TextBa:      "Башҡортостан — беҙҙең тыуған төйәгебеҙ.",
				ChapterInfo: nil,
			},
		},
	}

	// 3 человека проголосовали за один и тот же перевод из 3 проголосовавших (100% согласия)
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "башкортостан — наш родной край.", Count: 3},
		},
		totalVotes: 3,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if contentRepo.lastStatus != domain.StatusConfirmed {
		t.Errorf("ожидался статус confirmed, получен %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed == nil {
		t.Fatalf("запись не сохранена в confirmed_translations")
	}
	if !strings.Contains(subRepo.savedConfirmed.Translation, "Башкортостан") && !strings.Contains(subRepo.savedConfirmed.Translation, "башкортостан") {
		t.Errorf("некорректный текст сохраненного перевода: %s", subRepo.savedConfirmed.Translation)
	}
}

func TestConsensus_HoneypotTrapProtection(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()
	fragID := uuid.New()
	fakeTrap := "башкортостан — наша чужая страна."

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:                     ftID,
				FragmentID:             fragID,
				HasHoneypot:            true,
				HoneypotVariant:        &fakeTrap,
				InitialAutoTranslation: "Башкортостан — наш родной край.",
				Status:                 domain.StatusPending,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {ID: fragID, TextBa: "Башҡортостан — беҙҙең тыуған төйәгебеҙ."},
		},
	}

	// 3 невнимательных пользователя одобрили ловушку (100% голосов за fakeTrap)
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: fakeTrap, Count: 3},
		},
		totalVotes: 3,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	// Ловушка НЕ должна попадать в confirmed, а должна отправляться экспертам!
	if contentRepo.lastStatus != domain.StatusExpertReview {
		t.Errorf("ловушка должна была уйти в expert_review, получен статус %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed != nil {
		t.Errorf("ошибочный перевод ни в коем случае не должен сохраняться в confirmed_translations!")
	}
}

func TestConsensus_DisputeEscalation(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:     ftID,
				Status: domain.StatusPending,
			},
		},
	}

	// 5 пользователей дали 5 разных вариантов перевода (разброс 100%, топ вариант набрал всего 20%)
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "вариант 1", Count: 1},
			{Text: "вариант 2", Count: 1},
			{Text: "вариант 3", Count: 1},
			{Text: "вариант 4", Count: 1},
			{Text: "вариант 5", Count: 1},
		},
		totalVotes: 5,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if contentRepo.lastStatus != domain.StatusExpertReview {
		t.Errorf("спорный случай должен был эскалироваться в expert_review, получен %s", contentRepo.lastStatus)
	}
}

func TestConsensus_ReportedErrorProtection(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()
	fragID := uuid.New()

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:                     ftID,
				FragmentID:             fragID,
				InitialAutoTranslation: "Неверный машинный перевод",
				Status:                 domain.StatusPending,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {ID: fragID, TextBa: "Дөрөҫ түгел тәржемә"},
		},
	}

	// 3 пользователя сообщили об ошибке ([ERROR_REPORTED])
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "[ERROR_REPORTED]", Count: 3},
		},
		totalVotes: 3,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	// Ошибочный перевод должен быть передан экспертам, а не в базу подтвержденных!
	if contentRepo.lastStatus != domain.StatusExpertReview {
		t.Errorf("жалобы на ошибку должны отправлять в expert_review, получен статус %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed != nil {
		t.Errorf("[ERROR_REPORTED] ни в коем случае не должен сохраняться в confirmed_translations!")
	}
}

// TestConsensus_FuzzyPunctuationConsensus проверяет требование 4.4 ТЗ:
// Объединение вариантов с разной пунктуацией, дефисами/тире и пробелами
// ("Книга — источник знаний.", "Книга - источник знаний", "Книга источник знаний!")
// в один консенсусный кластер подтвержденного перевода.
func TestConsensus_FuzzyPunctuationConsensus(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()
	fragID := uuid.New()

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:                     ftID,
				FragmentID:             fragID,
				LangPair:               "ba-ru",
				InitialAutoTranslation: "Книга — источник знаний.",
				Status:                 domain.StatusPending,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {
				ID:     fragID,
				TextBa: "Китап — белем шишмәһе.",
			},
		},
	}

	// 3 пользователя ввели схожие переводы с разной пунктуацией и дефисами/тире:
	// "Книга — источник знаний." (с тире и точкой)
	// "Книга - источник знаний" (с дефисом без точки)
	// "Книга источник знаний!" (без тире с восклицательным знаком)
	// Без нечеткого консенсуса голоса разделились бы 1-1-1 и консенсус не был бы достигнут.
	// С нечетким консенсусом варианты объединяются в единый кластер с 3 голосами (100% согласия).
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "Книга — источник знаний.", Count: 1},
			{Text: "Книга - источник знаний", Count: 1},
			{Text: "Книга источник знаний!", Count: 1},
		},
		totalVotes: 3,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if contentRepo.lastStatus != domain.StatusConfirmed {
		t.Errorf("ожидался статус confirmed благодаря нечеткому консенсусу, получен %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed == nil {
		t.Fatalf("запись не сохранена в confirmed_translations")
	}
	if subRepo.savedConfirmed.ConfirmedCount != 3 {
		t.Errorf("ожидалось 3 подтверждения в кластере, получено %d", subRepo.savedConfirmed.ConfirmedCount)
	}
	if !strings.Contains(subRepo.savedConfirmed.Translation, "источник знаний") {
		t.Errorf("некорректный текст сохраненного перевода: %s", subRepo.savedConfirmed.Translation)
	}
}

// TestConsensus_ConflictingTranslations_ExpertReview проверяет требование 4.4 ТЗ:
// Спорные случаи (конфликтующие переводы, отсутствие явного большинства при >= 5 ответах)
// направляются экспертам УФИЦ РАН на решение (StatusExpertReview).
func TestConsensus_ConflictingTranslations_ExpertReview(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:     ftID,
				Status: domain.StatusPending,
			},
		},
	}

	// 5 пользователей предложили 3 действительно конфликтующих перевода
	// "Книга — источник знаний." (2 голоса = 40%)
	// "Собака бежит по улице." (2 голоса = 40%)
	// "Сегодня хорошая погода." (1 голос = 20%)
	// Ни один кластер не набрал 60% большинства при >= 5 голосах -> эскалация экспертам УФИЦ РАН
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "Книга — источник знаний.", Count: 2},
			{Text: "Собака бежит по улице.", Count: 2},
			{Text: "Сегодня хорошая погода.", Count: 1},
		},
		totalVotes: 5,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	if contentRepo.lastStatus != domain.StatusExpertReview {
		t.Errorf("конфликтующие варианты должны эскалироваться в expert_review, получен %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed != nil {
		t.Errorf("при отсутствии консенсуса запись не должна сохраняться в confirmed_translations")
	}
}

// TestConsensus_FuzzyHoneypotProtection проверяет защиту от ловушек:
// Даже если большинство одобрило вариант ловушки с вариацией дефиса/знаков,
// он распознается нечетким сравнением и эскалируется в expert_review.
func TestConsensus_FuzzyHoneypotProtection(t *testing.T) {
	cfg := &config.Config{ConsensusMinVotes: 3}
	ftID := uuid.New()
	fragID := uuid.New()
	honeypotText := "Башкортостан — наша чужая страна."

	contentRepo := &mockContentRepoForConsensus{
		trans: map[uuid.UUID]*domain.FragmentTranslation{
			ftID: {
				ID:              ftID,
				FragmentID:      fragID,
				HasHoneypot:     true,
				HoneypotVariant: &honeypotText,
				Status:          domain.StatusPending,
			},
		},
		fragments: map[uuid.UUID]*domain.Fragment{
			fragID: {ID: fragID, TextBa: "Башҡортостан — беҙҙең тыуған төйәгебеҙ."},
		},
	}

	// 3 пользователя одобрили ловушку, но с мелкой вариацией пунктуации/дефиса:
	// "башкортостан - наша чужая страна" (дефис вместо тире, без точки)
	subRepo := &mockSubRepoForConsensus{
		groups: []domain.SubmissionGroup{
			{Text: "башкортостан - наша чужая страна", Count: 2},
			{Text: "Башкортостан — наша чужая страна!", Count: 1},
		},
		totalVotes: 3,
	}

	consensusUC := usecase.NewConsensusUseCase(contentRepo, subRepo, cfg)
	err := consensusUC.ProcessConsensus(context.Background(), ftID)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	// Ловушка даже при нечетком совпадении ОБЯЗАНА отправляться экспертам, а не подтверждаться!
	if contentRepo.lastStatus != domain.StatusExpertReview {
		t.Errorf("нечеткая ловушка должна была уйти в expert_review, получен %s", contentRepo.lastStatus)
	}
	if subRepo.savedConfirmed != nil {
		t.Errorf("ловушка не должна сохраняться в confirmed_translations")
	}
}
