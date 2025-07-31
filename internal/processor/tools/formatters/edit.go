package formatters

import (
	"fmt"
	"html/template"

	"github.com/brads3290/cclogviewer/internal/processor/tools/diff"
)

// EditFormatter formats the Edit tool
type EditFormatter struct {
	BaseFormatter
}

// NewEditFormatter creates a new Edit formatter
func NewEditFormatter() *EditFormatter {
	return &EditFormatter{
		BaseFormatter: BaseFormatter{toolName: "Edit"},
	}
}

// FormatInput formats the input for the Edit tool
func (f *EditFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	oldString := f.extractString(data, "old_string")
	newString := f.extractString(data, "new_string")

	// Compute the diff
	diffLines := diff.ComputeLineDiff(oldString, newString)

	// Format as HTML
	return diff.FormatDiffHTML(diffLines), nil
}

// ValidateInput validates the input for the Edit tool
func (f *EditFormatter) ValidateInput(data map[string]interface{}) error {
	required := []string{"file_path", "old_string", "new_string"}
	for _, field := range required {
		if f.extractString(data, field) == "" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// GetDescription returns a custom description for the Edit tool
func (f *EditFormatter) GetDescription(data map[string]interface{}) string {
	filePath := f.extractString(data, "file_path")
	replaceAll := f.extractBool(data, "replace_all")

	desc := filePath
	if replaceAll {
		desc += " (replace all)"
	}

	return desc
}
