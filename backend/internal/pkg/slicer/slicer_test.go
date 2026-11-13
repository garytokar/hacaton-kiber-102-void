package slicer

import (
	"strings"
	"testing"
)

func TestSplitParagraphs(t *testing.T) {
	raw := `Беренсе абзац. Бында башкирса һүҙҙәр бар.

Икенсе абзац.
Дауамы бында.

Өсөнсө абзац.`

	paragraphs := SplitParagraphs(raw)
	if len(paragraphs) != 3 {
		t.Fatalf("Ожидалось 3 абзаца, получено %d", len(paragraphs))
	}

	if paragraphs[0] != "Беренсе абзац. Бында башкирса һүҙҙәр бар." {
		t.Errorf("Неверный первый абзац: %s", paragraphs[0])
	}
	if paragraphs[1] != "Икенсе абзац. Дауамы бында." {
		t.Errorf("Неверный второй абзац: %s", paragraphs[1])
	}
	if paragraphs[2] != "Өсөнсө абзац." {
		t.Errorf("Неверный третий абзац: %s", paragraphs[2])
	}
}

func TestSplitSentences_Basic(t *testing.T) {
	s := NewTextSlicer()
	text := "Башҡортостан — беҙҙең тыуған төйәгебеҙ. Ҡояш ҡалҡты! Ауыл өҫтөнә яҡтылыҡ төштө. Һин ҡайҙа бараһың?"

	sentences := s.SplitSentences(text)
	if len(sentences) != 4 {
		t.Fatalf("Ожидалось 4 предложения, получено %d: %v", len(sentences), sentences)
	}

	expected := []string{
		"Башҡортостан — беҙҙең тыуған төйәгебеҙ.",
		"Ҡояш ҡалҡты!",
		"Ауыл өҫтөнә яҡтылыҡ төштө.",
		"Һин ҡайҙа бараһың?",
	}

	for i, exp := range expected {
		if sentences[i] != exp {
			t.Errorf("[%d] Ожидалось: %q, Получено: %q", i, exp, sentences[i])
		}
	}
}

func TestSplitSentences_InitialsAndAbbrev(t *testing.T) {
	s := NewTextSlicer()
	text := "М. Кәрим яҙған әҫәр бик матур. Китапта 1976 й. ваҡиғалары, мәктәп һәм һ.б. тураһында һөйләнелә. Р. Ғарипов шиғырҙары онотолмай."

	sentences := s.SplitSentences(text)
	// Должно быть 3 предложения:
	// 1: "М. Кәрим яҙған әҫәр бик матур."
	// 2: "Китапта 1976 й. ваҡиғалары, мәктәп һәм һ.б. тураһында һөйләнелә."
	// 3: "Р. Ғарипов шиғырҙары онотолмай."
	if len(sentences) != 3 {
		t.Fatalf("Ожидалось 3 предложения, получено %d: %v", len(sentences), sentences)
	}

	if !strings.HasPrefix(sentences[0], "М. Кәрим") {
		t.Errorf("Ошибка с инициалом М. Кәрим: %s", sentences[0])
	}
	if !strings.Contains(sentences[1], "1976 й.") || !strings.Contains(sentences[1], "һ.б.") {
		t.Errorf("Ошибка с сокращениями й. или һ.б.: %s", sentences[1])
	}
	if !strings.HasPrefix(sentences[2], "Р. Ғарипов") {
		t.Errorf("Ошибка с инициалом Р. Ғарипов: %s", sentences[2])
	}
}

func TestSplitSentences_DirectSpeech(t *testing.T) {
	s := NewTextSlicer()
	text := "«Һаумыһығыҙ, дуҫтар!» — тине уҡытыусы. Барыһы ла баҫты."

	sentences := s.SplitSentences(text)
	if len(sentences) != 2 {
		t.Fatalf("Ожидалось 2 предложения (прямая речь + следующее), получено %d: %v", len(sentences), sentences)
	}

	if sentences[0] != "«Һаумыһығыҙ, дуҫтар!» — тине уҡытыусы." {
		t.Errorf("Прямая речь разорвана: %q", sentences[0])
	}
}

func TestSplitSentences_EllipsisAndExclamation(t *testing.T) {
	s := NewTextSlicer()
	text := "Ул оҙаҡ уйланды... Һуңынан: «Бәлки?!» — тип ҡысҡырҙы."

	sentences := s.SplitSentences(text)
	if len(sentences) != 2 {
		t.Fatalf("Ожидалось 2 предложения, получено %d: %v", len(sentences), sentences)
	}
	if sentences[0] != "Ул оҙаҡ уйланды..." {
		t.Errorf("Многоточие обработано неверно: %s", sentences[0])
	}
}

func TestSliceText_Coordinates(t *testing.T) {
	s := NewTextSlicer()
	raw := `Беренсе һөйләм. Икенсе һөйләм.

Өсөнсө һөйләм.`

	fragments := s.SliceText(raw, "Глава 1: Бала саҡ")
	if len(fragments) != 3 {
		t.Fatalf("Ожидалось 3 фрагмента, получено %d", len(fragments))
	}

	// 1-й фрагмент
	if fragments[0].ParagraphIdx != 1 || fragments[0].SentenceIdx != 1 || fragments[0].Text != "Беренсе һөйләм." {
		t.Errorf("Неверный фрагмент 0: %+v", fragments[0])
	}
	if fragments[0].ChapterInfo != "Глава 1: Бала саҡ" {
		t.Errorf("Неверная глава: %s", fragments[0].ChapterInfo)
	}

	// 2-й фрагмент
	if fragments[1].ParagraphIdx != 1 || fragments[1].SentenceIdx != 2 || fragments[1].Text != "Икенсе һөйләм." {
		t.Errorf("Неверный фрагмент 1: %+v", fragments[1])
	}

	// 3-й фрагмент во 2-м параграфе
	if fragments[2].ParagraphIdx != 2 || fragments[2].SentenceIdx != 3 || fragments[2].Text != "Өсөнсө һөйләм." {
		t.Errorf("Неверный фрагмент 2: %+v", fragments[2])
	}
}

func TestAudioSegmenter_WithCues(t *testing.T) {
	segmenter := NewAudioSegmenter()

	meta := AudioMetadata{
		SourceTitle: "Оҙон-оҙаҡ бала саҡ (Мустай Карим)",
		Chapter:     "Глава 1. Кендек",
		AudioURL:    "https://storage.yandexcloud.net/bashkir-audiobooks/karim_ch1.mp3",
		DurationMS:  30000,
		Cues: []AudioCue{
			{
				StartMS:     0,
				EndMS:       4500,
				TextBa:      "Кендегемде минең ҡартая әбейем киҫкән.",
				ChapterInfo: "Глава 1. Кендек",
			},
			{
				StartMS:     4500,
				EndMS:       9200,
				TextBa:      "Ул мине тәү башлап ҡулына алған.",
				ChapterInfo: "Глава 1. Кендек",
			},
		},
	}

	fragments := segmenter.SliceAudio(meta)
	if len(fragments) != 2 {
		t.Fatalf("Ожидалось 2 аудиофрагмента, получено %d", len(fragments))
	}

	if fragments[0].AudioStartMS != 0 || fragments[0].AudioEndMS != 4500 {
		t.Errorf("Неверные таймкоды 1: start=%d end=%d", fragments[0].AudioStartMS, fragments[0].AudioEndMS)
	}
	if fragments[0].TextBa != "Кендегемде минең ҡартая әбейем киҫкән." {
		t.Errorf("Неверный текст 1: %s", fragments[0].TextBa)
	}
	if fragments[1].AudioStartMS != 4500 || fragments[1].AudioEndMS != 9200 {
		t.Errorf("Неверные таймкоды 2: start=%d end=%d", fragments[1].AudioStartMS, fragments[1].AudioEndMS)
	}
}

func TestAudioSegmenter_DurationChunking(t *testing.T) {
	segmenter := NewAudioSegmenter(10000) // 10-секундные отрезки

	meta := AudioMetadata{
		SourceTitle: "Радио Башкортостана: Яңылыҡтар",
		Chapter:     "Выпуск 12:00",
		AudioURL:    "https://audio.bashkort.org/news_1200.mp3",
		DurationMS:  25000, // 25 секунд = 3 отрезка (0-10с, 10-20с, 20-25с)
	}

	fragments := segmenter.SliceAudio(meta)
	if len(fragments) != 3 {
		t.Fatalf("Ожидалось 3 фрагмента по длительности, получено %d", len(fragments))
	}

	if fragments[0].AudioStartMS != 0 || fragments[0].AudioEndMS != 10000 {
		t.Errorf("Фрагмент 1: start=%d end=%d", fragments[0].AudioStartMS, fragments[0].AudioEndMS)
	}
	if fragments[1].AudioStartMS != 10000 || fragments[1].AudioEndMS != 20000 {
		t.Errorf("Фрагмент 2: start=%d end=%d", fragments[1].AudioStartMS, fragments[1].AudioEndMS)
	}
	if fragments[2].AudioStartMS != 20000 || fragments[2].AudioEndMS != 25000 {
		t.Errorf("Фрагмент 3: start=%d end=%d", fragments[2].AudioStartMS, fragments[2].AudioEndMS)
	}
}

func TestGenerateHoneypot(t *testing.T) {
	// 1. Антоним
	hasH, variant := GenerateHoneypot("Всё очень хорошо сегодня.", "")
	if !hasH || !strings.Contains(variant, "плохо") {
		t.Errorf("Ожидалась замена хорошо -> плохо, получено: %s", variant)
	}

	// 2. Инверсия "не"
	hasH2, variant2 := GenerateHoneypot("Он не знал ответа на вопрос.", "")
	if !hasH2 || strings.Contains(variant2, "не ") {
		t.Errorf("Ожидалось удаление 'не ', получено: %s", variant2)
	}

	// 3. Дистрактор
	hasH3, variant3 := GenerateHoneypot("Яблоко красное.", "Лошадь бежит по полю.")
	if !hasH3 || variant3 != "Лошадь бежит по полю." {
		t.Errorf("Ожидался дистрактор, получено: %s", variant3)
	}
}

func TestShouldInjectHoneypot(t *testing.T) {
	rate := 0.20 // 20%
	count := 0
	total := 100

	for i := 1; i <= total; i++ {
		if ShouldInjectHoneypot(i, rate) {
			count++
		}
	}

	// Из 100 записей должно быть ровно 20 ловушек (каждая 5-я)
	if count != 20 {
		t.Errorf("Ожидалось 20 ловушек из 100, получено %d", count)
	}
}
