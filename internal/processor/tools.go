package processor

import (
	"cclogviewer/internal/models"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strings"
)

// ProcessToolUse processes a tool use message and returns a ToolCall
func ProcessToolUse(toolUse map[string]interface{}) models.ToolCall {
	tool := models.ToolCall{
		ID:   GetStringValue(toolUse, "id"),
		Name: GetStringValue(toolUse, "name"),
	}

	if input, ok := toolUse["input"].(map[string]interface{}); ok {
		tool.Description = GetStringValue(input, "description")
		
		// Special handling for Edit tool
		if tool.Name == "Edit" {
			tool.Input = formatEditToolInput(input)
		} else {
			// Format the input as JSON
			inputJSON, _ := json.MarshalIndent(input, "", "  ")
			tool.Input = template.HTML(fmt.Sprintf(`<pre class="tool-input">%s</pre>`, html.EscapeString(string(inputJSON))))
		}
	}

	return tool
}

// formatEditToolInput creates a diff view for Edit tool inputs
func formatEditToolInput(input map[string]interface{}) template.HTML {
	filePath := GetStringValue(input, "file_path")
	oldString := GetStringValue(input, "old_string")
	newString := GetStringValue(input, "new_string")
	replaceAll := false
	if val, ok := input["replace_all"].(bool); ok {
		replaceAll = val
	}

	// Split strings into lines for diff display
	oldLines := strings.Split(oldString, "\n")
	newLines := strings.Split(newString, "\n")

	// Build the diff HTML
	var diffHTML strings.Builder
	diffHTML.WriteString(`<div class="edit-diff">`)
	diffHTML.WriteString(fmt.Sprintf(`<div class="diff-header">📝 Edit: <span class="file-path">%s</span>`, html.EscapeString(filePath)))
	if replaceAll {
		diffHTML.WriteString(` <span class="replace-all">(Replace All)</span>`)
	}
	diffHTML.WriteString(`</div>`)
	diffHTML.WriteString(`<div class="diff-content">`)

	// Simple diff display - show removed and added sections
	if len(oldLines) > 0 || oldString != "" {
		diffHTML.WriteString(`<div class="diff-section removed">`)
		diffHTML.WriteString(`<div class="diff-section-header">- Removed</div>`)
		diffHTML.WriteString(`<pre class="diff-code">`)
		for i, line := range oldLines {
			lineNum := i + 1
			diffHTML.WriteString(fmt.Sprintf(`<span class="line-number">%3d</span> %s`, lineNum, html.EscapeString(line)))
			if i < len(oldLines)-1 {
				diffHTML.WriteString("\n")
			}
		}
		diffHTML.WriteString(`</pre>`)
		diffHTML.WriteString(`</div>`)
	}

	if len(newLines) > 0 || newString != "" {
		diffHTML.WriteString(`<div class="diff-section added">`)
		diffHTML.WriteString(`<div class="diff-section-header">+ Added</div>`)
		diffHTML.WriteString(`<pre class="diff-code">`)
		for i, line := range newLines {
			lineNum := i + 1
			diffHTML.WriteString(fmt.Sprintf(`<span class="line-number">%3d</span> %s`, lineNum, html.EscapeString(line)))
			if i < len(newLines)-1 {
				diffHTML.WriteString("\n")
			}
		}
		diffHTML.WriteString(`</pre>`)
		diffHTML.WriteString(`</div>`)
	}

	diffHTML.WriteString(`</div></div>`)
	return template.HTML(diffHTML.String())
}