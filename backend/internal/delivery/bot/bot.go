package bot

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/repository/redis"
	"bashkir-backend/internal/usecase"

	tele "gopkg.in/telebot.v3"
)

type BotService struct {
	bot      *tele.Bot
	handlers *BotHandlers
	kb       *Keyboards
	cfg      *config.Config
}

func NewBotService(
	cfg *config.Config,
	authUC *usecase.AuthUseCase,
	taskUC *usecase.TaskUseCase,
	gamifyUC *usecase.GamificationUseCase,
	lmsUC *usecase.LMSUseCase,
	redisRepo *redis.LeaderboardRedisRepo,
) (*BotService, error) {
	if cfg.TelegramBotToken == "" || cfg.TelegramBotToken == "mock_token" {
		log.Println("[Telegram Bot] Токен не задан. Бот работает в холостом режиме.")
		return nil, nil
	}

	pref := tele.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	kb := NewKeyboards(cfg.WebAppURL)
	handlers := NewBotHandlers(authUC, taskUC, gamifyUC, lmsUC, redisRepo, kb)

	svc := &BotService{
		bot:      b,
		handlers: handlers,
		kb:       kb,
		cfg:      cfg,
	}

	svc.registerRoutes()
	return svc, nil
}

func (s *BotService) registerRoutes() {
	b := s.bot

	// Команды
	b.Handle("/start", s.handlers.HandleStart)
	b.Handle("/daily", s.handlers.HandleDaily)
	b.Handle("/stats", s.handlers.HandleStats)
	b.Handle("/rating", s.handlers.HandleRating)
	b.Handle("/lang", s.handlers.HandleLang)
	b.Handle("/cancel", s.handlers.HandleCancel)

	// Текстовые кнопки меню
	b.Handle("📝 Тренировка в чате", s.handlers.HandleDaily)
	b.Handle("📊 Моя статистика", s.handlers.HandleStats)
	b.Handle("🏆 Рейтинг школ", s.handlers.HandleRating)
	b.Handle("🌐 Выбор языка", s.handlers.HandleLang)

	// Inline-кнопки управления карточкой
	b.Handle(&s.kb.BtnConfirm, s.handlers.HandleConfirm)
	b.Handle(&s.kb.BtnEdit, s.handlers.HandleEditRequest)
	b.Handle(&s.kb.BtnReport, s.handlers.HandleReport)
	b.Handle(&s.kb.BtnCancel, s.handlers.HandleCancel)

	// Inline-кнопки выбора языковой пары
	b.Handle(&s.kb.BtnLangBaRu, s.handlers.HandleSetLangBaRu)
	b.Handle(&s.kb.BtnLangBaKk, s.handlers.HandleSetLangBaKk)
	b.Handle(&s.kb.BtnLangBaEn, s.handlers.HandleSetLangBaEn)

	// Текстовые сообщения для FSM ввода исправлений
	b.Handle(tele.OnText, s.handlers.HandleTextMessage)
}

func (s *BotService) Start() {
	log.Println("[Telegram Bot] Запуск в режиме Long Polling...")
	s.bot.Start()
}

func (s *BotService) Stop() {
	if s.bot != nil {
		s.bot.Stop()
	}
}

func (s *BotService) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u tele.Update
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		s.bot.ProcessUpdate(u)
	}
}
