package main

import (
	"context"
	"log"
	"time"

	"bashkir-backend/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("=== Запуск сидера демонстрационных данных ===")
	cfg := config.LoadConfig()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	// 1. Создание школ
	school1ID := uuid.New()
	school2ID := uuid.New()

	_, _ = pool.Exec(ctx, `
		INSERT INTO schools (id, name, city) VALUES 
		($1, 'Башкирская республиканская гимназия-интернат № 1 им. Р. Гарипова', 'г. Уфа'),
		($2, 'Гимназия № 3 им. М. Горького', 'г. Уфа')
		ON CONFLICT DO NOTHING
	`, school1ID, school2ID)

	// 2. Создание классов
	class1ID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO classes (id, school_id, grade_number, letter) VALUES 
		($1, $2, 9, 'А')
		ON CONFLICT DO NOTHING
	`, class1ID, school1ID)

	// 3. Создание источника литературы
	sourceID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
		VALUES ($1, 'Мостай Кәрим «Оҙон-оҙаҡ бала саҡ» (Повесть)', 'text_literature', 'https://huggingface.co/bashkorttele', 'bashkorttele/datasets', '{"author": "Мустай Карим", "year": 1976}'::jsonb)
		ON CONFLICT DO NOTHING
	`, sourceID)

	// 4. Добавление тестовых фрагментов башкирского текста
	type SampleFragment struct {
		TextBa          string
		AutoTranslation string
		HasHoneypot     bool
		HoneypotVariant *string
		Chapter         string
		SentenceIdx     int
	}

	honeypotVar1 := "Башкортостан — наша чужая страна." // Намеренная ошибка: "чужая" вместо "родная"
	samples := []SampleFragment{
		{
			TextBa:          "Башҡортостан — беҙҙең тыуған төйәгебеҙ, ата-бабаларыбыҙ төйәге.",
			AutoTranslation: "Башкортостан — наш родной край, край наших предков.",
			HasHoneypot:     true,
			HoneypotVariant: &honeypotVar1,
			Chapter:         "Глава 1, «Кендек»",
			SentenceIdx:     1,
		},
		{
			TextBa:          "Ҡояш тау артына байыны, ауыл өҫтөнә тынлыҡ төштө.",
			AutoTranslation: "Солнце зашло за гору, на деревню опустилась тишина.",
			HasHoneypot:     false,
			HoneypotVariant: nil,
			Chapter:         "Глава 1, «Кендек»",
			SentenceIdx:     2,
		},
		{
			TextBa:          "Китап — белем шишмәһе, аҡылға юлдаш.",
			AutoTranslation: "Книга — источник знаний, спутник разума.",
			HasHoneypot:     false,
			HoneypotVariant: nil,
			Chapter:         "Учебные тексты",
			SentenceIdx:     1,
		},
		{
			TextBa:          "Дуҫлыҡ — тормоштоң иң ҙур байлығы.",
			AutoTranslation: "Дружба — самое большое богатство жизни.",
			HasHoneypot:     false,
			HoneypotVariant: nil,
			Chapter:         "Пословицы и поговорки",
			SentenceIdx:     4,
		},
		{
			TextBa:          "Ағиҙел буйлап аҡ пароход йөҙә.",
			AutoTranslation: "По реке Белой плывет белый теплоход.",
			HasHoneypot:     false,
			HoneypotVariant: nil,
			Chapter:         "Рассказы о родном крае",
			SentenceIdx:     7,
		},
		{
			TextBa:          "Бала саҡ — кеше ғүмеренең иң яҡты, ҡайғыһыҙ осоро.",
			AutoTranslation: "Детство — самый светлый, беззаботный период человеческой жизни.",
			HasHoneypot:     false,
			HoneypotVariant: nil,
			Chapter:         "Глава 2",
			SentenceIdx:     12,
		},
	}

	for _, s := range samples {
		fragID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO fragments (id, source_id, text_ba, chapter_info, sentence_idx, is_active)
			VALUES ($1, $2, $3, $4, $5, TRUE)
		`, fragID, sourceID, s.TextBa, s.Chapter, s.SentenceIdx)
		if err != nil {
			continue
		}

		ftID := uuid.New()
		_, _ = pool.Exec(ctx, `
			INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
			VALUES ($1, $2, 'ba-ru', $3, $4, $5, 'pending', 0)
		`, ftID, fragID, s.AutoTranslation, s.HasHoneypot, s.HoneypotVariant)
	}

	// 5. Добавление тестового пользователя в лидерборд
	testUserID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, tg_username, full_name, role, school_id, class_id, current_streak, max_streak, total_xp, last_active_date)
		VALUES ($1, 'salawat_hero', 'Салават Юлаев', 'student', $2, $3, 5, 10, 480, $4)
		ON CONFLICT DO NOTHING
	`, testUserID, school1ID, class1ID, time.Now())

	log.Println("✅ Демонстрационные данные успешно загружены в базу!")
}
