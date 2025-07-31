package formatters

import (
	"fmt"
	"html/template"

	"github.com/brads3290/cclogviewer/internal/utils"
)

// BaseFormatter provides common functionality for tool formatters.
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

// Helper methods that delegate to utils package
func (b *BaseFormatter) extractString(data map[string]interface{}, key string) string {
	return utils.ExtractString(data, key)
}

func (b *BaseFormatter) extractBool(data map[string]interface{}, key string) bool {
	return utils.ExtractBool(data, key)
}

func (b *BaseFormatter) extractFloat(data map[string]interface{}, key string) float64 {
	return utils.ExtractFloat64(data, key)
}

func (b *BaseFormatter) extractInt(data map[string]interface{}, key string) int {
	return utils.ExtractInt(data, key)
}

func (b *BaseFormatter) extractSlice(data map[string]interface{}, key string) []interface{} {
	return utils.ExtractSlice(data, key)
}

// escapeHTML escapes HTML special characters
func (b *BaseFormatter) escapeHTML(s string) string {
	return utils.EscapeHTML(s)
}

// formatPath formats a file path with styling
func (b *BaseFormatter) formatPath(path string) string {
	return fmt.Sprintf(`<span class="file-path">%s</span>`, utils.EscapeHTML(path))
}

// formatCode formats code content with proper escaping
func (b *BaseFormatter) formatCode(code string) string {
	return fmt.Sprintf(`<pre class="code-content">%s</pre>`, utils.EscapeHTML(code))
}

// formatInlineCode formats inline code
func (b *BaseFormatter) formatInlineCode(code string) string {
	return fmt.Sprintf(`<code>%s</code>`, utils.EscapeHTML(code))
}
