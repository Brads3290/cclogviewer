package processor

import (
	"html/template"

	"github.com/Brads3290/cclogviewer/internal/models"
	"github.com/Brads3290/cclogviewer/internal/processor/tools"
	"github.com/Brads3290/cclogviewer/internal/processor/tools/formatters"
)

var registry *tools.FormatterRegistry

func init() {
	registry = tools.NewFormatterRegistry()

	// Register all formatters
	registry.Register(formatters.NewEditFormatter())
	registry.Register(formatters.NewMultiEditFormatter())
	registry.Register(formatters.NewWriteFormatter())
	registry.Register(formatters.NewReadFormatter())
	registry.Register(formatters.NewBashFormatter())
	registry.Register(formatters.NewTodoWriteFormatter())
}

// ProcessToolUse processes a tool use message and returns a ToolCall
func ProcessToolUse(toolUse map[string]interface{}) models.ToolCall {
	tool := models.ToolCall{
		ID:   GetStringValue(toolUse, "id"),
		Name: GetStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.RawInput = input // Store raw input for later use

		// Get description from formatter
		tool.Description = registry.GetDescription(tool.Name, input)

		// Format the input
		formattedInput, err := registry.Format(tool.Name, input)
		if err != nil {
			// Fallback to empty on error
			tool.Input = template.HTML("")
		} else {
			tool.Input = formattedInput
		}

		// Generate compact view
		tool.CompactView = registry.GetCompactView(tool.Name, input)
	}

	return tool
}

// formatBashToolInput is kept for backward compatibility with templates
func formatBashToolInput(input map[string]interface{}, cwd string) template.HTML {
	html, err := registry.FormatWithCWD("Bash", input, cwd)
	if err != nil {
		return template.HTML("")
	}
	return html
}
