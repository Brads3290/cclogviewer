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
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return err
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

	return tmpl.Execute(file, data)
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
	// [0m = reset all
	
	// Simple approach: handle the most common cases
	ansiPattern := regexp.MustCompile(`\[(\d+)m`)
	
	var result strings.Builder
	lastIndex := 0
	openTags := []string{}
	
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
		case "0":
			// Reset all - close all open tags
			for i := len(openTags) - 1; i >= 0; i-- {
				result.WriteString(openTags[i])
			}
			openTags = nil
		}
		
		lastIndex = match[1]
	}
	
	// Add remaining text
	result.WriteString(text[lastIndex:])
	
	// Close any remaining open tags
	for i := len(openTags) - 1; i >= 0; i-- {
		result.WriteString(openTags[i])
	}
	
	return result.String()
}
