package processor

import (
	"cclogviewer/internal/models"
	"fmt"
	"html"
	"html/template"
	"strings"
)

// ProcessUserMessage processes user messages and returns formatted HTML content
func ProcessUserMessage(msg map[string]interface{}) template.HTML {
	content := GetStringValue(msg, "content")
	
	// Check if it's a tool result
	if contentArray, ok := msg["content"].([]interface{}); ok && len(contentArray) > 0 {
		if toolResult, ok := contentArray[0].(map[string]interface{}); ok {
			if toolType := GetStringValue(toolResult, "type"); toolType == "tool_result" {
				// Handle different content types
				var toolContent string
				if contentVal, ok := toolResult["content"].(string); ok {
					toolContent = contentVal
				} else if contentArray, ok := toolResult["content"].([]interface{}); ok && len(contentArray) > 0 {
					// Handle array content (like from Task tool)
					if textContent, ok := contentArray[0].(map[string]interface{}); ok {
						toolContent = GetStringValue(textContent, "text")
					}
				}
				
				isError := GetBoolValue(toolResult, "is_error")
				
				if isError {
					return template.HTML(fmt.Sprintf(`<div class="tool-result error">%s</div>`, html.EscapeString(toolContent)))
				}
				return template.HTML(fmt.Sprintf(`<div class="tool-result">%s</div>`, formatContent(toolContent)))
			}
		}
	}
	
	return template.HTML(formatContent(content))
}

// ProcessAssistantMessage processes assistant messages and returns formatted HTML content and tool calls
func ProcessAssistantMessage(msg map[string]interface{}) (template.HTML, []models.ToolCall) {
	var content strings.Builder
	var toolCalls []models.ToolCall

	if contentArray, ok := msg["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if contentItem, ok := item.(map[string]interface{}); ok {
				contentType := GetStringValue(contentItem, "type")
				
				switch contentType {
				case "text":
					text := GetStringValue(contentItem, "text")
					if text != "" {
						content.WriteString(formatContent(text))
					}
				case "tool_use":
					tool := ProcessToolUse(contentItem)
					toolCalls = append(toolCalls, tool)
				}
			}
		}
	}

	return template.HTML(content.String()), toolCalls
}

func formatContent(content string) string {
	// Escape HTML
	content = html.EscapeString(content)
	
	// Convert newlines to <br>
	content = strings.ReplaceAll(content, "\n", "<br>")
	
	// Wrap code blocks
	content = strings.ReplaceAll(content, "```", "</code></pre>CODE_BLOCK_MARKER<pre><code>")
	content = strings.ReplaceAll(content, "CODE_BLOCK_MARKER", "```")
	
	// Remove any empty pre/code tags at start/end
	content = strings.TrimPrefix(content, "</code></pre>```")
	content = strings.TrimSuffix(content, "```<pre><code>")
	
	return content
}