package moderation

import (
	"strings"
	"unicode"
)

var (
	// Корни и основы нецензурных слов (русский и заимствованный мат)
	profaneRoots = []string{
		"хуй", "хуе", "хуя", "хуи", "хул", "охуе", "охуи", "ахуе", "ахуи",
		"пизд", "ебат", "ебан", "ебл", "ебут", "еби", "заеб", "выеб", "доеб", "наеб", "поеб", "проеб", "уеб",
		"ёбан", "ёбл", "бляд", "блят", "сука", "сучк", "мудак", "мудил", "гондон",
		"гандон", "пидор", "педик", "залуп", "чмо", "дерьм", "говно", "говн",
		"дроч", "срать", "сран", "засра",
	}
)

func hasRepeatedChars(s string, maxRepeat int) bool {
	runes := []rune(s)
	count := 1
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// normalizeText заменяет похожие латинские символы и цифры на кириллицу для обхода фильтра
func normalizeText(s string) string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(
		"a", "а", "@", "а",
		"b", "б", "6", "б",
		"v", "в",
		"g", "г",
		"d", "д",
		"e", "е", "ё", "е",
		"z", "з", "3", "з",
		"i", "и", "1", "и", "!", "и",
		"k", "к",
		"l", "л",
		"m", "м",
		"n", "н",
		"o", "о", "0", "о",
		"p", "п",
		"r", "р",
		"s", "с", "$", "с",
		"t", "т",
		"u", "у", "y", "у",
		"f", "ф",
		"h", "х", "x", "х",
		"c", "ц",
		"ch", "ч", "4", "ч",
		"sh", "ш",
	)
	s = replacer.Replace(s)

	// Убираем все пробелы и спецсимволы для поиска скрытого мата (например, "х у й")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ContainsProfanity проверяет, содержит ли текст нецензурную брань
func ContainsProfanity(text string) bool {
	clean := normalizeText(text)
	if clean == "" {
		return false
	}

	for _, root := range profaneRoots {
		if strings.Contains(clean, root) {
			return true
		}
	}
	return false
}

// ValidateTranslationQuality проверяет адекватность предложенного текста перевода
func ValidateTranslationQuality(text string) (bool, string) {
	trimmed := strings.TrimSpace(text)
	if len([]rune(trimmed)) < 2 {
		return false, "Перевод слишком короткий"
	}

	if hasRepeatedChars(trimmed, 5) {
		return false, "Обнаружено бессмысленное повторение символов"
	}

	if ContainsProfanity(trimmed) {
		return false, "Обнаружена ненормативная лексика"
	}

	// Проверяем, что в тексте есть хотя бы одна буква
	hasLetter := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return false, "Перевод должен содержать слова"
	}

	return true, ""
}
