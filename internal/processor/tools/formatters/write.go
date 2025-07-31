package formatters

import (
	"fmt"
	"html/template"
)

// WriteFormatter formats the Write tool
type WriteFormatter struct {
	BaseFormatter
}

// NewWriteFormatter creates a new Write formatter
func NewWriteFormatter() *WriteFormatter {
	return &WriteFormatter{
		BaseFormatter: BaseFormatter{toolName: "Write"},
	}
}

// FormatInput formats the input for the Write tool
func (f *WriteFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	// For Write tool, show the content being written
	content := f.extractString(data, "content")
	filePath := f.extractString(data, "file_path")

	// Build the display
	html := fmt.Sprintf(`<div class="write-content">`)
	html += fmt.Sprintf(`<div class="write-header">Writing to: %s</div>`, f.formatPath(filePath))
	html += fmt.Sprintf(`<div class="write-body">%s</div>`, f.formatCode(content))
	html += `</div>`

	return template.HTML(html), nil
}

// ValidateInput validates the input for the Write tool
func (f *WriteFormatter) ValidateInput(data map[string]interface{}) error {
	required := []string{"file_path", "content"}
	for _, field := range required {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// GetDescription returns a custom description for the Write tool
func (f *WriteFormatter) GetDescription(data map[string]interface{}) string {
	return f.extractString(data, "file_path")
}
