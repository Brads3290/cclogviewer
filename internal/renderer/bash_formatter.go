package renderer

import (
	"fmt"
	"github.com/Brads3290/cclogviewer/internal/models"
	"github.com/Brads3290/cclogviewer/internal/processor"
	"html"
	"html/template"
	"strings"
)

// BashResultFormatter handles formatting of bash command results
type BashResultFormatter struct{}

// NewBashResultFormatter creates a new bash result formatter
func NewBashResultFormatter() *BashResultFormatter {
	return &BashResultFormatter{}
}

// Format formats a bash tool call result into HTML
func (f *BashResultFormatter) Format(toolCall interface{}) template.HTML {
	tc, ok := toolCall.(models.ToolCall)
	if !ok || tc.Name != "Bash" {
		return ""
	}

	var result strings.Builder
	input := f.extractInput(tc)

	result.WriteString(`<div class="bash-display">`)
	f.renderHeader(&result, input.description)
	f.renderTerminal(&result, input.command, tc)
	result.WriteString(`</div>`)

	return template.HTML(result.String())
}

// bashInputData holds extracted input data
type bashInputData struct {
	command     string
	description string
}

// extractInput extracts command and description from tool call input
func (f *BashResultFormatter) extractInput(tc models.ToolCall) bashInputData {
	input, _ := tc.RawInput.(map[string]interface{})
	data := bashInputData{}

	if input != nil {
		data.command = strings.TrimSpace(fmt.Sprintf("%v", input["command"]))
		data.description = strings.TrimSpace(fmt.Sprintf("%v", input["description"]))
	}

	return data
}

// renderHeader renders the bash command header
func (f *BashResultFormatter) renderHeader(result *strings.Builder, description string) {
	result.WriteString(`<div class="bash-header">`)
	result.WriteString(`<span class="terminal-icon">💻</span>`)
	result.WriteString(`<span class="command-label">Bash</span>`)

	if description != "" && description != "<nil>" {
		result.WriteString(fmt.Sprintf(`<span class="description">%s</span>`, html.EscapeString(description)))
	}

	result.WriteString(`</div>`)
}

// renderTerminal renders the terminal section with command and output
func (f *BashResultFormatter) renderTerminal(result *strings.Builder, command string, tc models.ToolCall) {
	result.WriteString(`<div class="bash-terminal">`)

	f.renderCWD(result, tc.CWD)
	f.renderCommandLine(result, command)
	f.renderOutput(result, tc)

	result.WriteString(`</div>`)
}

// renderCWD renders the current working directory
func (f *BashResultFormatter) renderCWD(result *strings.Builder, cwd string) {
	if cwd != "" {
		result.WriteString(fmt.Sprintf(`<div class="bash-cwd">%s</div>`, html.EscapeString(cwd)))
	}
}

// renderCommandLine renders the command prompt and command
func (f *BashResultFormatter) renderCommandLine(result *strings.Builder, command string) {
	result.WriteString(`<div class="bash-command-line">`)
	result.WriteString(`<span class="bash-prompt">$</span>`)
	result.WriteString(fmt.Sprintf(`<span class="bash-command">%s</span>`, html.EscapeString(command)))
	result.WriteString(`</div>`)
}

// renderOutput renders the command output with collapsible functionality for long outputs
func (f *BashResultFormatter) renderOutput(result *strings.Builder, tc models.ToolCall) {
	if tc.Result == nil || tc.Result.Content == "" {
		return
	}

	lines, isLong := f.processOutput(tc.Result.Content)

	if isLong {
		f.renderCollapsibleOutput(result, lines)
	} else {
		f.renderSimpleOutput(result, tc.Result.Content)
	}
}

// processOutput processes the output content and determines if it's long
func (f *BashResultFormatter) processOutput(output string) (lines []string, isLong bool) {
	lines = strings.Split(output, "\n")
	isLong = len(lines) > processor.BashOutputCollapseThreshold
	return
}

// renderCollapsibleOutput renders output with collapsible sections for long content
func (f *BashResultFormatter) renderCollapsibleOutput(result *strings.Builder, lines []string) {
	result.WriteString(`<div class="bash-output" style="position: relative;">`)

	// First 20 lines always visible
	visibleLines := lines[:processor.BashOutputCollapseThreshold]
	convertedLines := f.convertLinesToHTML(visibleLines)

	for i, line := range convertedLines {
		if i > 0 {
			result.WriteString("<br>")
		}
		result.WriteString(line)
	}

	// Hidden lines
	result.WriteString(`<div class="bash-more-content" style="display: none;">`)
	hiddenLines := lines[processor.BashOutputCollapseThreshold:]
	convertedHiddenLines := f.convertLinesToHTML(hiddenLines)

	for _, line := range convertedHiddenLines {
		result.WriteString("<br>")
		result.WriteString(line)
	}
	result.WriteString(`</div>`)

	f.renderMoreLessToggle(result)
	result.WriteString(`</div>`)
}

// renderSimpleOutput renders simple output without collapsible functionality
func (f *BashResultFormatter) renderSimpleOutput(result *strings.Builder, content string) {
	result.WriteString(`<div class="bash-output">`)
	output := ConvertANSIToHTML(content)
	result.WriteString(strings.ReplaceAll(output, "\n", "<br>"))
	result.WriteString(`</div>`)
}

// renderMoreLessToggle renders the More/Less toggle link
func (f *BashResultFormatter) renderMoreLessToggle(result *strings.Builder) {
	result.WriteString(`<div style="margin-top: 5px;">`)
	result.WriteString(`<a href="#" class="bash-more-link" style="color: #0066cc; text-decoration: none;" onclick="`)
	result.WriteString(`event.preventDefault(); `)
	result.WriteString(`var content = this.parentElement.previousElementSibling; `)
	result.WriteString(`var isHidden = content.style.display === 'none'; `)
	result.WriteString(`content.style.display = isHidden ? 'block' : 'none'; `)
	result.WriteString(`this.textContent = isHidden ? 'Less' : 'More'; `)
	result.WriteString(`return false;">More</a>`)
	result.WriteString(`</div>`)
}

// convertLinesToHTML converts an array of lines to HTML with ANSI conversion
func (f *BashResultFormatter) convertLinesToHTML(lines []string) []string {
	converted := make([]string, len(lines))
	for i, line := range lines {
		converted[i] = ConvertANSIToHTML(line)
	}
	return converted
}
