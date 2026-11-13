package repository

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations выполняет SQL-миграции при старте приложения
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationFilePath string) error {
	log.Printf("[Migration] Чтение файла миграции: %s", migrationFilePath)
	content, err := os.ReadFile(migrationFilePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла миграций: %w", err)
	}

	log.Println("[Migration] Применение схемы базы данных...")
	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}

	log.Println("[Migration] Схема базы данных успешно применена!")
	return nil
}
