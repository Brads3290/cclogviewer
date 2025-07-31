package processor

import (
	"strings"
	"unicode"
)

// EstimateTokens approximates token count using word-based heuristics.
func EstimateTokens(text string) int {
	// Remove HTML tags for more accurate counting
	cleaned := stripHTMLTags(text)

	// Simple approximation: count words and divide by typical token/word ratio
	words := countWords(cleaned)

	// On average, 1 word ≈ 1.3 tokens for English text
	// This is a rough approximation that works reasonably well
	return int(float64(words) * 1.3)
}

func stripHTMLTags(html string) string {
	// Simple HTML tag removal
	result := html
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + " " + result[start+end+1:]
	}
	return result
}

func countWords(text string) int {
	count := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		count++
	}

	return count
}
