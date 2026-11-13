package fuzzy_test

import (
	"testing"

	"bashkir-backend/internal/domain"
	"bashkir-backend/internal/pkg/fuzzy"
)

func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase and trim",
			input:    "  Башкортостан — НАШ край.  ",
			expected: "башкортостан - наш край",
		},
		{
			name:     "bashkir specific characters",
			input:    "Башҡортостан — беҙҙең тыуған төйәгебеҙ!",
			expected: "башҡортостан - беҙҙең тыуған төйәгебеҙ",
		},
		{
			name:     "normalize various dashes",
			input:    "Слово—тире –дефис―горизонталь−минус",
			expected: "слово-тире -дефис-горизонталь-минус",
		},
		{
			name:     "normalize quotes and trim boundary punctuation",
			input:    "«Книга — источник знаний!»",
			expected: "книга - источник знаний",
		},
		{
			name:     "multiple spaces collapse",
			input:    "Книга    источник   знаний",
			expected: "книга источник знаний",
		},
		{
			name:     "special sentinel ERROR_REPORTED",
			input:    " [ERROR_REPORTED] ",
			expected: "[ERROR_REPORTED]",
		},
		{
			name:     "case-insensitive sentinel error_reported",
			input:    "[error_reported]",
			expected: "[ERROR_REPORTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzy.Canonicalize(tt.input)
			if got != tt.expected {
				t.Errorf("Canonicalize(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "abc", 3},
		{"книга", "книга", 0},
		{"книга", "книги", 1},
		{"китап", "китаптар", 3},
		{"башкортостан", "башкортостан", 0},
	}

	for _, tt := range tests {
		got := fuzzy.LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, got, tt.expected)
		}
	}
}

func TestLevenshteinSimilarity(t *testing.T) {
	if sim := fuzzy.LevenshteinSimilarity("книга", "книга"); sim != 1.0 {
		t.Errorf("expected 1.0 for identical strings, got %f", sim)
	}
	if sim := fuzzy.LevenshteinSimilarity("", ""); sim != 1.0 {
		t.Errorf("expected 1.0 for empty strings, got %f", sim)
	}
	if sim := fuzzy.LevenshteinSimilarity("книга", ""); sim != 0.0 {
		t.Errorf("expected 0.0 when one string is empty, got %f", sim)
	}
	// "книга" vs "книги" (1 diff in 5 runes = 0.8)
	if sim := fuzzy.LevenshteinSimilarity("книга", "книги"); sim != 0.8 {
		t.Errorf("expected 0.8, got %f", sim)
	}
}

func TestTrigramSimilarity(t *testing.T) {
	if sim := fuzzy.TrigramSimilarity("книга", "книга"); sim != 1.0 {
		t.Errorf("expected 1.0 for identical strings, got %f", sim)
	}
	if sim := fuzzy.TrigramSimilarity("", ""); sim != 1.0 {
		t.Errorf("expected 1.0 for empty strings, got %f", sim)
	}
	if sim := fuzzy.TrigramSimilarity("книга", ""); sim != 0.0 {
		t.Errorf("expected 0.0 when one is empty, got %f", sim)
	}

	// Trigram should capture word reordering well
	reorderedSim := fuzzy.TrigramSimilarity("наш родной край", "родной наш край")
	if reorderedSim < 0.75 {
		t.Errorf("expected trigram similarity for reordered words >= 0.75, got %f", reorderedSim)
	}

	// Completely unrelated strings
	disjointSim := fuzzy.TrigramSimilarity("башкортостан", "автомобиль")
	if disjointSim > 0.3 {
		t.Errorf("expected low similarity for unrelated strings, got %f", disjointSim)
	}
}

func TestSimilarity_PunctuationVariations(t *testing.T) {
	// Task requirement:
	// "Книга — источник знаний.", "Книга - источник знаний", "Книга источник знаний!"
	s1 := "Книга — источник знаний."
	s2 := "Книга - источник знаний"
	s3 := "Книга источник знаний!"

	sim12 := fuzzy.Similarity(s1, s2)
	if sim12 < 0.99 {
		t.Errorf("expected similarity(s1, s2) close to 1.0, got %f", sim12)
	}

	sim13 := fuzzy.Similarity(s1, s3)
	if sim13 < 0.82 {
		t.Errorf("expected similarity(s1, s3) >= 0.82, got %f", sim13)
	}

	sim23 := fuzzy.Similarity(s2, s3)
	if sim23 < 0.82 {
		t.Errorf("expected similarity(s2, s3) >= 0.82, got %f", sim23)
	}
}

func TestClusterSubmissions(t *testing.T) {
	t.Run("merges differing punctuation variants into dominant cluster", func(t *testing.T) {
		input := []domain.SubmissionGroup{
			{Text: "Книга — источник знаний.", Count: 1},
			{Text: "Книга - источник знаний", Count: 1},
			{Text: "Книга источник знаний!", Count: 1},
		}

		clusters := fuzzy.ClusterSubmissions(input, 0.82)
		if len(clusters) != 1 {
			t.Fatalf("expected 1 cluster, got %d: %+v", len(clusters), clusters)
		}

		if clusters[0].Count != 3 {
			t.Errorf("expected cluster count 3, got %d", clusters[0].Count)
		}
	})

	t.Run("keeps distinct translations in separate clusters", func(t *testing.T) {
		input := []domain.SubmissionGroup{
			{Text: "Книга — источник знаний.", Count: 2},
			{Text: "Собака бежит по улице.", Count: 2},
		}

		clusters := fuzzy.ClusterSubmissions(input, 0.82)
		if len(clusters) != 2 {
			t.Fatalf("expected 2 clusters, got %d", len(clusters))
		}
		if clusters[0].Count != 2 || clusters[1].Count != 2 {
			t.Errorf("expected both clusters to have count 2, got %d and %d", clusters[0].Count, clusters[1].Count)
		}
	})

	t.Run("handles error reports separately without merging into text", func(t *testing.T) {
		input := []domain.SubmissionGroup{
			{Text: "[ERROR_REPORTED]", Count: 2},
			{Text: "[error_reported]", Count: 1},
			{Text: "Хороший перевод текста", Count: 2},
		}

		clusters := fuzzy.ClusterSubmissions(input, 0.82)
		if len(clusters) != 2 {
			t.Fatalf("expected 2 clusters (1 error, 1 text), got %d", len(clusters))
		}

		// First should be error report with count 3
		if clusters[0].Text != "[ERROR_REPORTED]" || clusters[0].Count != 3 {
			t.Errorf("expected first cluster to be [ERROR_REPORTED] with count 3, got %s (%d)", clusters[0].Text, clusters[0].Count)
		}
	})
}
