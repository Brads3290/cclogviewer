package processor

import (
	"github.com/brads3290/cclogviewer/internal/models"
	"github.com/brads3290/cclogviewer/internal/utils"
	"strings"
)

// ProcessUserMessage extracts content from user messages.
func ProcessUserMessage(msg map[string]interface{}) string {
	content := utils.ExtractString(msg, "content")

	// Check if content is an array
	if contentArray, ok := msg["content"].([]interface{}); ok && len(contentArray) > 0 {
		if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
			contentType := utils.ExtractString(contentItem, "type")

			switch contentType {
			case ContentTypeText:
				// Handle text content (including interrupted messages)
				text := utils.ExtractString(contentItem, "text")

				return text
			case ContentTypeToolResult:
				// Handle tool result content
				var toolContent string
				if contentVal, ok := contentItem["content"].(string); ok {
					toolContent = contentVal
				} else if contentArray, ok := contentItem["content"].([]interface{}); ok && len(contentArray) > 0 {
					// Handle array content (like from Task tool)
					if textContent, ok := contentArray[0].(map[string]interface{}); ok {
						toolContent = utils.ExtractString(textContent, "text")
					}
				}
				return toolContent
			}
		}
	}

	// Also check direct string content

	return content
}

// ProcessAssistantMessage extracts content and tool calls from assistant messages.
func ProcessAssistantMessage(msg map[string]interface{}, cwd string) (string, []models.ToolCall) {
	var content strings.Builder
	var toolCalls []models.ToolCall

	if contentArray, ok := msg["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if contentItem, ok := item.(map[string]interface{}); ok {
				contentType := utils.ExtractString(contentItem, "type")

				switch contentType {
				case ContentTypeText:
					text := utils.ExtractString(contentItem, "text")
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
