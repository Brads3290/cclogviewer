package formatters

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/brads3290/cclogviewer/internal/processor/tools/diff"
	"github.com/brads3290/cclogviewer/internal/utils"
)

// MultiEditFormatter formats MultiEdit tool inputs and outputs.
type MultiEditFormatter struct {
	BaseFormatter
}

// NewMultiEditFormatter creates a new MultiEdit formatter
func NewMultiEditFormatter() *MultiEditFormatter {
	return &MultiEditFormatter{
		BaseFormatter: BaseFormatter{toolName: "MultiEdit"},
	}
}

// FormatInput formats the input for the MultiEdit tool
func (f *MultiEditFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	edits := f.extractSlice(data, "edits")
	if edits == nil || len(edits) == 0 {
		return template.HTML("<div>No edits specified</div>"), nil
	}

	var result strings.Builder

	// Process each edit
	for i, editInterface := range edits {
		edit, ok := editInterface.(map[string]interface{})
		if !ok {
			continue
		}

		oldString := f.extractString(edit, "old_string")
		newString := f.extractString(edit, "new_string")
		replaceAll := f.extractBool(edit, "replace_all")

		// Compute the diff for this edit
		diffLines := diff.ComputeLineDiff(oldString, newString)

		// Add separator between edits
		if i > 0 {
			result.WriteString(`<div style="border-top: 1px solid #dee2e6; margin: 10px 0;"></div>`)
		}

		// Edit header
		result.WriteString(fmt.Sprintf(`<div style="color: #6c757d; font-size: 0.85em; margin-bottom: 5px;">Edit #%d`, i+1))
		if replaceAll {
			result.WriteString(` <span style="background: #6c757d; color: white; padding: 2px 6px; border-radius: 3px; font-size: 0.9em;">(Replace All)</span>`)
		}
		result.WriteString(`</div>`)

		// Add the diff
		result.WriteString(string(diff.FormatDiffHTML(diffLines)))
	}

	return template.HTML(result.String()), nil
}

// ValidateInput validates the input for the MultiEdit tool
func (f *MultiEditFormatter) ValidateInput(data map[string]interface{}) error {
	if err := utils.ValidateRequiredField(data, "file_path"); err != nil {
		return err
	}

	edits := f.extractSlice(data, "edits")
	if edits == nil || len(edits) == 0 {
		return fmt.Errorf("missing or empty edits array")
	}

	return nil
}

// GetDescription returns a custom description for the MultiEdit tool
func (f *MultiEditFormatter) GetDescription(data map[string]interface{}) string {
	filePath := f.extractString(data, "file_path")
	edits := f.extractSlice(data, "edits")

	if len(edits) > 0 {
		return fmt.Sprintf("%s (%d edits)", filePath, len(edits))
	}
	return filePath
}
