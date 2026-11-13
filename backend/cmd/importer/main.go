package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/slicer"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HFPhraseItem запись датасета bashkorttele/trilingual-parallel-phrasebooks-bgpu
type HFPhraseItem struct {
	ID        string `json:"id"`
	TextBa    string `json:"text_ba"`
	TextRu    string `json:"text_ru"`
	TextOther string `json:"text_other,omitempty"`
	LangOther string `json:"lang_other,omitempty"`
}

// HFAudiobookItem запись датасета bashkorttele/narrated-audiobooks-brsbs
type HFAudiobookItem struct {
	AudioURL      string `json:"audio_url"`
	BookTitle     string `json:"book_title"`
	Author        string `json:"author,omitempty"`
	Chapter       string `json:"chapter,omitempty"`
	StartMS       int    `json:"start_ms"`
	EndMS         int    `json:"end_ms"`
	DurationMS    int    `json:"duration_ms,omitempty"`
	TextBa        string `json:"text_ba"`
	TranslationRu string `json:"translation_ru,omitempty"`
}

// HFTextItem запись художественного произведения, новостей или учебника
type HFTextItem struct {
	SourceTitle string `json:"source_title"`
	SourceType  string `json:"source_type"` // text_literature, text_news
	Author      string `json:"author,omitempty"`
	Chapter     string `json:"chapter,omitempty"`
	Text        string `json:"text"`
}

// ImportStats статистика работы импортера
type ImportStats struct {
	SourcesCount      int
	FragmentsCount    int
	AudioFragments    int
	TextFragments     int
	TranslationsCount int
	HoneypotsCount    int
}

func main() {
	var (
		datasetType  = flag.String("type", "all", "Тип импортируемого датасета: all, phrasebook, audiobook, text")
		filePath     = flag.String("file", "", "Путь к конкретному файлу датасета (jsonl/json)")
		dbURL        = flag.String("db", "", "Строка подключения к PostgreSQL (DATABASE_URL)")
		limit        = flag.Int("limit", 0, "Максимальное число фрагментов (0 = без ограничений)")
		honeypotRate = flag.Float64("honeypot-rate", 0.20, "Доля контрольных ловушек античита Honeypot (по умолчанию 0.20 = 20%)")
		dryRun       = flag.Bool("dry-run", false, "Режим симуляции (нарезка и проверка без записи в БД)")
	)
	flag.Parse()

	printBanner()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			*dbURL = "postgres://postgres:postgres@localhost:5432/bashkir_school?sslmode=disable"
		}
	}

	ctx := context.Background()
	var pool *pgxpool.Pool
	var err error

	if !*dryRun {
		pool, err = pgxpool.New(ctx, *dbURL)
		if err != nil {
			log.Printf("⚠️  Не удалось инициализировать пул БД (%v). Переключение в режим --dry-run.", err)
			*dryRun = true
		} else {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := pool.Ping(pingCtx); err != nil {
				log.Printf("⚠️  PostgreSQL недоступен (%v). Переключение в режим --dry-run (нарезка и валидация).", err)
				*dryRun = true
				pool.Close()
				pool = nil
			} else {
				log.Println("✅ Успешное подключение к PostgreSQL!")
				defer pool.Close()
			}
		}
	}

	stats := &ImportStats{}

	// Определяем пути к встроенным датасетам
	dataDir := findDataDirectory()
	log.Printf("📁 Директория с данными: %s", dataDir)

	// 1. Импорт параллельных разговорников Hugging Face (trilingual phrasebooks)
	if *datasetType == "all" || *datasetType == "phrasebook" {
		targetFile := *filePath
		if targetFile == "" || *datasetType == "all" {
			targetFile = filepath.Join(dataDir, "data-ba-ru-kk.jsonl")
		}
		if fileExists(targetFile) {
			log.Printf("📖 [1/3] Обработка датасета разговорников: %s", targetFile)
			importPhrasebookDataset(ctx, pool, targetFile, *limit, *honeypotRate, *dryRun, stats)
		} else {
			log.Printf("ℹ️ Файл разговорников не найден по пути: %s (пропуск)", targetFile)
		}
	}

	// 2. Импорт аудиокниг Hugging Face (bashkorttele/narrated-audiobooks-brsbs)
	if *datasetType == "all" || *datasetType == "audiobook" {
		targetAudioFile := *filePath
		if targetAudioFile == "" || *datasetType == "all" {
			targetAudioFile = filepath.Join(dataDir, "audiobooks_brsbs.jsonl")
		}
		if fileExists(targetAudioFile) {
			log.Printf("🎧 [2/3] Обработка датасета аудиокниг (BRSBS): %s", targetAudioFile)
			importAudiobooksDataset(ctx, pool, targetAudioFile, *limit, *honeypotRate, *dryRun, stats)
		} else {
			log.Printf("ℹ️ Файл аудиокниг не найден по пути: %s (пропуск)", targetAudioFile)
		}
	}

	// 3. Нарезка художественной литературы, новостей и учебников
	if *datasetType == "all" || *datasetType == "text" {
		targetTextFile := *filePath
		if targetTextFile == "" || *datasetType == "all" {
			targetTextFile = filepath.Join(dataDir, "texts_samples.jsonl")
		}
		if fileExists(targetTextFile) {
			log.Printf("📚 [3/3] Нарезка художественных произведений и новостей: %s", targetTextFile)
			importTextDataset(ctx, pool, targetTextFile, *limit, *honeypotRate, *dryRun, stats)
		} else {
			log.Printf("ℹ️ Файл текстовых произведений не найден: %s (пропуск)", targetTextFile)
		}
	}

	printSummary(stats, *dryRun)
}

func printBanner() {
	fmt.Println("================================================================================")
	fmt.Println("🎓 «ЙӘНЛӘНДЕР» — СИСТЕМА ИМПОРТА И НАРЕЗКИ ПАРАЛЛЕЛЬНОГО КОРПУСА (HUGGING FACE)")
	fmt.Println("   • Текстовые материалы: художественные произведения, новости, учебники")
	fmt.Println("   • Аудиоматериалы: аудиокниги БРСБС, радиопередачи с таймкодами (audio_start_ms)")
	fmt.Println("   • Античит Honeypot: автоматическое внедрение ловушек в ~20% автопереводов")
	fmt.Println("================================================================================")
}

func printSummary(stats *ImportStats, dryRun bool) {
	fmt.Println("--------------------------------------------------------------------------------")
	if dryRun {
		fmt.Println("📊 ИТОГИ НАРЕЗКИ И ВАЛИДАЦИИ (РЕЖИМ DRY-RUN — БЕЗ ЗАПИСИ В БД):")
	} else {
		fmt.Println("📊 ИТОГИ ИМПОРТА В БАЗУ ДАННЫХ POSTGRESQL:")
	}
	fmt.Printf("   • Создано источников (sources):             %d\n", stats.SourcesCount)
	fmt.Printf("   • Всего нарезано фрагментов (fragments):    %d\n", stats.FragmentsCount)
	fmt.Printf("     - Текстовых фрагментов:                   %d\n", stats.TextFragments)
	fmt.Printf("     - Аудиофрагментов с таймкодами:           %d\n", stats.AudioFragments)
	fmt.Printf("   • Зарегистрировано переводов (translations): %d\n", stats.TranslationsCount)
	fmt.Printf("   • Внедрено античит-ловушек (honeypot ~20%%): %d\n", stats.HoneypotsCount)
	fmt.Println("================================================================================")
	fmt.Println("✅ Процесс успешно завершен!")
}

// importPhrasebookDataset импортирует разговорники БГПУ (ba-ru-kk)
func importPhrasebookDataset(ctx context.Context, pool *pgxpool.Pool, path string, limit int, honeypotRate float64, dryRun bool, stats *ImportStats) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Ошибка чтения файла разговорников: %v", err)
		return
	}
	defer f.Close()

	sourceID := uuid.New()
	sourceTitle := "БГПУ им. М. Акмуллы: трехъязычные параллельные разговорники"
	hfName := "bashkorttele/trilingual-parallel-phrasebooks-bgpu"
	metaJSON, _ := json.Marshal(map[string]any{
		"institution": "БГПУ",
		"license":     "cc-by-4.0",
		"type":        "parallel_phrasebook",
	})

	if !dryRun && pool != nil {
		_, err := pool.Exec(ctx, `
			INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`, sourceID, sourceTitle, string(domain.SourceHFDataset),
			"https://huggingface.co/datasets/bashkorttele/trilingual-parallel-phrasebooks-bgpu",
			hfName, string(metaJSON))
		if err != nil {
			log.Printf("Предупреждение при вставке источника: %v", err)
		}
	}
	stats.SourcesCount++

	scanner := bufio.NewScanner(f)
	idx := 0
	var prevRu string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item HFPhraseItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		idx++
		if limit > 0 && idx > limit {
			break
		}

		fragID := uuid.New()
		dialogNum := (idx / 5) + 1
		chapter := fmt.Sprintf("Разговорник БГПУ, диалог %d", dialogNum)
		paragraphIdx := dialogNum
		sentenceIdx := idx

		// Античит ловушка Honeypot
		hasHoneypot := slicer.ShouldInjectHoneypot(idx, honeypotRate)
		var honeypotVariant *string

		if hasHoneypot {
			distractor := prevRu
			_, variant := slicer.GenerateHoneypot(item.TextRu, distractor)
			honeypotVariant = &variant
			stats.HoneypotsCount++
		}
		prevRu = item.TextRu

		if !dryRun && pool != nil {
			// Вставка фрагмента
			_, err = pool.Exec(ctx, `
				INSERT INTO fragments (id, source_id, text_ba, chapter_info, paragraph_idx, sentence_idx, is_active)
				VALUES ($1, $2, $3, $4, $5, $6, TRUE)
			`, fragID, sourceID, item.TextBa, chapter, paragraphIdx, sentenceIdx)
			if err != nil {
				continue
			}

			// 1. ba-ru перевод
			ftID := uuid.New()
			_, _ = pool.Exec(ctx, `
				INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
				VALUES ($1, $2, 'ba-ru', $3, $4, $5, 'pending', 0)
			`, ftID, fragID, item.TextRu, hasHoneypot, honeypotVariant)
			stats.TranslationsCount++

			// 2. ba-kk (казахский перевод, если есть)
			if item.TextOther != "" {
				ftKkID := uuid.New()
				_, _ = pool.Exec(ctx, `
					INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
					VALUES ($1, $2, 'ba-kk', $3, FALSE, NULL, 'pending', 0)
				`, ftKkID, fragID, item.TextOther)
				stats.TranslationsCount++
			}
		} else {
			stats.TranslationsCount++
			if item.TextOther != "" {
				stats.TranslationsCount++
			}
		}

		stats.FragmentsCount++
		stats.TextFragments++
	}

	log.Printf("   ✓ Обработано %d фраз разговорников", idx)
}

// importAudiobooksDataset нарезает и импортирует аудиокниги bashkorttele/narrated-audiobooks-brsbs
func importAudiobooksDataset(ctx context.Context, pool *pgxpool.Pool, path string, limit int, honeypotRate float64, dryRun bool, stats *ImportStats) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Ошибка чтения файла аудиокниг: %v", err)
		return
	}
	defer f.Close()

	segmenter := slicer.NewAudioSegmenter(10000)

	scanner := bufio.NewScanner(f)
	idx := 0
	var prevRu string
	sourcesCache := make(map[string]uuid.UUID)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item HFAudiobookItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		idx++
		if limit > 0 && idx > limit {
			break
		}

		// Идентификатор источника
		sourceTitle := item.BookTitle
		if sourceTitle == "" {
			sourceTitle = "Башкирская аудиокнига БРСБС"
		}

		sourceID, ok := sourcesCache[sourceTitle]
		if !ok {
			sourceID = uuid.New()
			sourcesCache[sourceTitle] = sourceID

			metaJSON, _ := json.Marshal(map[string]any{
				"author":      item.Author,
				"collection":  "BRSBS",
				"audio_track": item.AudioURL,
			})

			if !dryRun && pool != nil {
				_, _ = pool.Exec(ctx, `
					INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
					VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT DO NOTHING
				`, sourceID, sourceTitle, string(domain.SourceAudioBook),
					item.AudioURL, "bashkorttele/narrated-audiobooks-brsbs", string(metaJSON))
			}
			stats.SourcesCount++
		}

		// Сборка аудиометаданных для нарезчика
		meta := slicer.AudioMetadata{
			SourceTitle: sourceTitle,
			Chapter:     item.Chapter,
			AudioURL:    item.AudioURL,
			DurationMS:  item.DurationMS,
			Cues: []slicer.AudioCue{
				{
					StartMS:     item.StartMS,
					EndMS:       item.EndMS,
					TextBa:      item.TextBa,
					TextRu:      item.TranslationRu,
					ChapterInfo: item.Chapter,
				},
			},
		}

		fragments := segmenter.SliceAudio(meta)
		for _, frag := range fragments {
			fragID := uuid.New()

			// Античит Honeypot
			hasHoneypot := slicer.ShouldInjectHoneypot(idx, honeypotRate)
			var honeypotVariant *string

			ruText := item.TranslationRu
			if ruText == "" {
				ruText = "[Черновой автоперевод аудиофрагмента]"
			}

			if hasHoneypot {
				_, variant := slicer.GenerateHoneypot(ruText, prevRu)
				honeypotVariant = &variant
				stats.HoneypotsCount++
			}
			prevRu = ruText

			if !dryRun && pool != nil {
				_, err = pool.Exec(ctx, `
					INSERT INTO fragments (id, source_id, text_ba, audio_url, audio_start_ms, audio_end_ms, chapter_info, paragraph_idx, sentence_idx, is_active)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE)
				`, fragID, sourceID, frag.TextBa, frag.AudioURL, frag.AudioStartMS, frag.AudioEndMS, frag.ChapterInfo, frag.ParagraphIdx, frag.SentenceIdx)
				if err != nil {
					continue
				}

				ftID := uuid.New()
				_, _ = pool.Exec(ctx, `
					INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
					VALUES ($1, $2, 'ba-ru', $3, $4, $5, 'pending', 0)
				`, ftID, fragID, ruText, hasHoneypot, honeypotVariant)
			}

			stats.FragmentsCount++
			stats.AudioFragments++
			stats.TranslationsCount++
		}
	}

	log.Printf("   ✓ Нарезано %d аудиофрагментов с таймкодами", stats.AudioFragments)
}

// importTextDataset нарезает художественную литературу, новости и учебники с помощью TextSlicer
func importTextDataset(ctx context.Context, pool *pgxpool.Pool, path string, limit int, honeypotRate float64, dryRun bool, stats *ImportStats) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Ошибка чтения файла текстов: %v", err)
		return
	}
	defer f.Close()

	textSlicer := slicer.NewTextSlicer()
	scanner := bufio.NewScanner(f)
	lineIdx := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item HFTextItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		lineIdx++
		sourceID := uuid.New()
		srcType := domain.SourceType(item.SourceType)
		if srcType == "" {
			srcType = domain.SourceTextLiterature
		}

		metaJSON, _ := json.Marshal(map[string]any{
			"author": item.Author,
			"source": "Литературный и новостной корпус Башкортостана",
		})

		if !dryRun && pool != nil {
			_, _ = pool.Exec(ctx, `
				INSERT INTO sources (id, title, source_type, external_url, hf_dataset_name, metadata)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT DO NOTHING
			`, sourceID, item.SourceTitle, string(srcType),
				"https://kitaptar.bashkort.org", "bashkorttele/literature-corpus", string(metaJSON))
		}
		stats.SourcesCount++

		// Нарезка на параграфы и предложения с координатами источника
		sliced := textSlicer.SliceText(item.Text, item.Chapter)
		var prevTranslation string

		for _, frag := range sliced {
			if limit > 0 && stats.FragmentsCount >= limit {
				break
			}

			fragID := uuid.New()
			// Генерируем ориентировочный черновой автоперевод
			autoRu := fmt.Sprintf("[Машинный перевод] %s", frag.Text)

			hasHoneypot := slicer.ShouldInjectHoneypot(frag.SentenceIdx, honeypotRate)
			var honeypotVariant *string

			if hasHoneypot {
				_, variant := slicer.GenerateHoneypot(autoRu, prevTranslation)
				honeypotVariant = &variant
				stats.HoneypotsCount++
			}
			prevTranslation = autoRu

			if !dryRun && pool != nil {
				_, err = pool.Exec(ctx, `
					INSERT INTO fragments (id, source_id, text_ba, chapter_info, paragraph_idx, sentence_idx, is_active)
					VALUES ($1, $2, $3, $4, $5, $6, TRUE)
				`, fragID, sourceID, frag.Text, frag.ChapterInfo, frag.ParagraphIdx, frag.SentenceIdx)
				if err != nil {
					continue
				}

				ftID := uuid.New()
				_, _ = pool.Exec(ctx, `
					INSERT INTO fragment_translations (id, fragment_id, lang_pair, initial_auto_translation, has_honeypot, honeypot_variant, status, confirmed_count)
					VALUES ($1, $2, 'ba-ru', $3, $4, $5, 'pending', 0)
				`, ftID, fragID, autoRu, hasHoneypot, honeypotVariant)
			}

			stats.FragmentsCount++
			stats.TextFragments++
			stats.TranslationsCount++
		}
	}

	log.Printf("   ✓ Нарезано текстов художественных произведений и новостей: %d предложений", stats.TextFragments)
}

func findDataDirectory() string {
	candidates := []string{
		"data",
		"backend/data",
		"../data",
		"C:/Users/user/Desktop/latest/hacaton-kiber-102-void/backend/data",
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return "data"
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir() || (err == nil && info.IsDir())
}
