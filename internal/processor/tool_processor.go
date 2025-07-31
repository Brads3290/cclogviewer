package processor

import (
	"html/template"

	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/processor/tools"
)

// ToolProcessor formats and processes tool calls.
type ToolProcessor struct {
	registry *tools.FormatterRegistry
}

// globalToolProcessor is the singleton instance
var globalToolProcessor *ToolProcessor

// initGlobalToolProcessor initializes the global tool processor
func initGlobalToolProcessor() {
	if globalToolProcessor == nil {
		globalToolProcessor = &ToolProcessor{
			registry: registry, // Uses the global registry from tools.go
		}
	}
}

// GetToolProcessor returns the singleton tool processor.
func GetToolProcessor() *ToolProcessor {
	initGlobalToolProcessor()
	return globalToolProcessor
}

// ProcessToolCall processes a tool call, applying formatting and extracting metadata
func (tp *ToolProcessor) ProcessToolCall(toolCall *models.ToolCall) {
	if toolCall == nil {
		return
	}

	// Get description from formatter
	if input, ok := toolCall.RawInput.(map[string]interface{}); ok {
		toolCall.Description = tp.registry.GetDescription(toolCall.Name, input)

		// Format the input
		formattedInput, err := tp.registry.Format(toolCall.Name, input)
		if err != nil {
			// Fallback to empty on error
			toolCall.Input = template.HTML("")
		} else {
			toolCall.Input = formattedInput
		}

		// Generate compact view
		toolCall.CompactView = tp.registry.GetCompactView(toolCall.Name, input)
	}
}

// ProcessToolUseWithRegistry processes a tool use message and returns a ToolCall
// This replaces the standalone ProcessToolUse function
func (tp *ToolProcessor) ProcessToolUseWithRegistry(toolUse map[string]interface{}) models.ToolCall {
	tool := models.ToolCall{
		ID:   GetStringValue(toolUse, "id"),
		Name: GetStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.RawInput = input // Store raw input for later use
		tp.ProcessToolCall(&tool)
	}

	return tool
}
