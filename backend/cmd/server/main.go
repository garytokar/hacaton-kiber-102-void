package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bashkir-backend/internal/config"
	deliveryBot "bashkir-backend/internal/delivery/bot"
	deliveryHTTP "bashkir-backend/internal/delivery/http"
	v1 "bashkir-backend/internal/delivery/http/v1"
	"bashkir-backend/internal/repository"
	"bashkir-backend/internal/repository/postgres"
	"bashkir-backend/internal/repository/redis"
	"bashkir-backend/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	log.Println("=== Запуск платформы «Башкирская языковая школа» (Unified Backend) ===")

	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Подключение к PostgreSQL
	log.Println("[PostgreSQL] Подключение к пулу БД...")
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[PostgreSQL] Ошибка подключения: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("[PostgreSQL] База данных недоступна: %v", err)
	}
	log.Println("[PostgreSQL] Успешно подключено!")

	// 2. Автоматическое применение миграций схемы
	migrationPath := "migrations/000001_init_schema.up.sql"
	if err := repository.RunMigrations(ctx, dbPool, migrationPath); err != nil {
		log.Printf("[Migration WARNING] Не удалось применить миграцию (%v), возможно таблицы уже существуют.", err)
	}

	// 3. Подключение к Redis
	log.Println("[Redis] Подключение...")
	redisOpts, err := goredis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[Redis] Некорректный REDIS_URL: %v", err)
	}
	redisClient := goredis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis WARNING] Redis недоступен: %v. Продолжаем без кеша.", err)
	} else {
		log.Println("[Redis] Успешно подключено!")
	}

	// 4. Инициализация репозиториев
	userRepo := postgres.NewUserRepository(dbPool)
	contentRepo := postgres.NewContentRepository(dbPool)
	submissionRepo := postgres.NewSubmissionRepository(dbPool)
	lmsRepo := postgres.NewLMSRepository(dbPool)
	redisRepo := redis.NewLeaderboardRedisRepo(redisClient)

	// 5. Инициализация единого слоя Use Cases
	authUC := usecase.NewAuthUseCase(userRepo, cfg)
	gamifyUC := usecase.NewGamificationUseCase(userRepo, submissionRepo, redisRepo, cfg)
	consensusUC := usecase.NewConsensusUseCase(contentRepo, submissionRepo, cfg)
	taskUC := usecase.NewTaskUseCase(contentRepo, submissionRepo, userRepo, gamifyUC, consensusUC)
	lmsUC := usecase.NewLMSUseCase(lmsRepo)
	exportUC := usecase.NewExportUseCase(contentRepo)

	// 6. Инициализация Telegram бота (Dual-Ingress)
	botSvc, err := deliveryBot.NewBotService(cfg, authUC, taskUC, gamifyUC, lmsUC, redisRepo)
	if err != nil {
		log.Printf("[Telegram Bot WARNING] Ошибка инициализации бота: %v", err)
	}

	var webhookHandler http.HandlerFunc
	if botSvc != nil && cfg.BotMode == "webhook" {
		webhookHandler = botSvc.WebhookHandler()
	}

	// 7. Инициализация HTTP роутера (REST API)
	authH := v1.NewAuthHandler(authUC, userRepo)
	taskH := v1.NewTaskHandler(taskUC)
	lbH := v1.NewLeaderboardHandler(gamifyUC, lmsUC)
	lmsH := v1.NewLMSHandler(lmsUC)
	exportH := v1.NewExportHandler(exportUC)

	httpRouter := deliveryHTTP.NewRouter(deliveryHTTP.RouterParams{
		Cfg:            cfg,
		AuthHandler:    authH,
		TaskHandler:    taskH,
		LeaderboardH:   lbH,
		LMSHandler:     lmsH,
		ExportHandler:  exportH,
		WebhookHandler: webhookHandler,
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 8. Запуск сервера и бота на горутинах
	go func() {
		log.Printf("[HTTP Server] Запущен на порту :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[HTTP Server] Фатальная ошибка: %v", err)
		}
	}()

	if botSvc != nil && cfg.BotMode == "polling" {
		go botSvc.Start()
	}

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n[Shutdown] Получен сигнал завершения. Остановка сервисов...")
	if botSvc != nil {
		botSvc.Stop()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Shutdown] Ошибка остановки HTTP сервера: %v", err)
	}

	log.Println("[Shutdown] Сервис успешно остановлен.")
}
