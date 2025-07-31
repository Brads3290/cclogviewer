package renderer

import (
	"fmt"
	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/renderer/ansi"
	"github.com/brads3290/cclogviewer/internal/renderer/builders"
	"html"
	"html/template"
	"os"
	"regexp"
	"strings"
)

var ansiConverter = ansi.NewANSIConverter()

// GenerateHTML renders processed entries to an HTML file.
func GenerateHTML(entries []*models.ProcessedEntry, outputFile string, debugMode bool) error {
	// Create custom function map
	funcMap := template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
		"mod": func(a, b int) int {
			return a % b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"formatNumber": func(n int) string {
			if n < constants.NumberFormattingThreshold {
				return fmt.Sprintf("%d", n)
			}
			// Format with thousands separators
			str := fmt.Sprintf("%d", n)
			result := ""
			for i, digit := range str {
				if i > 0 && (len(str)-i)%constants.ThousandsSeparatorInterval == 0 {
					result += ","
				}
				result += string(digit)
			}
			return result
		},
		"formatContent": func(content string) template.HTML {
			// Check if content is enclosed in square brackets
			trimmed := strings.TrimSpace(content)
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				// Check if it's an ANSI escape sequence
				if !regexp.MustCompile(`\x1b\[\d+m`).MatchString(content) {
					// Regular bracketed message (like [Request interrupted by user])
					stripped := trimmed[1 : len(trimmed)-1]
					// Escape the content and wrap in styled span
					return template.HTML(fmt.Sprintf(`<span style="color: #999; font-style: italic;">%s</span>`, html.EscapeString(stripped)))
				}
			}

			// Convert ANSI escape codes to HTML (this handles escaping internally)
			content = ConvertANSIToHTML(content)

			// Convert newlines to <br>
			content = strings.ReplaceAll(content, "\n", "<br>")

			return template.HTML(content)
		},
		"shortUUID": func(uuid string) string {
			// Return first N characters of UUID for brevity
			if len(uuid) >= constants.ShortUUIDLength {
				return uuid[:constants.ShortUUIDLength]
			}
			return uuid
		},
		"formatReadResult": func(content string) template.HTML {
			// Format Read tool results with line numbers
			lines := strings.Split(content, "\n")
			var result strings.Builder

			result.WriteString(`<div class="read-content">`)
			result.WriteString(`<div class="read-code">`)

			for _, line := range lines {
				// Extract line number from the format: "   123→content"
				lineNum := ""
				lineContent := line

				if idx := strings.Index(line, "→"); idx > 0 {
					lineNum = strings.TrimSpace(line[:idx])
					// Get the content after the arrow, handling UTF-8 properly
					runes := []rune(line)
					arrowIdx := strings.Index(line, "→")
					if arrowIdx >= 0 && arrowIdx+len("→") < len(line) {
						lineContent = string(runes[len([]rune(line[:arrowIdx]))+1:])
					} else {
						lineContent = ""
					}
				}

				result.WriteString(`<div class="read-line">`)
				if lineNum != "" {
					result.WriteString(fmt.Sprintf(`<span class="line-number">%s</span>`, html.EscapeString(lineNum)))
				}
				// Use separate span for content to enable proper wrapping
				result.WriteString(`<span class="line-content">`)
				escapedContent := html.EscapeString(lineContent)
				result.WriteString(escapedContent)
				result.WriteString(`</span>`)
				result.WriteString(`</div>`)
			}

			result.WriteString(`</div>`)
			result.WriteString(`</div>`)

			return template.HTML(result.String())
		},
		"formatBashResult": func(toolCall interface{}) template.HTML {
			formatter := NewBashResultFormatter()
			return formatter.Format(toolCall)
		},
	}

	// Load templates from embedded filesystem
	tmpl, err := LoadTemplates(funcMap)
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create template data with entries and debug flag
	data := struct {
		Entries []*models.ProcessedEntry
		Debug   bool
	}{
		Entries: entries,
		Debug:   debugMode,
	}

	return ExecuteTemplate(tmpl, file, data)
}

// ConvertANSIToHTML converts ANSI escape sequences to styled HTML.
func ConvertANSIToHTML(input string) string {
	html, err := ansiConverter.ConvertToHTML(input)
	if err != nil {
		// Fallback to escaped text
		return builders.EscapeHTML(input)
	}
	return html
}
