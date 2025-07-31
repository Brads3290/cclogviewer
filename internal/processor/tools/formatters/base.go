package formatters

import (
	"fmt"
	"html"
	"html/template"
)

// BaseFormatter provides common functionality for tool formatters
type BaseFormatter struct {
	toolName string
}

// Name returns the tool name
func (b *BaseFormatter) Name() string {
	return b.toolName
}

// FormatOutput provides a default implementation for output formatting
func (b *BaseFormatter) FormatOutput(output interface{}) (template.HTML, error) {
	// Default implementation - most tools don't need special output formatting
	return template.HTML(""), nil
}

// GetCompactView provides a default implementation returning empty
func (b *BaseFormatter) GetCompactView(data map[string]interface{}) template.HTML {
	return template.HTML("")
}

// extractString safely extracts a string value from a map
func (b *BaseFormatter) extractString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// extractBool safely extracts a boolean value from a map
func (b *BaseFormatter) extractBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

// extractFloat safely extracts a float64 value from a map
func (b *BaseFormatter) extractFloat(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return 0
}

// extractInt safely extracts an int value from a map
func (b *BaseFormatter) extractInt(data map[string]interface{}, key string) int {
	return int(b.extractFloat(data, key))
}

// extractSlice safely extracts a slice value from a map
func (b *BaseFormatter) extractSlice(data map[string]interface{}, key string) []interface{} {
	if val, ok := data[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			return slice
		}
	}
	return nil
}

// escapeHTML escapes HTML special characters
func (b *BaseFormatter) escapeHTML(s string) string {
	return html.EscapeString(s)
}

// formatPath formats a file path with styling
func (b *BaseFormatter) formatPath(path string) string {
	return fmt.Sprintf(`<span class="file-path">%s</span>`, b.escapeHTML(path))
}

// formatCode formats code content with proper escaping
func (b *BaseFormatter) formatCode(code string) string {
	return fmt.Sprintf(`<pre class="code-content">%s</pre>`, b.escapeHTML(code))
}

// formatInlineCode formats inline code
func (b *BaseFormatter) formatInlineCode(code string) string {
	return fmt.Sprintf(`<code>%s</code>`, b.escapeHTML(code))
}
