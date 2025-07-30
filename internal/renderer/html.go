package renderer

import (
	"fmt"
	"github.com/Brads3290/cclogviewer/internal/models"
	"html"
	"html/template"
	"os"
	"strings"
	"regexp"
)

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
			// Escape HTML
			content = html.EscapeString(content)
			
			// Check if content is enclosed in square brackets
			trimmed := strings.TrimSpace(content)
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				// Check if it's an ANSI escape sequence
				if !regexp.MustCompile(`\[\d+m`).MatchString(trimmed) {
					// Regular bracketed message (like [Request interrupted by user])
					stripped := trimmed[1:len(trimmed)-1]
					content = fmt.Sprintf(`<span style="color: #999; font-style: italic;">%s</span>`, stripped)
				}
			}
			
			// Convert ANSI escape codes to HTML
			content = convertANSIToHTML(content)
			
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
						result.WriteString(convertANSIToHTML(html.EscapeString(line)))
					}
					
					// Hidden lines
					result.WriteString(`<div class="bash-more-content" style="display: none;">`)
					for i := 20; i < lineCount; i++ {
						result.WriteString("<br>")
						result.WriteString(convertANSIToHTML(html.EscapeString(lines[i])))
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
					// Convert ANSI codes after escaping HTML
					output := html.EscapeString(tc.Result.Content)
					output = convertANSIToHTML(output)
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

// convertANSIToHTML converts ANSI escape codes to HTML formatting
func convertANSIToHTML(text string) string {
	// Common ANSI codes:
	// [1m = bold
	// [22m = normal (not bold)
	// [3m = italic
	// [23m = not italic
	// [4m = underline
	// [24m = not underline
	// [30-37m = foreground colors
	// [40-47m = background colors
	// [90-97m = bright foreground colors
	// [0m = reset all
	// [39m = default foreground color
	// [49m = default background color
	
	// Simple approach: handle the most common cases
	ansiPattern := regexp.MustCompile(`\[(\d+)m`)
	
	var result strings.Builder
	lastIndex := 0
	openTags := []string{}
	openSpans := []string{} // Track color spans separately
	
	for _, match := range ansiPattern.FindAllStringSubmatchIndex(text, -1) {
		// Add text before the match
		result.WriteString(text[lastIndex:match[0]])
		
		// Get the ANSI code
		code := text[match[2]:match[3]]
		
		switch code {
		case "1":
			result.WriteString(`<strong>`)
			openTags = append(openTags, "</strong>")
		case "22":
			// Close bold if open
			for i := len(openTags) - 1; i >= 0; i-- {
				if openTags[i] == "</strong>" {
					result.WriteString("</strong>")
					openTags = append(openTags[:i], openTags[i+1:]...)
					break
				}
			}
		case "3":
			result.WriteString(`<em>`)
			openTags = append(openTags, "</em>")
		case "23":
			// Close italic if open
			for i := len(openTags) - 1; i >= 0; i-- {
				if openTags[i] == "</em>" {
					result.WriteString("</em>")
					openTags = append(openTags[:i], openTags[i+1:]...)
					break
				}
			}
		case "4":
			result.WriteString(`<u>`)
			openTags = append(openTags, "</u>")
		case "24":
			// Close underline if open
			for i := len(openTags) - 1; i >= 0; i-- {
				if openTags[i] == "</u>" {
					result.WriteString("</u>")
					openTags = append(openTags[:i], openTags[i+1:]...)
					break
				}
			}
		// Foreground colors
		case "30":
			result.WriteString(`<span style="color: #000000">`)
			openSpans = append(openSpans, "</span>")
		case "31":
			result.WriteString(`<span style="color: #cc0000">`)
			openSpans = append(openSpans, "</span>")
		case "32":
			result.WriteString(`<span style="color: #4e9a06">`)
			openSpans = append(openSpans, "</span>")
		case "33":
			result.WriteString(`<span style="color: #c4a000">`)
			openSpans = append(openSpans, "</span>")
		case "34":
			result.WriteString(`<span style="color: #3465a4">`)
			openSpans = append(openSpans, "</span>")
		case "35":
			result.WriteString(`<span style="color: #75507b">`)
			openSpans = append(openSpans, "</span>")
		case "36":
			result.WriteString(`<span style="color: #06989a">`)
			openSpans = append(openSpans, "</span>")
		case "37":
			result.WriteString(`<span style="color: #d3d7cf">`)
			openSpans = append(openSpans, "</span>")
		// Bright foreground colors
		case "90":
			result.WriteString(`<span style="color: #555753">`)
			openSpans = append(openSpans, "</span>")
		case "91":
			result.WriteString(`<span style="color: #ef2929">`)
			openSpans = append(openSpans, "</span>")
		case "92":
			result.WriteString(`<span style="color: #8ae234">`)
			openSpans = append(openSpans, "</span>")
		case "93":
			result.WriteString(`<span style="color: #fce94f">`)
			openSpans = append(openSpans, "</span>")
		case "94":
			result.WriteString(`<span style="color: #729fcf">`)
			openSpans = append(openSpans, "</span>")
		case "95":
			result.WriteString(`<span style="color: #ad7fa8">`)
			openSpans = append(openSpans, "</span>")
		case "96":
			result.WriteString(`<span style="color: #34e2e2">`)
			openSpans = append(openSpans, "</span>")
		case "97":
			result.WriteString(`<span style="color: #eeeeec">`)
			openSpans = append(openSpans, "</span>")
		case "39":
			// Default foreground color - close any open color spans
			for i := len(openSpans) - 1; i >= 0; i-- {
				result.WriteString(openSpans[i])
			}
			openSpans = nil
		case "0":
			// Reset all - close all open tags and spans
			for i := len(openTags) - 1; i >= 0; i-- {
				result.WriteString(openTags[i])
			}
			for i := len(openSpans) - 1; i >= 0; i-- {
				result.WriteString(openSpans[i])
			}
			openTags = nil
			openSpans = nil
		}
		
		lastIndex = match[1]
	}
	
	// Add remaining text
	result.WriteString(text[lastIndex:])
	
	// Close any remaining open tags and spans
	for i := len(openTags) - 1; i >= 0; i-- {
		result.WriteString(openTags[i])
	}
	for i := len(openSpans) - 1; i >= 0; i-- {
		result.WriteString(openSpans[i])
	}
	
	return result.String()
}