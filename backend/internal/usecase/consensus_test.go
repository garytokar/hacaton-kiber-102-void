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
