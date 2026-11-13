package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bashkir-backend/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HFPhraseItem struct {
	ID        string `json:"id"`
	TextBa    string `json:"text_ba"`
	TextRu    string `json:"text_ru"`
	TextOther string `json:"text_other"`
	LangOther string `json:"lang_other"`
}

func main() {
	log.Println("=== Запуск сидера базы данных (Hugging Face + Организации) ===")
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

	// 2. Создание классов с инвайт-кодами
	class1ID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO classes (id, school_id, grade_number, letter, invite_code) VALUES 
		($1, $2, 9, 'А', 'INV-9A-GIMN1')
		ON CONFLICT DO NOTHING
	`, class1ID, school1ID)

	// 3. Создание источников материалов
	sourceLitID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
		VALUES ($1, 'Мостай Кәрим «Оҙон-оҙаҡ бала саҡ» (Повесть)', 'text_literature', 'https://huggingface.co/bashkorttele', 'bashkorttele/datasets', '{"author": "Мустай Карим", "year": 1976}'::jsonb)
		ON CONFLICT DO NOTHING
	`, sourceLitID)

	sourceHFID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
		VALUES ($1, 'БГПУ им. М. Акмуллы: трехъязычные параллельные разговорники', 'hf_dataset', 'https://huggingface.co/datasets/bashkorttele/trilingual-parallel-phrasebooks-bgpu', 'bashkorttele/trilingual-parallel-phrasebooks-bgpu', '{"institution": "БГПУ", "license": "cc-by-4.0"}'::jsonb)
		ON CONFLICT DO NOTHING
	`, sourceHFID)

	// 4. Поиск файла датасета с Hugging Face
	possiblePaths := []string{
		"data/data-ba-ru-kk.jsonl",
		"backend/data/data-ba-ru-kk.jsonl",
		"../data/data-ba-ru-kk.jsonl",
		"C:/Users/user/Desktop/latest/hacaton-kiber-102-void/backend/data/data-ba-ru-kk.jsonl",
	}

	var jsonlFile *os.File
	for _, p := range possiblePaths {
		f, err := os.Open(p)
		if err == nil {
			jsonlFile = f
			log.Printf("Загрузка датасета из файла: %s", p)
			break
		}
	}

	hfLoadedCount := 0
	if jsonlFile != nil {
		defer jsonlFile.Close()
		scanner := bufio.NewScanner(jsonlFile)
		idx := 0
		var prevRu string

		for scanner.Scan() && idx < 60 {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var item HFPhraseItem
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				continue
			}

			idx++
			fragID := uuid.New()
			chapter := fmt.Sprintf("Разговорник БГПУ, диалог %d", (idx/5)+1)

			_, err := pool.Exec(ctx, `
				INSERT INTO fragments (id, source_id, text_ba, chapter_info, sentence_idx, is_active)
				VALUES ($1, $2, $3, $4, $5, TRUE)
			`, fragID, sourceHFID, item.TextBa, chapter, idx)
			if err != nil {
				continue
			}

			// Античит-ловушка Honeypot: в 20% записей искажаем автоперевод
			hasHoneypot := false
			var honeypotVariant *string

			if idx%5 == 0 && prevRu != "" {
				hasHoneypot = true
				wrong := fmt.Sprintf("[НЕВЕРНО] %s", prevRu)
				honeypotVariant = &wrong
			}
			prevRu = item.TextRu

			// 1. Пара башкирский <-> русский (ba-ru)
			ftID := uuid.New()
			_, _ = pool.Exec(ctx, `
				INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
				VALUES ($1, $2, 'ba-ru', $3, $4, $5, 'pending', 0)
			`, ftID, fragID, item.TextRu, hasHoneypot, honeypotVariant)

			// 2. Пара башкирский <-> казахский (ba-kk)
			if item.TextOther != "" {
				ftOtherID := uuid.New()
				_, _ = pool.Exec(ctx, `
					INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
					VALUES ($1, $2, 'ba-kk', $3, FALSE, NULL, 'pending', 0)
				`, ftOtherID, fragID, item.TextOther)
			}

			hfLoadedCount++
		}
		log.Printf("Успешно импортировано %d фрагментов из Hugging Face!", hfLoadedCount)
	}

	// 5. Образцы литературы
	samples := []struct {
		TextBa          string
		AutoTranslation string
		Chapter         string
		SentenceIdx     int
	}{
		{
			TextBa:          "Башҡортостан — беҙҙең тыуған төйәгебеҙ, ата-бабаларыбыҙ төйәге.",
			AutoTranslation: "Башкортостан — наш родной край, край наших предков.",
			Chapter:         "Глава 1, «Кендек»",
			SentenceIdx:     1,
		},
		{
			TextBa:          "Ҡояш тау артына байыны, ауыл өҫтөнә тынлыҡ төштө.",
			AutoTranslation: "Солнце зашло за гору, на деревню опустилась тишина.",
			Chapter:         "Глава 1, «Кендек»",
			SentenceIdx:     2,
		},
		{
			TextBa:          "Китап — белем шишмәһе, аҡылға юлдаш.",
			AutoTranslation: "Книга — источник знаний, спутник разума.",
			Chapter:         "Учебные тексты",
			SentenceIdx:     1,
		},
		{
			TextBa:          "Дуҫлыҡ — тормоштоң иң ҙур байлығы.",
			AutoTranslation: "Дружба — самое большое богатство жизни.",
			Chapter:         "Пословицы и поговорки",
			SentenceIdx:     4,
		},
	}

	for i, s := range samples {
		fragID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO fragments (id, source_id, text_ba, chapter_info, sentence_idx, is_active)
			VALUES ($1, $2, $3, $4, $5, TRUE)
		`, fragID, sourceLitID, s.TextBa, s.Chapter, s.SentenceIdx)
		if err != nil {
			continue
		}

		status := "pending"
		if i == 0 {
			// Направляем первый фрагмент на экспертизу УФИЦ РАН для демонстрации очередей проверки
			status = "expert_review"
		}

		ftID := uuid.New()
		_, _ = pool.Exec(ctx, `
			INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
			VALUES ($1, $2, 'ba-ru', $3, FALSE, NULL, $4, 0)
		`, ftID, fragID, s.AutoTranslation, status)
	}

	// 6. Добавление тестового пользователя в лидерборд
	testUserID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, tg_username, full_name, role, preferred_lang_pair, school_id, class_id, current_streak, max_streak, total_xp, last_active_date)
		VALUES ($1, 'salawat_hero', 'Салават Юлаев', 'student', 'ba-ru', $2, $3, 5, 10, 480, $4)
		ON CONFLICT DO NOTHING
	`, testUserID, school1ID, class1ID, time.Now())

	// 7. Добавление учителя и эксперта УФИЦ РАН для тестирования ЛМС и панели экспертов
	teacherID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, email, full_name, role, preferred_lang_pair, school_id, class_id, current_streak, total_xp)
		VALUES ($1, 'teacher@school1.ru', 'Зухра Рашитовна (Учитель)', 'teacher', 'ba-ru', $2, $3, 10, 850)
		ON CONFLICT DO NOTHING
	`, teacherID, school1ID, class1ID)

	expertID := uuid.New()
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, email, full_name, role, preferred_lang_pair, total_xp)
		VALUES ($1, 'expert@ufic-ran.ru', 'Азат Салаватович (Эксперт УФИЦ РАН)', 'expert', 'ba-ru', 1200)
		ON CONFLICT DO NOTHING
	`, expertID)

	log.Println("✅ Демонстрационные данные успешно загружены в базу!")
}
