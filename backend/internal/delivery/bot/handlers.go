package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/repository/redis"
	"bashkir-backend/internal/usecase"

	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
)

type BotHandlers struct {
	authUC    *usecase.AuthUseCase
	taskUC    *usecase.TaskUseCase
	gamifyUC  *usecase.GamificationUseCase
	lmsUC     *usecase.LMSUseCase
	redisRepo *redis.LeaderboardRedisRepo
	kb        *Keyboards
}

func NewBotHandlers(
	authUC *usecase.AuthUseCase,
	taskUC *usecase.TaskUseCase,
	gamifyUC *usecase.GamificationUseCase,
	lmsUC *usecase.LMSUseCase,
	redisRepo *redis.LeaderboardRedisRepo,
	kb *Keyboards,
) *BotHandlers {
	return &BotHandlers{
		authUC:    authUC,
		taskUC:    taskUC,
		gamifyUC:  gamifyUC,
		lmsUC:     lmsUC,
		redisRepo: redisRepo,
		kb:        kb,
	}
}

func (h *BotHandlers) HandleStart(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	user, err := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, fmt.Sprintf("%s %s", sender.FirstName, sender.LastName))
	if err != nil {
		return c.Send("Ошибка регистрации: " + err.Error())
	}

	msg := fmt.Sprintf(
		"Һаумыһығыҙ, %s! 👋\n\n"+
			"Добро пожаловать на платформу **«Башкирская языковая школа»**!\n\n"+
			"Здесь ты можешь тренировать башкирский язык, проверять переводы, соревноваться с одноклассниками и помогать создавать базу знаний родного языка.\n\n"+
			"Твой опыт: ⭐ **%d XP**\n"+
			"Стрик: 🔥 **%d дней**\n\n"+
			"Нажми кнопку ниже, чтобы запустить интерактивное приложение или тренироваться прямо здесь!",
		user.FullName, user.TotalXP, user.CurrentStreak,
	)

	return c.Send(msg, h.kb.MainMenu, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleDaily(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	user, err := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")
	if err != nil {
		return c.Send("Ошибка профиля: " + err.Error())
	}

	langPair := user.PreferredLangPair
	if langPair == "" {
		langPair = "ba-ru"
	}

	dailyResp, err := h.taskUC.GetDailyTasks(ctx, user.ID, langPair)
	if err != nil || len(dailyResp.Tasks) == 0 {
		return c.Send("🎉 Ты выполнил все доступные задания на сегодня! Возвращайся завтра или открой приложение для просмотра рейтинга.")
	}

	firstTask := dailyResp.Tasks[0]
	// Сохраняем текущий task_id в сессию Redis
	_ = h.redisRepo.SetBotFSMState(ctx, sender.ID, "reviewing", firstTask.FragmentTranslationID.String(), 30*time.Minute)

	text := fmt.Sprintf(
		"📖 **Карточка задания (1 из %d)**\n\n"+
			"🔹 **Башкирский текст:**\n«%s»\n\n"+
			"🔸 **Предложенный перевод:**\n«%s»\n\n"+
			"📍 _Источник: %s (%s)_\n\n"+
			"Твоя задача — подтвердить правильность или предложить более точный вариант:",
		len(dailyResp.Tasks),
		firstTask.TextBa,
		firstTask.SuggestedTranslation,
		firstTask.SourceTitle,
		firstTask.SourceCoordinates,
	)

	return c.Send(text, h.kb.TaskControls, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleConfirm(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	state, transIDStr, _ := h.redisRepo.GetBotFSMState(ctx, sender.ID)
	if state == "" || transIDStr == "" {
		return c.Send("Задание устарело. Нажми «📝 Тренировка в чате» для нового задания.")
	}

	transID, err := uuid.Parse(transIDStr)
	if err != nil {
		return c.Send("Некорректный ID задания.")
	}

	user, _ := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")

	req := domain.SubmitAnswerReq{
		UserID:                user.ID,
		FragmentTranslationID: transID,
		Action:                domain.ActionConfirmed,
		SubmittedText:         "confirmed",
		TimeSpentSeconds:      5,
	}

	result, err := h.taskUC.SubmitAnswer(ctx, req)
	if err != nil {
		return c.Send("Ошибка сохранения ответа: " + err.Error())
	}

	_ = h.redisRepo.ClearBotFSM(ctx, sender.ID)

	bonusMsg := ""
	if result.IsHoneypotCaught {
		bonusMsg = "\n🎯 **Ты распознал контрольную ловушку!** (+30 бонусных XP)"
	}
	if result.IsDailyCompleted {
		bonusMsg += "\n🔥 **Дневная норма выполнена!** (+50 бонусных XP за стрик)"
	}

	reply := fmt.Sprintf(
		"✅ **Ответ принят!**\n\n"+
			"Начислено: +%d XP ⭐\n"+
			"Твой опыт: %d XP\n"+
			"Стрик: %d дн. 🔥%s\n\n"+
			"Перейти к следующему заданию: /daily",
		result.XPEarned, result.TotalXP, result.CurrentStreak, bonusMsg,
	)

	return c.Send(reply, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleReport(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	state, transIDStr, _ := h.redisRepo.GetBotFSMState(ctx, sender.ID)
	if state == "" || transIDStr == "" {
		return c.Send("Задание устарело. Нажми «📝 Тренировка в чате» для нового задания.")
	}

	transID, err := uuid.Parse(transIDStr)
	if err != nil {
		return c.Send("Некорректный ID задания.")
	}

	user, _ := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")

	req := domain.SubmitAnswerReq{
		UserID:                user.ID,
		FragmentTranslationID: transID,
		Action:                domain.ActionReportedError,
		SubmittedText:         "[ERROR_REPORTED]",
		TimeSpentSeconds:      5,
	}

	result, err := h.taskUC.SubmitAnswer(ctx, req)
	if err != nil {
		return c.Send("Ошибка сохранения ответа: " + err.Error())
	}

	_ = h.redisRepo.ClearBotFSM(ctx, sender.ID)

	bonusMsg := ""
	if result.IsHoneypotCaught {
		bonusMsg = "\n🎯 **Ты разоблачил контрольную ловушку!** (+30 бонусных XP)"
	}
	if result.IsDailyCompleted {
		bonusMsg += "\n🔥 **Дневная норма выполнена!** (+50 бонусных XP за стрик)"
	}

	reply := fmt.Sprintf(
		"⚠️ **Сообщение об ошибке принято!**\n\n"+
			"Материал будет направлен на экспертизу в УФИЦ РАН.\n"+
			"Начислено: +%d XP ⭐\n"+
			"Твой опыт: %d XP\n"+
			"Стрик: %d дн. 🔥%s\n\n"+
			"Следующее задание: /daily",
		result.XPEarned, result.TotalXP, result.CurrentStreak, bonusMsg,
	)

	return c.Send(reply, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleEditRequest(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	state, transIDStr, _ := h.redisRepo.GetBotFSMState(ctx, sender.ID)
	if state == "" || transIDStr == "" {
		return c.Send("Нет активного задания. Нажми «📝 Тренировка в чате».")
	}

	// Переводим FSM в ожидание текстового ввода
	_ = h.redisRepo.SetBotFSMState(ctx, sender.ID, "waiting_text", transIDStr, 15*time.Minute)

	return c.Send("✏️ Введи и отправь сообщением твой правильный вариант перевода на русский язык:")
}

func (h *BotHandlers) HandleTextMessage(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	state, transIDStr, _ := h.redisRepo.GetBotFSMState(ctx, sender.ID)
	if state != "waiting_text" || transIDStr == "" {
		// Обычное текстовое сообщение вне контекста исправления
		return nil
	}

	transID, err := uuid.Parse(transIDStr)
	if err != nil {
		return nil
	}

	user, _ := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")
	correctionText := strings.TrimSpace(c.Text())

	req := domain.SubmitAnswerReq{
		UserID:                user.ID,
		FragmentTranslationID: transID,
		Action:                domain.ActionEdited,
		SubmittedText:         correctionText,
		TimeSpentSeconds:      15,
	}

	result, err := h.taskUC.SubmitAnswer(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrProfanityDetected) {
			return c.Send("🚫 **Внимание:** В предложенном переводе обнаружена ненормативная лексика!\n\nПлатформа предназначена для школьников и изучения языка. Пожалуйста, отправь корректный перевод:", tele.ModeMarkdown)
		}
		if errors.Is(err, domain.ErrInvalidTranslationText) {
			return c.Send("⚠️ **Некорректный перевод:** Текст слишком короткий или бессмысленный.\n\nПожалуйста, отправь осмысленный вариант перевода:", tele.ModeMarkdown)
		}
		return c.Send("Ошибка: " + err.Error())
	}

	_ = h.redisRepo.ClearBotFSM(ctx, sender.ID)

	reply := fmt.Sprintf(
		"👏 **Твое исправление сохранено!**\n\n"+
			"«%s»\n\n"+
			"Начислено: +%d XP ⭐\n"+
			"Твой опыт: %d XP\n\n"+
			"Следующее задание: /daily",
		correctionText, result.XPEarned, result.TotalXP,
	)

	return c.Send(reply, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleStats(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	user, err := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")
	if err != nil {
		return c.Send("Ошибка профиля.")
	}

	msg := fmt.Sprintf(
		"📊 **Личный профиль**\n\n"+
			"👤 Имя: %s\n"+
			"⭐ Опыт: **%d XP**\n"+
			"🔥 Текущий стрик: **%d дней подряд**\n"+
			"🏆 Рекордный стрик: **%d дней**\n\n"+
			"Продолжай каждый день, чтобы не потерять стрик!",
		user.FullName, user.TotalXP, user.CurrentStreak, user.MaxStreak,
	)

	return c.Send(msg, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleRating(c tele.Context) error {
	ctx := context.Background()
	schools, err := h.lmsUC.GetSchoolsLeaderboard(ctx, 10)
	if err != nil || len(schools) == 0 {
		return c.Send("Рейтинг пока формируется. Будь первым, кто принесет очки своей школе!")
	}

	var sb strings.Builder
	sb.WriteString("🏆 **Топ-10 школ по активности:**\n\n")

	medals := []string{"🥇", "🥈", "🥉"}
	for i, s := range schools {
		medal := fmt.Sprintf("%d.", i+1)
		if i < len(medals) {
			medal = medals[i]
		}
		sb.WriteString(fmt.Sprintf("%s **%s** (%s) — %d XP (%d учеников)\n", medal, s.SchoolName, s.City, s.TotalXP, s.ActiveStudents))
	}

	return c.Send(sb.String(), tele.ModeMarkdown)
}

func (h *BotHandlers) HandleCancel(c tele.Context) error {
	ctx := context.Background()
	_ = h.redisRepo.ClearBotFSM(ctx, c.Sender().ID)
	return c.Send("Тренировка остановлена. Возвращайся в любое время!", h.kb.MainMenu)
}

func (h *BotHandlers) HandleLang(c tele.Context) error {
	sender := c.Sender()
	ctx := context.Background()

	user, err := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")
	if err != nil {
		return c.Send("Ошибка профиля: " + err.Error())
	}

	payload := ""
	if c.Message() != nil {
		payload = strings.TrimSpace(c.Message().Payload)
	}

	if payload != "" {
		payload = strings.ToLower(payload)
		if payload == "ba-ru" || payload == "ba-kk" || payload == "ba-en" {
			_ = h.authUC.UpdatePreferredLangPair(ctx, user.ID, payload)
			return c.Send(fmt.Sprintf("✅ Языковая пара успешно изменена на **%s**!\nТеперь тренировки будут использовать этот язык: /daily", payload), tele.ModeMarkdown)
		}
	}

	curLang := user.PreferredLangPair
	if curLang == "" {
		curLang = "ba-ru"
	}

	msg := fmt.Sprintf(
		"🌐 **Выбор языковой пары (Раздел 4.1 ТЗ)**\n\n"+
			"Текущая языковая пара: **%s**\n\n"+
			"Выбери направление перевода для тренировок:",
		curLang,
	)
	return c.Send(msg, h.kb.LangControls, tele.ModeMarkdown)
}

func (h *BotHandlers) HandleSetLangBaRu(c tele.Context) error {
	return h.setLangAndReply(c, "ba-ru")
}

func (h *BotHandlers) HandleSetLangBaKk(c tele.Context) error {
	return h.setLangAndReply(c, "ba-kk")
}

func (h *BotHandlers) HandleSetLangBaEn(c tele.Context) error {
	return h.setLangAndReply(c, "ba-en")
}

func (h *BotHandlers) setLangAndReply(c tele.Context, lang string) error {
	sender := c.Sender()
	ctx := context.Background()

	user, err := h.authUC.GetOrCreateTgUser(ctx, sender.ID, sender.Username, "")
	if err != nil {
		return c.Send("Ошибка: " + err.Error())
	}

	_ = h.authUC.UpdatePreferredLangPair(ctx, user.ID, lang)
	return c.Send(fmt.Sprintf("✅ Языковая пара изменена на **%s**!\nТеперь ежедневные задания будут выдаваться для этой пары: /daily", lang), tele.ModeMarkdown)
}

