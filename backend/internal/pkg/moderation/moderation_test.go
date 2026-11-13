package moderation_test

import (
	"testing"

	"bashkir-backend/internal/pkg/moderation"
)

func TestModeration_ProfanityDetection(t *testing.T) {
	badWords := []string{
		"хуй",
		"какой-то хуй",
		"пизда",
		"ебать",
		"заебись",
		"блядь",
		"сука",
		"х у й",
		"xуй", // латинская x
		"залупа",
	}

	for _, w := range badWords {
		if !moderation.ContainsProfanity(w) {
			t.Errorf("слово '%s' должно было определиться как ненормативная лексика", w)
		}
		valid, reason := moderation.ValidateTranslationQuality(w)
		if valid {
			t.Errorf("слово '%s' не должно проходить валидацию качества (причина: %s)", w, reason)
		}
	}
}

func TestModeration_ValidTranslations(t *testing.T) {
	goodTranslations := []string{
		"Башкортостан — наша родина.",
		"Книга — источник знаний.",
		"Дружба — самое большое богатство в жизни.",
		"Солнце село за гору, на деревню опустилась тишина.",
	}

	for _, s := range goodTranslations {
		if moderation.ContainsProfanity(s) {
			t.Errorf("корректный текст ошибочно помечен как мат: '%s'", s)
		}
		valid, reason := moderation.ValidateTranslationQuality(s)
		if !valid {
			t.Errorf("корректный текст не прошел валидацию: '%s' (%s)", s, reason)
		}
	}
}

func TestModeration_SpamAndRepetitions(t *testing.T) {
	spams := []string{
		"а",
		"ааааааааааааааа",
		".....",
		"12345",
	}

	for _, s := range spams {
		valid, _ := moderation.ValidateTranslationQuality(s)
		if valid {
			t.Errorf("спам '%s' не должен проходить валидацию качества", s)
		}
	}
}
