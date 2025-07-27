package processor

import (
	"cclogviewer/internal/models"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
)

// ProcessToolUse processes a tool use message and returns a ToolCall
func ProcessToolUse(toolUse map[string]interface{}) models.ToolCall {
	tool := models.ToolCall{
		ID:   GetStringValue(toolUse, "id"),
		Name: GetStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.Description = GetStringValue(input, "description")
		
		// Format the input as JSON
		inputJSON, _ := json.MarshalIndent(input, "", "  ")
		tool.Input = template.HTML(fmt.Sprintf(`<pre class="tool-input">%s</pre>`, html.EscapeString(string(inputJSON))))
	}

	return tool
}