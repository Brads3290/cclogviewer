package formatters

import (
	"fmt"
	"html/template"

	"github.com/brads3290/cclogviewer/internal/processor/tools"
)

// BashFormatter formats the Bash tool
type BashFormatter struct {
	BaseFormatter
}

// Ensure BashFormatter implements tools.BashFormatter interface
var _ tools.BashFormatter = (*BashFormatter)(nil)

// NewBashFormatter creates a new Bash formatter
func NewBashFormatter() *BashFormatter {
	return &BashFormatter{
		BaseFormatter: BaseFormatter{toolName: "Bash"},
	}
}

// FormatInput formats the input for the Bash tool
func (f *BashFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	// For Bash, we return empty HTML since we'll handle formatting in FormatInputWithCWD
	return template.HTML(""), nil
}

// FormatInputWithCWD formats the input for the Bash tool with current working directory
func (f *BashFormatter) FormatInputWithCWD(data map[string]interface{}, cwd string) (template.HTML, error) {
	command := f.extractString(data, "command")
	description := f.extractString(data, "description")
	timeout := f.extractFloat(data, "timeout")

	// Build the bash display HTML
	var html string
	html += `<div class="bash-display">`

	// Header with terminal icon and description
	html += `<div class="bash-header">`
	html += `<span class="terminal-icon">💻</span>`
	html += `<span class="command-label">Bash</span>`
	if description != "" {
		html += fmt.Sprintf(`<span class="description">%s</span>`, f.escapeHTML(description))
	}
	html += `</div>`

	// Terminal display
	html += `<div class="bash-terminal">`

	// Show timeout if specified
	if timeout > 0 {
		html += fmt.Sprintf(`<span class="bash-timeout">timeout: %dms</span>`, int(timeout))
	}

	// Current working directory
	if cwd != "" {
		html += fmt.Sprintf(`<div class="bash-cwd">%s</div>`, f.escapeHTML(cwd))
	}

	// Command line with prompt
	html += `<div class="bash-command-line">`
	html += `<span class="bash-prompt">$</span>`
	html += fmt.Sprintf(`<span class="bash-command">%s</span>`, f.escapeHTML(command))
	html += `</div>`

	// Note: The result will be added by the template when rendering
	html += `</div>`
	html += `</div>`

	return template.HTML(html), nil
}

// ValidateInput validates the input for the Bash tool
func (f *BashFormatter) ValidateInput(data map[string]interface{}) error {
	if f.extractString(data, "command") == "" {
		return fmt.Errorf("missing required field: command")
	}
	return nil
}

// GetDescription returns a custom description for the Bash tool
func (f *BashFormatter) GetDescription(data map[string]interface{}) string {
	// Clear description for Bash since we show it in the custom display
	return ""
}
