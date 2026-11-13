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
	AdminHandler   *v1.AdminHandler
	WebhookHandler http.HandlerFunc
}

func registerAPIRoutes(r chi.Router, p RouterParams) {
	// 1. Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", p.AuthHandler.Register)
		r.Post("/login", p.AuthHandler.Login)
		r.Post("/telegram-webapp", p.AuthHandler.TelegramWebAppAuth)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me", p.AuthHandler.Me)
	})

	// 2. User routes
	r.Route("/users", func(r chi.Router) {
		r.Get("/schools", p.LMSHandler.GetSchoolsOptions)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Patch("/me", p.AuthHandler.UpdateMe)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me/activity", p.AuthHandler.UserActivity)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Post("/me/protect-streak", p.AuthHandler.ProtectStreak)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me/lang-pair", p.AuthHandler.GetLangPair)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Put("/me/lang-pair", p.AuthHandler.UpdateLangPair)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Patch("/me/lang-pair", p.AuthHandler.UpdateLangPair)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me/badges", p.LeaderboardH.GetAchievements)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Get("/me/achievements", p.LeaderboardH.GetAchievements)
	})

	// 3. Fragments & Tasks
	r.Route("/fragments", func(r chi.Router) {
		r.Get("/daily", p.TaskHandler.GetDailyFragment)
		r.Get("/", p.TaskHandler.GetDailyFragment)
		r.Get("/{id}", p.TaskHandler.GetFragmentByID)
	})

	// 4. Translations & Voting
	r.Route("/translations", func(r chi.Router) {
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Post("/", p.TaskHandler.ProposeTranslation)
		r.With(middleware.Auth(p.Cfg.JWTSecret)).Post("/{id}/vote", p.TaskHandler.VoteTranslation)
		r.Get("/{id}/history", p.TaskHandler.GetTranslationHistory)
	})

	// 5. Classic task routes (Dual-Support)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(p.Cfg.JWTSecret))
		r.Get("/tasks/daily", p.TaskHandler.GetDailyTasks)
		r.Post("/tasks/submit", p.TaskHandler.SubmitTask)
	})

	// 6. Leaderboards (Раздел 4.5 ТЗ: срезы по времени, классы, словарный запас)
	r.Route("/leaderboard", func(r chi.Router) {
		r.Get("/", p.LeaderboardH.GetUsers)
		r.Get("/users", p.LeaderboardH.GetUsers)
		r.Get("/schools", p.LeaderboardH.GetSchools)
		r.Get("/classes", p.LeaderboardH.GetClasses)
		r.Get("/vocabulary", p.LeaderboardH.GetVocabulary)
		r.Get("/teams", p.LeaderboardH.GetTeams)
		r.Get("/achievements", p.LeaderboardH.GetAchievements)
	})
	r.Route("/leaderboards", func(r chi.Router) {
		r.Get("/", p.LeaderboardH.GetUsers)
		r.Get("/users", p.LeaderboardH.GetUsers)
		r.Get("/schools", p.LeaderboardH.GetSchools)
		r.Get("/classes", p.LeaderboardH.GetClasses)
		r.Get("/vocabulary", p.LeaderboardH.GetVocabulary)
		r.Get("/achievements", p.LeaderboardH.GetAchievements)
	})

	// 7. Stats & Treasury
	r.Route("/stats", func(r chi.Router) {
		r.Get("/overview", p.ExportHandler.GetOverview)
		r.Get("/treasury", p.ExportHandler.GetTreasury)
	})
	r.Route("/export", func(r chi.Router) {
		r.Get("/translations", p.ExportHandler.ExportTranslations)
	})

	// 8. Achievements
	r.Get("/achievements", p.LeaderboardH.GetAchievements)

	// 9. Dictionary
	if p.AdminHandler != nil {
		r.Route("/dictionary", func(r chi.Router) {
			r.Get("/", p.AdminHandler.GetDictionary)
			r.Post("/", p.AdminHandler.AddDictionaryWord)
			r.Patch("/{id}/favorite", p.AdminHandler.ToggleFavoriteWord)
			r.Delete("/{id}", p.AdminHandler.DeleteDictionaryWord)
		})

		// 10. Quests
		r.Route("/quests", func(r chi.Router) {
			r.Get("/daily", p.AdminHandler.GetDailyQuests)
			r.Post("/{id}/claim", p.AdminHandler.ClaimQuest)
			r.Get("/expedition", p.AdminHandler.GetWeeklyExpedition)
		})

		// 11. Admin & Expert Review (Защищено ролью expert / admin)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(p.Cfg.JWTSecret))
			r.Use(middleware.RequireRoles("admin", "expert"))
			r.Get("/review", p.AdminHandler.GetReviewQueue)
			r.Post("/review/{id}/approve", p.AdminHandler.ApproveTranslation)
			r.Post("/review/{id}/reject", p.AdminHandler.RejectTranslation)
			r.Get("/stats", p.AdminHandler.GetAdminStats)
			r.Get("/users", p.AdminHandler.GetAdminUsers)
			r.Patch("/users/{id}/role", p.AdminHandler.UpdateUserRole)
		})
	}

	// 12. LMS teacher & class routes (Раздел 4.1 и 6 ТЗ)
	r.Route("/lms", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(p.Cfg.JWTSecret))
			r.Post("/classes", p.LMSHandler.CreateClass)
			r.Post("/classes/{id}/invite", p.LMSHandler.GenerateInviteCode)
			r.Post("/classes/join", p.LMSHandler.JoinClass)
			r.Post("/classes/{id}/students", p.LMSHandler.AddStudent)
			r.Get("/classes/{class_id}", p.LMSHandler.GetClassStats)
		})
	})
	r.Route("/teacher", func(r chi.Router) {
		r.Use(middleware.Auth(p.Cfg.JWTSecret))
		r.Post("/classes", p.LMSHandler.CreateClass)
		r.Post("/classes/{id}/invite", p.LMSHandler.GenerateInviteCode)
		r.Post("/classes/{id}/students", p.LMSHandler.AddStudent)
		r.Get("/classes/{class_id}", p.LMSHandler.GetClassStats)
	})

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

	// Webhook Telegram бота
	if p.WebhookHandler != nil {
		r.Post("/api/telegram/webhook", p.WebhookHandler)
		r.Post("/api/v1/telegram/webhook", p.WebhookHandler)
	}

	// Поддерживаем как /api, так и /api/v1
	r.Route("/api", func(r chi.Router) {
		registerAPIRoutes(r, p)
	})
	r.Route("/api/v1", func(r chi.Router) {
		registerAPIRoutes(r, p)
	})

	return r
}
