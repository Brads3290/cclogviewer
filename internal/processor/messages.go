package processor

import (
	"github.com/Brads3290/cclogviewer/internal/models"
	"strings"
)

// ProcessUserMessage processes user messages and returns raw content
func ProcessUserMessage(msg map[string]interface{}) string {
	content := GetStringValue(msg, "content")

	// Check if content is an array
	if contentArray, ok := msg["content"].([]interface{}); ok && len(contentArray) > 0 {
		if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
			contentType := GetStringValue(contentItem, "type")

			switch contentType {
			case ContentTypeText:
				// Handle text content (including interrupted messages)
				text := GetStringValue(contentItem, "text")

				return text
			case ContentTypeToolResult:
				// Handle tool result content
				var toolContent string
				if contentVal, ok := contentItem["content"].(string); ok {
					toolContent = contentVal
				} else if contentArray, ok := contentItem["content"].([]interface{}); ok && len(contentArray) > 0 {
					// Handle array content (like from Task tool)
					if textContent, ok := contentArray[0].(map[string]interface{}); ok {
						toolContent = GetStringValue(textContent, "text")
					}
				}
				return toolContent
			}
		}
	}

	// Also check direct string content

	return content
}

// ProcessAssistantMessage processes assistant messages and returns raw content and tool calls
func ProcessAssistantMessage(msg map[string]interface{}, cwd string) (string, []models.ToolCall) {
	var content strings.Builder
	var toolCalls []models.ToolCall

	if contentArray, ok := msg["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if contentItem, ok := item.(map[string]interface{}); ok {
				contentType := GetStringValue(contentItem, "type")

				switch contentType {
				case ContentTypeText:
					text := GetStringValue(contentItem, "text")
					if text != "" {
						content.WriteString(text)
					}
				case ContentTypeToolUse:
					tool := ProcessToolUse(contentItem)
					tool.CWD = cwd
					toolCalls = append(toolCalls, tool)
				}
			}
		}
	}

	return content.String(), toolCalls
}
