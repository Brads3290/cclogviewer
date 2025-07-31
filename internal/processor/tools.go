package processor

import (
	"html/template"

	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/processor/tools"
	"github.com/brads3290/cclogviewer/internal/processor/tools/formatters"
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

// ProcessToolUse processes a tool invocation and formats its display.
func ProcessToolUse(toolUse map[string]interface{}) models.ToolCall {
	return GetToolProcessor().ProcessToolUseWithRegistry(toolUse)
}

// formatBashToolInput is kept for backward compatibility with templates
func formatBashToolInput(input map[string]interface{}, cwd string) template.HTML {
	html, err := registry.FormatWithCWD(ToolNameBash, input, cwd)
	if err != nil {
		return template.HTML("")
	}
	return html
}
