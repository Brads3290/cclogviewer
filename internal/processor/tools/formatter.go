package tools

import (
	"fmt"
	"html"
	"html/template"
	"strings"
	"sync"
)

// ToolFormatter formats tool inputs and outputs for display.
type ToolFormatter interface {
	Name() string
	FormatInput(data map[string]interface{}) (template.HTML, error)
	FormatOutput(output interface{}) (template.HTML, error)
	ValidateInput(data map[string]interface{}) error
	GetDescription(data map[string]interface{}) string
	GetCompactView(data map[string]interface{}) template.HTML
}

// BashFormatter extends ToolFormatter with CWD context.
type BashFormatter interface {
	ToolFormatter
	FormatInputWithCWD(data map[string]interface{}, cwd string) (template.HTML, error)
}

// FormatterRegistry manages tool-specific formatters.
type FormatterRegistry struct {
	formatters map[string]ToolFormatter
	mu         sync.RWMutex
}

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry() *FormatterRegistry {
	return &FormatterRegistry{
		formatters: make(map[string]ToolFormatter),
	}
}

// Register adds a formatter to the registry
func (r *FormatterRegistry) Register(formatter ToolFormatter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formatters[formatter.Name()] = formatter
}

// Format formats tool input using the appropriate formatter
func (r *FormatterRegistry) Format(toolName string, data map[string]interface{}) (template.HTML, error) {
	r.mu.RLock()
	formatter, exists := r.formatters[toolName]
	r.mu.RUnlock()

	if !exists {
		return r.formatGeneric(toolName, data)
	}

	if err := formatter.ValidateInput(data); err != nil {
		return "", fmt.Errorf("invalid input for %s: %w", toolName, err)
	}

	return formatter.FormatInput(data)
}

// GetDescription gets the tool description using the appropriate formatter
func (r *FormatterRegistry) GetDescription(toolName string, data map[string]interface{}) string {
	r.mu.RLock()
	formatter, exists := r.formatters[toolName]
	r.mu.RUnlock()

	if !exists {
		// Default to extracting description field
		if desc, ok := data["description"].(string); ok {
			return desc
		}
		return ""
	}

	return formatter.GetDescription(data)
}

// GetCompactView gets the compact view for a tool
func (r *FormatterRegistry) GetCompactView(toolName string, data map[string]interface{}) template.HTML {
	r.mu.RLock()
	formatter, exists := r.formatters[toolName]
	r.mu.RUnlock()

	if !exists {
		return template.HTML("")
	}

	return formatter.GetCompactView(data)
}

// FormatWithCWD formats tool input with current working directory (for Bash tool)
func (r *FormatterRegistry) FormatWithCWD(toolName string, data map[string]interface{}, cwd string) (template.HTML, error) {
	r.mu.RLock()
	formatter, exists := r.formatters[toolName]
	r.mu.RUnlock()

	if !exists {
		return r.formatGeneric(toolName, data)
	}

	// Check if it's a BashFormatter
	if bashFormatter, ok := formatter.(BashFormatter); ok {
		if err := formatter.ValidateInput(data); err != nil {
			return "", fmt.Errorf("invalid input for %s: %w", toolName, err)
		}
		return bashFormatter.FormatInputWithCWD(data, cwd)
	}

	// Fallback to regular formatting
	return r.Format(toolName, data)
}

// formatGeneric formats tools that don't have specific formatters
func (r *FormatterRegistry) formatGeneric(toolName string, data map[string]interface{}) (template.HTML, error) {
	// Build a JSON-like representation
	var sb strings.Builder
	sb.WriteString(`<div class="tool-input"><pre>`)

	// Format the data as key-value pairs
	first := true
	sb.WriteString("{")
	for key, value := range data {
		if !first {
			sb.WriteString(",")
		}
		first = false
		sb.WriteString("\n  ")
		sb.WriteString(fmt.Sprintf(`"%s": `, key))

		// Format the value based on type
		switch v := value.(type) {
		case string:
			sb.WriteString(fmt.Sprintf(`"%s"`, html.EscapeString(v)))
		case float64:
			sb.WriteString(fmt.Sprintf("%v", v))
		case bool:
			sb.WriteString(fmt.Sprintf("%v", v))
		case []interface{}:
			sb.WriteString("[...]")
		case map[string]interface{}:
			sb.WriteString("{...}")
		default:
			sb.WriteString(fmt.Sprintf("%v", v))
		}
	}
	if len(data) > 0 {
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	sb.WriteString(`</pre></div>`)
	return template.HTML(sb.String()), nil
}
