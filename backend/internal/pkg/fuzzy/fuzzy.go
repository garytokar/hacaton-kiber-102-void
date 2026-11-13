package fuzzy

import (
	"sort"
	"strings"
	"unicode"

	"bashkir-backend/internal/domain"
)

// DefaultSimilarityThreshold is the default similarity threshold for clustering submissions.
const DefaultSimilarityThreshold = 0.82

// Canonicalize normalizes text for fuzzy comparison:
// - Preserves sentinel values like "[ERROR_REPORTED]"
// - Converts to lowercase (supporting Cyrillic, Bashkir, and Latin alphabets)
// - Normalizes various Unicode dash forms to a standard hyphen '-'
// - Normalizes various Unicode quote forms to standard double quotes '"'
// - Collapses multiple whitespace characters to a single space
// - Trims leading and trailing punctuation and whitespace
func Canonicalize(s string) string {
	trimmed := strings.TrimSpace(s)
	if strings.EqualFold(trimmed, "[ERROR_REPORTED]") {
		return "[ERROR_REPORTED]"
	}

	lower := strings.ToLower(trimmed)

	var sb strings.Builder
	sb.Grow(len(lower))

	lastWasSpace := false
	for _, r := range lower {
		// Normalize dashes
		switch r {
		case '—', '–', '―', '−', '‒', '‐':
			r = '-'
		// Normalize quotes
		case '«', '»', '“', '”', '„', '‘', '’', '‚', '`', '\'':
			r = '"'
		}

		if unicode.IsSpace(r) {
			if !lastWasSpace && sb.Len() > 0 {
				sb.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			sb.WriteRune(r)
			lastWasSpace = false
		}
	}

	result := sb.String()

	// Trim leading and trailing punctuation & spaces
	isPunctOrSpace := func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		}
		switch r {
		case '.', ',', '!', '?', ';', ':', '"', '\'', '-', '(', ')', '[', ']', '{', '}', '<', '>', '/', '\\', '…', '`':
			return true
		default:
			return unicode.IsPunct(r)
		}
	}

	result = strings.TrimFunc(result, isPunctOrSpace)
	return strings.TrimSpace(result)
}

// LevenshteinDistance calculates the edit distance between two strings using Unicode runes.
func LevenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	if n < m {
		r1, r2 = r2, r1
		n, m = m, n
	}

	row := make([]int, m+1)
	for j := 0; j <= m; j++ {
		row[j] = j
	}

	for i := 1; i <= n; i++ {
		prev := row[0]
		row[0] = i
		for j := 1; j <= m; j++ {
			temp := row[j]
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			val := row[j] + 1
			if prev+cost < val {
				val = prev + cost
			}
			if row[j-1]+1 < val {
				val = row[j-1] + 1
			}
			row[j] = val
			prev = temp
		}
	}

	return row[m]
}

// LevenshteinSimilarity calculates normalized Levenshtein similarity between 0.0 and 1.0.
func LevenshteinSimilarity(s1, s2 string) float64 {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 && m == 0 {
		return 1.0
	}
	if n == 0 || m == 0 {
		return 0.0
	}

	maxLen := n
	if m > maxLen {
		maxLen = m
	}

	dist := LevenshteinDistance(s1, s2)
	sim := 1.0 - float64(dist)/float64(maxLen)
	if sim < 0 {
		return 0
	}
	return sim
}

// TrigramSimilarity calculates the similarity between two strings using character trigrams (Dice coefficient).
// Returns a value between 0.0 and 1.0.
func TrigramSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	r1, r2 := []rune(s1), []rune(s2)
	if len(r1) == 0 && len(r2) == 0 {
		return 1.0
	}
	if len(r1) == 0 || len(r2) == 0 {
		return 0.0
	}

	t1 := extractTrigrams(r1)
	t2 := extractTrigrams(r2)

	if len(t1) == 0 && len(t2) == 0 {
		return 1.0
	}
	if len(t1) == 0 || len(t2) == 0 {
		return 0.0
	}

	counts1 := make(map[string]int, len(t1))
	for _, t := range t1 {
		counts1[t]++
	}

	intersection := 0
	for _, t := range t2 {
		if count, ok := counts1[t]; ok && count > 0 {
			intersection++
			counts1[t]--
		}
	}

	return float64(2*intersection) / float64(len(t1)+len(t2))
}

func extractTrigrams(runes []rune) []string {
	if len(runes) == 0 {
		return nil
	}
	// Pad with two spaces before and one space after (consistent with PostgreSQL pg_trgm approach)
	padded := make([]rune, 0, len(runes)+3)
	padded = append(padded, ' ', ' ')
	padded = append(padded, runes...)
	padded = append(padded, ' ')

	trigrams := make([]string, 0, len(padded)-2)
	for i := 0; i <= len(padded)-3; i++ {
		trigrams = append(trigrams, string(padded[i:i+3]))
	}
	return trigrams
}

func extractDigits(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Similarity calculates the overall fuzzy similarity between two strings,
// applying text canonicalization, ensuring numerical facts match,
// and combining Levenshtein and Trigram similarities.
func Similarity(s1, s2 string) float64 {
	c1 := Canonicalize(s1)
	c2 := Canonicalize(s2)
	if c1 == c2 {
		return 1.0
	}
	if len(c1) == 0 || len(c2) == 0 {
		return 0.0
	}

	// Digits represent quantitative facts (numbers, dates, chapter numbers, etc.).
	// If digits differ, they cannot be considered the same translation variant.
	if extractDigits(c1) != extractDigits(c2) {
		return 0.0
	}

	levSim := LevenshteinSimilarity(c1, c2)
	triSim := TrigramSimilarity(c1, c2)

	return (levSim + triSim) / 2.0
}

// ClusterSubmissions takes raw submission groups and clusters similar texts together
// into a single dominant representative text, summing their vote counts.
// Groups with similarity >= similarityThreshold are merged.
func ClusterSubmissions(submissions []domain.SubmissionGroup, similarityThreshold float64) []domain.SubmissionGroup {
	if len(submissions) <= 1 {
		return submissions
	}
	if similarityThreshold <= 0 {
		similarityThreshold = DefaultSimilarityThreshold
	}

	// Sort a copy by Count DESC so dominant variants become cluster centers
	subs := make([]domain.SubmissionGroup, len(submissions))
	copy(subs, submissions)
	sort.SliceStable(subs, func(i, j int) bool {
		if subs[i].Count != subs[j].Count {
			return subs[i].Count > subs[j].Count
		}
		// If counts are equal, prefer longer text (often preserves proper formatting)
		return len(subs[i].Text) > len(subs[j].Text)
	})

	type cluster struct {
		representative string
		canonicalRep   string
		count          int
		isErrorReport  bool
	}

	var clusters []*cluster

	for _, sub := range subs {
		trimmed := strings.TrimSpace(sub.Text)
		isErr := strings.EqualFold(trimmed, "[ERROR_REPORTED]")

		if isErr {
			found := false
			for _, c := range clusters {
				if c.isErrorReport {
					c.count += sub.Count
					found = true
					break
				}
			}
			if !found {
				clusters = append(clusters, &cluster{
					representative: "[ERROR_REPORTED]",
					canonicalRep:   "[ERROR_REPORTED]",
					count:          sub.Count,
					isErrorReport:  true,
				})
			}
			continue
		}

		canon := Canonicalize(sub.Text)

		// Find best matching existing cluster
		bestIdx := -1
		bestSim := -1.0
		for i, c := range clusters {
			if c.isErrorReport {
				continue
			}

			// Check canonical match first
			if c.canonicalRep == canon && canon != "" {
				bestSim = 1.0
				bestIdx = i
				break
			}

			sim := Similarity(sub.Text, c.representative)
			if sim > bestSim {
				bestSim = sim
				bestIdx = i
			}
		}

		if bestIdx >= 0 && bestSim >= similarityThreshold {
			clusters[bestIdx].count += sub.Count
		} else {
			clusters = append(clusters, &cluster{
				representative: sub.Text,
				canonicalRep:   canon,
				count:          sub.Count,
				isErrorReport:  false,
			})
		}
	}

	// Sort clusters by count DESC
	sort.SliceStable(clusters, func(i, j int) bool {
		return clusters[i].count > clusters[j].count
	})

	result := make([]domain.SubmissionGroup, len(clusters))
	for i, c := range clusters {
		result[i] = domain.SubmissionGroup{
			Text:  c.representative,
			Count: c.count,
		}
	}

	return result
}
