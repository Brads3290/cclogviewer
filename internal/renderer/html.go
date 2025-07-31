package renderer

import (
	"fmt"
	"github.com/Brads3290/cclogviewer/internal/models"
	"github.com/Brads3290/cclogviewer/internal/renderer/ansi"
	"github.com/Brads3290/cclogviewer/internal/renderer/builders"
	"html"
	"html/template"
	"os"
	"regexp"
	"strings"
)

var ansiConverter = ansi.NewANSIConverter()

// GenerateHTML generates an HTML file from processed entries
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
			if n < 1000 {
				return fmt.Sprintf("%d", n)
			}
			// Format with thousands separators
			str := fmt.Sprintf("%d", n)
			result := ""
			for i, digit := range str {
				if i > 0 && (len(str)-i)%3 == 0 {
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
			// Return first 8 characters of UUID for brevity
			if len(uuid) >= 8 {
				return uuid[:8]
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
			// Special formatting for Bash tool to integrate result into terminal
			tc, ok := toolCall.(models.ToolCall)
			if !ok || tc.Name != "Bash" {
				return ""
			}

			var result strings.Builder

			// Add the bash display with integrated result
			input, _ := tc.RawInput.(map[string]interface{})
			command := ""
			description := ""
			if input != nil {
				command = strings.TrimSpace(fmt.Sprintf("%v", input["command"]))
				description = strings.TrimSpace(fmt.Sprintf("%v", input["description"]))
			}

			result.WriteString(`<div class="bash-display">`)

			// Header
			result.WriteString(`<div class="bash-header">`)
			result.WriteString(`<span class="terminal-icon">💻</span>`)
			result.WriteString(`<span class="command-label">Bash</span>`)
			if description != "" && description != "<nil>" {
				result.WriteString(fmt.Sprintf(`<span class="description">%s</span>`, html.EscapeString(description)))
			}
			result.WriteString(`</div>`)

			// Terminal
			result.WriteString(`<div class="bash-terminal">`)

			// CWD
			if tc.CWD != "" {
				result.WriteString(fmt.Sprintf(`<div class="bash-cwd">%s</div>`, html.EscapeString(tc.CWD)))
			}

			// Command
			result.WriteString(`<div class="bash-command-line">`)
			result.WriteString(`<span class="bash-prompt">$</span>`)
			result.WriteString(fmt.Sprintf(`<span class="bash-command">%s</span>`, html.EscapeString(command)))
			result.WriteString(`</div>`)

			// Output
			if tc.Result != nil && tc.Result.Content != "" {
				lines := strings.Split(tc.Result.Content, "\n")
				lineCount := len(lines)

				if lineCount > 20 {
					// For outputs over 20 lines, add collapsible functionality
					result.WriteString(`<div class="bash-output" style="position: relative;">`)

					// First 20 lines always visible
					visibleLines := lines[:20]
					for i, line := range visibleLines {
						if i > 0 {
							result.WriteString("<br>")
						}
						result.WriteString(ConvertANSIToHTML(line))
					}

					// Hidden lines
					result.WriteString(`<div class="bash-more-content" style="display: none;">`)
					for i := 20; i < lineCount; i++ {
						result.WriteString("<br>")
						result.WriteString(ConvertANSIToHTML(lines[i]))
					}
					result.WriteString(`</div>`)

					// More/Less link
					result.WriteString(`<div style="margin-top: 5px;">`)
					result.WriteString(`<a href="#" class="bash-more-link" style="color: #0066cc; text-decoration: none;" onclick="`)
					result.WriteString(`event.preventDefault(); `)
					result.WriteString(`var content = this.parentElement.previousElementSibling; `)
					result.WriteString(`var isHidden = content.style.display === 'none'; `)
					result.WriteString(`content.style.display = isHidden ? 'block' : 'none'; `)
					result.WriteString(`this.textContent = isHidden ? 'Less' : 'More'; `)
					result.WriteString(`return false;">More</a>`)
					result.WriteString(`</div>`)

					result.WriteString(`</div>`)
				} else {
					// For outputs under 20 lines, show all
					result.WriteString(`<div class="bash-output">`)
					// Convert ANSI codes (ConvertANSIToHTML handles escaping)
					output := ConvertANSIToHTML(tc.Result.Content)
					result.WriteString(strings.ReplaceAll(output, "\n", "<br>"))
					result.WriteString(`</div>`)
				}
			}

			result.WriteString(`</div>`)
			result.WriteString(`</div>`)

			return template.HTML(result.String())
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

// ConvertANSIToHTML converts ANSI escape sequences to HTML
func ConvertANSIToHTML(input string) string {
	html, err := ansiConverter.ConvertToHTML(input)
	if err != nil {
		// Fallback to escaped text
		return builders.EscapeHTML(input)
	}
	return html
}
