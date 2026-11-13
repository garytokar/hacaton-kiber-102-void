package http

import (
	"net/http"

	"bashkir-backend/internal/config"
	"bashkir-backend/internal/delivery/http/middleware"
	v1 "bashkir-backend/internal/delivery/http/v1"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type RouterParams struct {
	Cfg            *config.Config
	AuthHandler    *v1.AuthHandler
	TaskHandler    *v1.TaskHandler
	LeaderboardH   *v1.LeaderboardHandler
	LMSHandler     *v1.LMSHandler
	ExportHandler  *v1.ExportHandler
	WebhookHandler http.HandlerFunc
}

func NewRouter(p RouterParams) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Cors())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"bashkir-school-backend"}`))
	})

	// Webhook Telegram бота (если включен режим webhook)
	if p.WebhookHandler != nil {
		r.Post("/api/v1/telegram/webhook", p.WebhookHandler)
	}

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", p.AuthHandler.Register)
			r.Post("/login", p.AuthHandler.Login)
			r.Post("/telegram-webapp", p.AuthHandler.TelegramWebAppAuth)

			// Protected /me
			r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me", p.AuthHandler.Me)
		})

		// Public Leaderboards
		r.Route("/leaderboards", func(r chi.Router) {
			r.Get("/users", p.LeaderboardH.GetUsers)
			r.Get("/schools", p.LeaderboardH.GetSchools)
		})

		// Public / Export (Требование раздела 5-6 ТЗ)
		r.Route("/export", func(r chi.Router) {
			r.Get("/translations", p.ExportHandler.ExportTranslations)
		})

		// Protected Tasks (Ученики)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(p.Cfg.JWTSecret))

			r.Get("/tasks/daily", p.TaskHandler.GetDailyTasks)
			r.Post("/tasks/submit", p.TaskHandler.SubmitTask)
		})

		// Protected LMS (Учителя)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(p.Cfg.JWTSecret))

			r.Get("/teacher/classes/{class_id}", p.LMSHandler.GetClassStats)
		})
	})

	return r
}
