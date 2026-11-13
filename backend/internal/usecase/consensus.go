package usecase

import (
	"context"
	"fmt"
	"strings"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/fuzzy"

	"github.com/google/uuid"
)

type ConsensusUseCase struct {
	contentRepo    domain.ContentRepository
	submissionRepo domain.SubmissionRepository
	cfg            *config.Config
}

func NewConsensusUseCase(
	contentRepo domain.ContentRepository,
	submissionRepo domain.SubmissionRepository,
	cfg *config.Config,
) *ConsensusUseCase {
	return &ConsensusUseCase{
		contentRepo:    contentRepo,
		submissionRepo: submissionRepo,
		cfg:            cfg,
	}
}

// ProcessConsensus анализирует поступившие ответы и выносит решение большинства
func (uc *ConsensusUseCase) ProcessConsensus(ctx context.Context, fragmentTranslationID uuid.UUID) error {
	ft, err := uc.contentRepo.GetTranslationByID(ctx, fragmentTranslationID)
	if err != nil {
		return err
	}

	// Если уже подтвержден или отправлен экспертам, не пересчитываем
	if ft.Status != domain.StatusPending {
		return nil
	}

	groups, totalVotes, err := uc.submissionRepo.GetSubmissionsGroupedByText(ctx, fragmentTranslationID)
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return nil
	}

	// Нечеткая кластеризация для объединения вариантов с разной пунктуацией,
	// дефисами/тире, кавычками и мелкими опечатками
	clusteredGroups := fuzzy.ClusterSubmissions(groups, fuzzy.DefaultSimilarityThreshold)
	if len(clusteredGroups) == 0 {
		return nil
	}

	topGroup := clusteredGroups[0]
	topRatio := float64(topGroup.Count) / float64(totalVotes)

	// Правило 1: Если один и тот же вариант набрал >= 3 голосов и составляет большинство (>= 60%)
	if topGroup.Count >= uc.cfg.ConsensusMinVotes && topRatio >= 0.60 {
		// Защита от ошибочных переводов: если большинство пользователей пожаловались на ошибку,
		// направляем на ручную проверку экспертам (УФИЦ РАН), а не в эталоны
		if strings.EqualFold(strings.TrimSpace(topGroup.Text), "[ERROR_REPORTED]") {
			return uc.contentRepo.UpdateTranslationStatus(ctx, ft.ID, domain.StatusExpertReview, topGroup.Text, topRatio, topGroup.Count)
		}

		// Защита от ловушек: если большинство пользователей ошибочно одобрило ловушку (honeypot),
		// мы ни в коем случае не сохраняем ошибку в базу эталонов — отправляем экспертам
		isHoneypot := ft.HasHoneypot && ft.HoneypotVariant != nil &&
			(strings.EqualFold(strings.TrimSpace(topGroup.Text), strings.TrimSpace(*ft.HoneypotVariant)) ||
				fuzzy.Similarity(topGroup.Text, *ft.HoneypotVariant) >= fuzzy.DefaultSimilarityThreshold)
		if isHoneypot {
			return uc.contentRepo.UpdateTranslationStatus(ctx, ft.ID, domain.StatusExpertReview, topGroup.Text, topRatio, topGroup.Count)
		}

		// Статус confirmed!
		err = uc.contentRepo.UpdateTranslationStatus(ctx, ft.ID, domain.StatusConfirmed, topGroup.Text, topRatio, topGroup.Count)
		if err != nil {
			return err
		}

		// Получаем данные фрагмента для записи в подтвержденную базу (Раздел 5 ТЗ)
		fragment, err := uc.contentRepo.GetFragmentByID(ctx, ft.FragmentID)
		if err == nil && fragment != nil {
			coord := "Источник материалов"
			if fragment.ChapterInfo != nil {
				coord = *fragment.ChapterInfo
			}
			if fragment.SentenceIdx != nil {
				coord = fmt.Sprintf("%s, предл. %d", coord, *fragment.SentenceIdx)
			}

			confirmedRecord := &domain.ConfirmedTranslation{
				ID:             uuid.New(),
				FragmentID:     ft.FragmentID,
				TextBa:         fragment.TextBa,
				LangPair:       ft.LangPair,
				Translation:    strings.TrimSpace(topGroup.Text),
				SourceCoord:    coord,
				ConfirmedCount: topGroup.Count,
				Status:         "confirmed",
			}
			_ = uc.submissionRepo.SaveConfirmedTranslation(ctx, confirmedRecord)
		}
		return nil
	}

	// Правило 2: Если получено >= 5 голосов, но сильный разброс мнений (нет явного большинства)
	if totalVotes >= 5 && topRatio < 0.60 {
		// Эскалация в экспертный совет (УФИЦ РАН)
		return uc.contentRepo.UpdateTranslationStatus(ctx, ft.ID, domain.StatusExpertReview, topGroup.Text, topRatio, topGroup.Count)
	}

	return nil
}

// GetReviewQueue возвращает список фрагментов, направленных на экспертизу в УФИЦ РАН
func (uc *ConsensusUseCase) GetReviewQueue(ctx context.Context) ([]domain.ExpertReviewItem, error) {
	return uc.contentRepo.GetReviewQueue(ctx)
}

// ApproveTranslation утверждает перевод экспертом УФИЦ РАН и вносит его в эталонный корпус
func (uc *ConsensusUseCase) ApproveTranslation(ctx context.Context, id uuid.UUID, approvedText string) (*domain.ConfirmedTranslation, error) {
	return uc.contentRepo.ApproveExpertReview(ctx, id, approvedText)
}

// RejectTranslation отклоняет некорректный или спамный вариант перевода
func (uc *ConsensusUseCase) RejectTranslation(ctx context.Context, id uuid.UUID) error {
	return uc.contentRepo.RejectExpertReview(ctx, id)
}

