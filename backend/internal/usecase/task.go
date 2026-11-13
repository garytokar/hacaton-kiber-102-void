package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/moderation"

	"github.com/google/uuid"
)

type TaskUseCase struct {
	contentRepo    domain.ContentRepository
	submissionRepo domain.SubmissionRepository
	userRepo       domain.UserRepository
	gamifyUC       *GamificationUseCase
	consensusUC    *ConsensusUseCase
}

func NewTaskUseCase(
	contentRepo domain.ContentRepository,
	submissionRepo domain.SubmissionRepository,
	userRepo domain.UserRepository,
	gamifyUC *GamificationUseCase,
	consensusUC *ConsensusUseCase,
) *TaskUseCase {
	return &TaskUseCase{
		contentRepo:    contentRepo,
		submissionRepo: submissionRepo,
		userRepo:       userRepo,
		gamifyUC:       gamifyUC,
		consensusUC:    consensusUC,
	}
}

// GetDailyTasks возвращает 5 ежедневных карточек для тренировки
func (uc *TaskUseCase) GetDailyTasks(ctx context.Context, userID uuid.UUID, langPair string) (*domain.DailyTasksResponse, error) {
	if langPair == "" && userID != uuid.Nil && uc.userRepo != nil {
		if u, err := uc.userRepo.FindByID(ctx, userID); err == nil && u != nil && u.PreferredLangPair != "" {
			langPair = u.PreferredLangPair
		}
	}
	if langPair == "" {
		langPair = "ba-ru"
	}

	tasks, err := uc.contentRepo.GetDailyFragmentsForUser(ctx, userID, langPair, 5)
	if err != nil {
		return nil, err
	}

	return &domain.DailyTasksResponse{
		Date:  time.Now().Format("2006-01-02"),
		Count: len(tasks),
		Tasks: tasks,
	}, nil
}

// SubmitAnswer единый метод сохранения ответа (вызывается Веб-клиентом и Ботом)
func (uc *TaskUseCase) SubmitAnswer(ctx context.Context, req domain.SubmitAnswerReq) (*domain.SubmitResultDTO, error) {
	// 1. Проверяем, не давал ли пользователь ответ на этот фрагмент
	alreadySubmitted, err := uc.submissionRepo.HasUserSubmitted(ctx, req.UserID, req.FragmentTranslationID)
	if err != nil {
		return nil, err
	}
	if alreadySubmitted {
		return nil, domain.ErrAlreadySubmitted
	}

	// 2. Получаем запись перевода
	ft, err := uc.contentRepo.GetTranslationByID(ctx, req.FragmentTranslationID)
	if err != nil {
		return nil, err
	}

	// 3. Получаем пользователя
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// 4. Подготовка текста перевода и модерация ненормативной лексики / спама
	submittedText := strings.TrimSpace(req.SubmittedText)
	if req.Action == domain.ActionReportedError {
		submittedText = "[ERROR_REPORTED]"
	} else if req.Action == domain.ActionConfirmed || submittedText == "" || submittedText == "confirmed" {
		if ft.HasHoneypot && ft.HoneypotVariant != nil {
			submittedText = *ft.HoneypotVariant
		} else {
			submittedText = ft.InitialAutoTranslation
		}
	} else if req.Action == domain.ActionEdited {
		// Пользователь предложил свой вариант: проверяем на мат и бессмыслицу
		valid, _ := moderation.ValidateTranslationQuality(submittedText)
		if !valid {
			if moderation.ContainsProfanity(submittedText) {
				return nil, domain.ErrProfanityDetected
			}
			return nil, domain.ErrInvalidTranslationText
		}
	}

	// 5. Проверяем разоблачение honeypot-ловушки
	// Если в фрагменте была намеренная ошибка, а пользователь исправил ее или пожаловался
	isHoneypotCaught := false
	if ft.HasHoneypot {
		if req.Action == domain.ActionEdited || req.Action == domain.ActionReportedError {
			// Пользователь не поверил ложной подсказке!
			isHoneypotCaught = true
		}
	}

	// 6. Начисляем опыт и обновляем стрик
	reward, err := uc.gamifyUC.AwardSubmissionXP(ctx, user, req.Action, isHoneypotCaught)
	if err != nil {
		return nil, err
	}

	// 7. Сохраняем сабмит пользователя с реальным текстом перевода
	sub := &domain.UserSubmission{
		ID:                    uuid.New(),
		UserID:                req.UserID,
		FragmentTranslationID: req.FragmentTranslationID,
		Action:                req.Action,
		SubmittedText:         submittedText,
		TimeSpentSeconds:      req.TimeSpentSeconds,
		DetectedHoneypot:      isHoneypotCaught,
		XPAwarded:             reward.XPEarned,
		CreatedAt:             time.Now(),
	}

	if err := uc.submissionRepo.CreateSubmission(ctx, sub); err != nil {
		return nil, err
	}

	// 8. Запускаем проверку консенсуса в фоне (горутина) с защитой от паники
	go func(translationID uuid.UUID) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("consensus background task panicked",
					"translation_id", translationID.String(),
					"panic", fmt.Sprint(r),
					"stack", string(debug.Stack()),
				)
			}
		}()

		if uc.consensusUC == nil {
			return
		}

		if err := uc.consensusUC.ProcessConsensus(context.Background(), translationID); err != nil {
			slog.Error("consensus background task failed",
				"translation_id", translationID.String(),
				"error", err,
			)
		}
	}(req.FragmentTranslationID)

	return &domain.SubmitResultDTO{
		XPEarned:         reward.XPEarned,
		TotalXP:          reward.TotalXP,
		CurrentStreak:    reward.CurrentStreak,
		IsHoneypotCaught: isHoneypotCaught,
		IsDailyCompleted: reward.IsDailyCompleted,
	}, nil
}


// GetTranslationByFragment возвращает запись автоперевода по ID фрагмента и языковой паре
func (uc *TaskUseCase) GetTranslationByFragment(ctx context.Context, fragmentID uuid.UUID, langPair string) (*domain.FragmentTranslation, error) {
	return uc.contentRepo.GetTranslationByFragmentAndLang(ctx, fragmentID, langPair)
}
