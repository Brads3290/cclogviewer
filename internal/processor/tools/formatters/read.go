package formatters

import (
	"fmt"
	"html/template"

	"github.com/brads3290/cclogviewer/internal/constants"
	"github.com/brads3290/cclogviewer/internal/utils"
)

// ReadFormatter formats Read tool inputs and outputs.
type ReadFormatter struct {
	BaseFormatter
}

// NewReadFormatter creates a new Read formatter
func NewReadFormatter() *ReadFormatter {
	return &ReadFormatter{
		BaseFormatter: BaseFormatter{toolName: constants.ToolNameRead},
	}
}

// FormatInput formats the input for the Read tool
func (f *ReadFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	// For Read tool, we return empty HTML since the file path is shown in the description
	// and the content is shown in the result
	return template.HTML(""), nil
}

// ValidateInput validates the input for the Read tool
func (f *ReadFormatter) ValidateInput(data map[string]interface{}) error {
	return utils.ValidateRequiredField(data, "file_path")
}

// GetDescription returns a custom description for the Read tool
func (f *ReadFormatter) GetDescription(data map[string]interface{}) string {
	filePath := f.extractString(data, "file_path")
	offset := f.extractInt(data, "offset")
	limit := f.extractInt(data, "limit")

	desc := filePath

	// Add line info if offset/limit specified
	if offset > 0 || limit > 0 {
		if offset > 0 && limit > 0 {
			desc += fmt.Sprintf(" (lines %d-%d)", offset, offset+limit-1)
		} else if offset > 0 {
			desc += fmt.Sprintf(" (starting at line %d)", offset)
		} else if limit > 0 {
			desc += fmt.Sprintf(" (first %d lines)", limit)
		}
	}

	return desc
}
