package ansi

import (
	"regexp"
	"strconv"
	"strings"
)

// TokenType represents the type of ANSI token
type TokenType int

const (
	// TokenText represents plain text content
	TokenText TokenType = iota
	// TokenEscapeSequence represents an ANSI escape sequence
	TokenEscapeSequence
)

// ANSIToken represents a parsed token from ANSI input
type ANSIToken struct {
	Type    TokenType
	Content string
	Codes   []int
}

// ANSIParser parses ANSI escape sequences from text
type ANSIParser struct {
	escapeRegex *regexp.Regexp
}

// NewANSIParser creates a new ANSI parser
func NewANSIParser() *ANSIParser {
	return &ANSIParser{
		// Match ANSI escape sequences: ESC[...m
		// Also handle cases where ESC is represented as \x1b or \033
		escapeRegex: regexp.MustCompile(`\x1b\[([0-9;]*)m`),
	}
}

// Parse parses the input string and returns a slice of ANSI tokens
func (p *ANSIParser) Parse(input string) ([]ANSIToken, error) {
	var tokens []ANSIToken
	lastEnd := 0

	// Find all escape sequences in the input
	matches := p.escapeRegex.FindAllStringSubmatchIndex(input, -1)

	for _, match := range matches {
		// Add text before escape sequence if any
		if match[0] > lastEnd {
			tokens = append(tokens, ANSIToken{
				Type:    TokenText,
				Content: input[lastEnd:match[0]],
			})
		}

		// Parse escape codes
		codesStr := ""
		if match[2] >= 0 && match[3] >= 0 {
			codesStr = input[match[2]:match[3]]
		}
		codes := p.parseCodes(codesStr)

		tokens = append(tokens, ANSIToken{
			Type:  TokenEscapeSequence,
			Codes: codes,
		})

		lastEnd = match[1]
	}

	// Add remaining text after last escape sequence
	if lastEnd < len(input) {
		tokens = append(tokens, ANSIToken{
			Type:    TokenText,
			Content: input[lastEnd:],
		})
	}

	return tokens, nil
}

// parseCodes parses the numeric codes from an ANSI escape sequence
func (p *ANSIParser) parseCodes(codesStr string) []int {
	if codesStr == "" {
		// Empty code string means reset (equivalent to 0)
		return []int{0}
	}

	parts := strings.Split(codesStr, ";")
	codes := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			// Empty part means 0
			codes = append(codes, 0)
		} else if code, err := strconv.Atoi(part); err == nil {
			codes = append(codes, code)
		}
	}

	// If no valid codes were parsed, return reset
	if len(codes) == 0 {
		return []int{0}
	}

	return codes
}

// ParseSimple is a helper method that parses input and returns just the text content
// with escape sequences removed. Useful for extracting plain text.
func (p *ANSIParser) ParseSimple(input string) string {
	tokens, err := p.Parse(input)
	if err != nil {
		return input
	}

	var result strings.Builder
	for _, token := range tokens {
		if token.Type == TokenText {
			result.WriteString(token.Content)
		}
	}

	return result.String()
}
