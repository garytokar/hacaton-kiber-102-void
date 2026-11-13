package slicer

import (
	"math/rand"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TextFragment представляет атомарный фрагмент текста с координатами первоисточника.
type TextFragment struct {
	ParagraphIdx int    `json:"paragraph_idx"`
	SentenceIdx  int    `json:"sentence_idx"`
	Text         string `json:"text"`
	ChapterInfo  string `json:"chapter_info,omitempty"`
}

// AudioCue описывает размеченный отрезок аудиозаписи с транскрипцией.
type AudioCue struct {
	StartMS     int    `json:"audio_start_ms"`
	EndMS       int    `json:"audio_end_ms"`
	TextBa      string `json:"text_ba,omitempty"`
	TextRu      string `json:"text_ru,omitempty"`
	ChapterInfo string `json:"chapter_info,omitempty"`
	Speaker     string `json:"speaker,omitempty"`
}

// AudioMetadata метаданные аудиоисточника (аудиокнига, радиопередача).
type AudioMetadata struct {
	SourceTitle string     `json:"source_title"`
	Chapter     string     `json:"chapter"`
	AudioURL    string     `json:"audio_url"`
	DurationMS  int        `json:"duration_ms"`
	Cues        []AudioCue `json:"cues,omitempty"`
}

// AudioFragment фрагмент аудиозаписи с таймкодами и метаданными.
type AudioFragment struct {
	SourceTitle  string `json:"source_title"`
	ChapterInfo  string `json:"chapter_info"`
	AudioURL     string `json:"audio_url"`
	AudioStartMS int    `json:"audio_start_ms"`
	AudioEndMS   int    `json:"audio_end_ms"`
	TextBa       string `json:"text_ba"`
	SentenceIdx  int    `json:"sentence_idx"`
	ParagraphIdx int    `json:"paragraph_idx"`
}

// HoneypotResult результат проверки или генерации античит-ловушки.
type HoneypotResult struct {
	HasHoneypot     bool   `json:"has_honeypot"`
	HoneypotVariant string `json:"honeypot_variant,omitempty"`
}

// TextSlicerConfig параметры токенизатора предложений.
type TextSlicerConfig struct {
	MinRunes        int
	MaxRunes        int
	AbbrevExceptions map[string]bool
}

// DefaultTextSlicerConfig конфигурация по умолчанию.
func DefaultTextSlicerConfig() TextSlicerConfig {
	return TextSlicerConfig{
		MinRunes: 3,
		MaxRunes: 600,
		AbbrevExceptions: map[string]bool{
			"һ.б.":   true, // һәм башҡалар (и другие)
			"һ.б.ҡ.": true, // һәм башҡа ҡайһы берҙәр
			"т.б.":   true, // тағы башҡалар
			"б-са":   true, // буйынса (согласно)
			"б.":     true, // бит / быуат
			"й.":     true, // йыл
			"йй.":    true, // йылдар
			"г.":     true, // город
			"ул.":    true, // улица
			"им.":    true, // имени
			"проф.":  true, // профессор
			"акад.":  true, // академик
			"доц.":   true, // доцент
			"д.":     true, // деревня / дом
			"руб.":   true, // рубль
			"коп.":   true, // копейка
			"стр.":   true, // страница
			"рис.":   true, // рисунок
			"экз.":   true, // экземпляр
			"мин.":   true, // минута
			"сек.":   true, // секунда
			"см.":    true, // сантиметр
			"мм.":    true, // миллиметр
			"км.":    true, // километр
			"кг.":    true, // килограмм
			"т.е.":   true, // то есть
			"т.к.":   true, // так как
			"в.":     true, // век
			"вв.":    true, // века
			"н.э.":   true, // нашей эры
			"б.э.т.": true, // до н.э.
		},
	}
}

// TextSlicer разбивает текст художественных произведений, новостей и учебников на предложения.
type TextSlicer struct {
	cfg TextSlicerConfig
}

// NewTextSlicer создает экземпляр нарезчика текста.
func NewTextSlicer(cfg ...TextSlicerConfig) *TextSlicer {
	c := DefaultTextSlicerConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return &TextSlicer{cfg: c}
}

// NormalizeSpaces сжимает пробельные символы в один пробел.
func NormalizeSpaces(s string) string {
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}

// SplitParagraphs разделяет текст на абзацы по пустым строкам.
func SplitParagraphs(rawText string) []string {
	normalized := strings.ReplaceAll(rawText, "\r\n", "\n")
	rawParagraphs := strings.Split(normalized, "\n")

	var paragraphs []string
	var current []string

	for _, line := range rawParagraphs {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(current) > 0 {
				p := NormalizeSpaces(strings.Join(current, " "))
				if p != "" {
					paragraphs = append(paragraphs, p)
				}
				current = nil
			}
		} else {
			current = append(current, trimmed)
		}
	}

	if len(current) > 0 {
		p := NormalizeSpaces(strings.Join(current, " "))
		if p != "" {
			paragraphs = append(paragraphs, p)
		}
	}

	return paragraphs
}

// isInitial проверяет, является ли токен инициалом (например, "Р.", "М.", "Ғ.").
func isInitial(token string) bool {
	runes := []rune(token)
	if len(runes) == 2 && runes[1] == '.' && unicode.IsUpper(runes[0]) {
		return true
	}
	return false
}

// isSentenceTerminator определяет знаки окончания предложения.
func isSentenceTerminator(r rune) bool {
	return r == '.' || r == '!' || r == '?' || r == '…'
}

// isQuoteClose проверяет закрывающие кавычки.
func isQuoteClose(r rune) bool {
	return r == '»' || r == '”' || r == '"' || r == '\'' || r == '’'
}

// SplitSentences разбивает абзац на законченные предложения с учетом специфики башкирского языка.
func (s *TextSlicer) SplitSentences(paragraph string) []string {
	paragraph = strings.TrimSpace(paragraph)
	if paragraph == "" {
		return nil
	}

	runes := []rune(paragraph)
	n := len(runes)
	var sentences []string
	start := 0

	for i := 0; i < n; i++ {
		if !isSentenceTerminator(runes[i]) {
			continue
		}

		// Поглощаем серию знаков препинания (..., ?!, !!!, ..)
		punctEnd := i
		for punctEnd < n && (isSentenceTerminator(runes[punctEnd]) || runes[punctEnd] == '.') {
			punctEnd++
		}

		// Поглощаем закрывающие кавычки или скобки
		afterPunct := punctEnd
		for afterPunct < n && (isQuoteClose(runes[afterPunct]) || runes[afterPunct] == ')') {
			afterPunct++
		}

		// Если это конец текста, завершаем предложение
		if afterPunct >= n {
			sentence := NormalizeSpaces(string(runes[start:n]))
			if utf8.RuneCountInString(sentence) >= s.cfg.MinRunes {
				sentences = append(sentences, sentence)
			}
			start = n
			break
		}

		// Проверяем следующий символ после кавычек/пунктуации
		// Если за кавычкой следует запятая или тире с продолжением авторской речи
		// (например: «Һаумыһығыҙ!» — тине ул.), не разрываем
		peekIdx := afterPunct
		for peekIdx < n && unicode.IsSpace(runes[peekIdx]) {
			peekIdx++
		}

		if peekIdx < n {
			// Проверяем прямую речь с продолжением автора: ", —", "—" с маленькой буквы
			if runes[peekIdx] == ',' || runes[peekIdx] == '—' || runes[peekIdx] == '-' {
				// Смотрим дальше за тире/запятую
				nextWordIdx := peekIdx + 1
				for nextWordIdx < n && (unicode.IsSpace(runes[nextWordIdx]) || runes[nextWordIdx] == '—' || runes[nextWordIdx] == '-') {
					nextWordIdx++
				}
				if nextWordIdx < n && unicode.IsLower(runes[nextWordIdx]) {
					// Это авторская реплика типа "— тине", не разделяем!
					i = nextWordIdx
					continue
				}
			}

			// Проверка на инициалы и сокращения
			tokenBefore := extractLastToken(string(runes[start:punctEnd]))
			lowerToken := strings.ToLower(tokenBefore)

			if isInitial(tokenBefore) && unicode.IsUpper(runes[peekIdx]) {
				// Инициал перед фамилией: "Р. Ғарипов"
				i = punctEnd - 1
				continue
			}

			if s.cfg.AbbrevExceptions[lowerToken] {
				// Известное сокращение (например, "һ.б.", "й.", "г.")
				// Если следующее слово с маленькой буквы — точно не конец предложения
				if unicode.IsLower(runes[peekIdx]) {
					i = punctEnd - 1
					continue
				}
			}

			// Проверка на порядковые числительные типа "1. ", "2. "
			if isAllDigits(strings.TrimSuffix(tokenBefore, ".")) && unicode.IsLower(runes[peekIdx]) {
				i = punctEnd - 1
				continue
			}
		}

		// Если разделитель валиден (после него пробел или новое предложение с большой буквы)
		if unicode.IsSpace(runes[afterPunct]) || afterPunct == n {
			sentence := NormalizeSpaces(string(runes[start:afterPunct]))
			if utf8.RuneCountInString(sentence) >= s.cfg.MinRunes {
				sentences = append(sentences, sentence)
			}
			start = afterPunct
			// Пропускаем пробелы в начале следующего предложения
			for start < n && unicode.IsSpace(runes[start]) {
				start++
			}
			i = start - 1
		}
	}

	if start < n {
		remaining := NormalizeSpaces(string(runes[start:n]))
		if utf8.RuneCountInString(remaining) >= s.cfg.MinRunes {
			sentences = append(sentences, remaining)
		}
	}

	return sentences
}

// SliceText нарезает сырой текст на параграфы и предложения с координатами.
func (s *TextSlicer) SliceText(text string, chapterInfo string) []TextFragment {
	paragraphs := SplitParagraphs(text)
	var fragments []TextFragment

	globalSentenceIdx := 1
	for pIdx, p := range paragraphs {
		sentences := s.SplitSentences(p)
		for _, sent := range sentences {
			fragments = append(fragments, TextFragment{
				ParagraphIdx: pIdx + 1,
				SentenceIdx:  globalSentenceIdx,
				Text:         sent,
				ChapterInfo:  chapterInfo,
			})
			globalSentenceIdx++
		}
	}

	return fragments
}

// Helper: извлечь последнее слово перед знаком препинания
func extractLastToken(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ==========================================
// СЕГМЕНТАТОР АУДИОМАТЕРИАЛОВ (Audio Segmenter)
// ==========================================

// AudioSegmenter хелпер для разбивки аудиокниг и радиопередач на фрагменты.
type AudioSegmenter struct {
	defaultChunkMS int
}

// NewAudioSegmenter конструктор сегментатора аудио.
func NewAudioSegmenter(defaultChunkMS ...int) *AudioSegmenter {
	chunk := 10000 // По умолчанию 10 секунд
	if len(defaultChunkMS) > 0 && defaultChunkMS[0] > 0 {
		chunk = defaultChunkMS[0]
	}
	return &AudioSegmenter{defaultChunkMS: chunk}
}

// SliceAudio нарезает аудиокнигу/передачу на фрагменты с точными таймкодами (audio_start_ms, audio_end_ms).
func (as *AudioSegmenter) SliceAudio(meta AudioMetadata) []AudioFragment {
	var fragments []AudioFragment

	chapter := meta.Chapter
	if chapter == "" {
		chapter = "Основной раздел"
	}

	// 1. Если переданы точные разметки (Cues / Subtitles / VAD)
	if len(meta.Cues) > 0 {
		for idx, cue := range meta.Cues {
			ch := cue.ChapterInfo
			if ch == "" {
				ch = chapter
			}
			fragments = append(fragments, AudioFragment{
				SourceTitle:  meta.SourceTitle,
				ChapterInfo:  ch,
				AudioURL:     meta.AudioURL,
				AudioStartMS: cue.StartMS,
				AudioEndMS:   cue.EndMS,
				TextBa:       cue.TextBa,
				SentenceIdx:  idx + 1,
				ParagraphIdx: 1,
			})
		}
		return fragments
	}

	// 2. Если разметка отсутствует, нарезаем аудиопоток по интервалам длительности
	chunk := as.defaultChunkMS
	if meta.DurationMS <= 0 {
		// Одиночный фрагмент без известной длины
		return []AudioFragment{
			{
				SourceTitle:  meta.SourceTitle,
				ChapterInfo:  chapter,
				AudioURL:     meta.AudioURL,
				AudioStartMS: 0,
				AudioEndMS:   0,
				SentenceIdx:  1,
				ParagraphIdx: 1,
			},
		}
	}

	currentStart := 0
	sentenceIdx := 1
	for currentStart < meta.DurationMS {
		end := currentStart + chunk
		if end > meta.DurationMS {
			end = meta.DurationMS
		}

		fragments = append(fragments, AudioFragment{
			SourceTitle:  meta.SourceTitle,
			ChapterInfo:  chapter,
			AudioURL:     meta.AudioURL,
			AudioStartMS: currentStart,
			AudioEndMS:   end,
			SentenceIdx:  sentenceIdx,
			ParagraphIdx: 1,
		})

		currentStart = end
		sentenceIdx++
	}

	return fragments
}

// ==========================================
// ГЕНЕРАТОР АНТИЧИТ-ЛОВУШЕК (Honeypot Traps)
// ==========================================

var antonymMap = map[string]string{
	"хорошо":     "плохо",
	"плохо":      "хорошо",
	"да":         "нет",
	"нет":        "да",
	"всегда":     "никогда",
	"никогда":    "всегда",
	"можно":      "нельзя",
	"нельзя":     "можно",
	"друг":       "враг",
	"враг":       "друг",
	"день":       "ночь",
	"ночь":       "день",
	"большой":    "маленький",
	"маленький":  "большой",
	"много":      "мало",
	"мало":       "много",
	"новый":      "старый",
	"старый":     "новый",
	"быстро":     "медленно",
	"медленно":   "быстро",
	"правда":     "ложь",
	"ложь":       "правда",
	"добрый":     "злой",
	"злой":       "добрый",
	"свет":       "тьма",
	"тепло":      "холодно",
	"холодно":    "тепло",
	"любит":      "ненавидит",
	"ненавидит":  "любит",
	"белый":      "чёрный",
	"чёрный":     "белый",
	"начало":     "конец",
	"конец":      "начало",
	"красивый":   "безобразный",
	"богатый":    "бедный",
	"бедный":     "богатый",
	"радость":    "печаль",
	"печаль":     "радость",
	"умный":      "глупый",
	"глупый":     "умный",
}

// GenerateHoneypot искажает перевод для создания контрольной ловушки античита.
// Использует контекстную подмену или антонимическую инверсию смысла.
func GenerateHoneypot(originalTranslation string, alternativeDistractor string) (bool, string) {
	orig := strings.TrimSpace(originalTranslation)
	if orig == "" {
		return false, ""
	}

	// 1. Попытка инвертировать смысл с помощью словаря антонимов
	words := strings.Fields(orig)
	for i, w := range words {
		clean := strings.ToLower(strings.Trim(w, ".,!?:;«»\"'"))
		if ant, ok := antonymMap[clean]; ok {
			// Заменяем антоним с сохранением регистра
			repl := ant
			if len(w) > 0 && unicode.IsUpper([]rune(w)[0]) {
				runes := []rune(repl)
				runes[0] = unicode.ToUpper(runes[0])
				repl = string(runes)
			}
			words[i] = repl
			distorted := strings.Join(words, " ")
			return true, distorted
		}
	}

	// 2. Инверсия частицы "не "
	if strings.Contains(orig, "не ") {
		distorted := strings.Replace(orig, "не ", "", 1)
		return true, distorted
	} else if strings.Contains(orig, "Не ") {
		distorted := strings.Replace(orig, "Не ", "", 1)
		return true, distorted
	}

	// 3. Подмена на альтернативную фразу из контекста (семантический дистрактор)
	if alternativeDistractor != "" && alternativeDistractor != orig {
		return true, alternativeDistractor
	}

	// 4. Дефолтное искажение (вставка смысловой ошибки)
	return true, "[ИСКАЖЕНО] " + orig
}

// ShouldInjectHoneypot определяет, должен ли фрагмент стать ловушкой (по умолчанию 20% = каждый 5-й).
func ShouldInjectHoneypot(index int, rate float64) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 1.0 {
		return true
	}
	// Например rate 0.20 -> каждые 5 элементов
	interval := int(1.0 / rate)
	if interval < 1 {
		interval = 5
	}
	return (index % interval) == 0
}

// RandomChoice вспомогательный хелпер случайного выбора.
func RandomChoice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[rand.Intn(len(items))]
}

// CompileRegexSafety безопасный компилятор регулярок
var whitespaceRegex = regexp.MustCompile(`\s+`)
