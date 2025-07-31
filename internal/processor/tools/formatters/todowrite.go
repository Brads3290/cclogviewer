package formatters

import (
	"fmt"
	"html/template"
)

// TodoWriteFormatter formats the TodoWrite tool
type TodoWriteFormatter struct {
	BaseFormatter
}

// NewTodoWriteFormatter creates a new TodoWrite formatter
func NewTodoWriteFormatter() *TodoWriteFormatter {
	return &TodoWriteFormatter{
		BaseFormatter: BaseFormatter{toolName: "TodoWrite"},
	}
}

// FormatInput formats the input for the TodoWrite tool
func (f *TodoWriteFormatter) FormatInput(data map[string]interface{}) (template.HTML, error) {
	// For TodoWrite, we use the compact view
	return f.GetCompactView(data), nil
}

// ValidateInput validates the input for the TodoWrite tool
func (f *TodoWriteFormatter) ValidateInput(data map[string]interface{}) error {
	todos := f.extractSlice(data, "todos")
	if todos == nil {
		return fmt.Errorf("missing required field: todos")
	}
	return nil
}

// GetDescription returns a custom description for the TodoWrite tool
func (f *TodoWriteFormatter) GetDescription(data map[string]interface{}) string {
	return f.extractString(data, "description")
}

// GetCompactView returns a compact view of the todo list
func (f *TodoWriteFormatter) GetCompactView(data map[string]interface{}) template.HTML {
	todos := f.extractSlice(data, "todos")
	if todos == nil || len(todos) == 0 {
		return template.HTML("")
	}

	// Build compact todo display
	var html string
	html += `<div class="todo-compact">`

	// Count tasks by status
	pending, inProgress, completed := 0, 0, 0
	for _, todoInterface := range todos {
		todo, ok := todoInterface.(map[string]interface{})
		if !ok {
			continue
		}
		status := f.extractString(todo, "status")
		switch status {
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		case "completed":
			completed++
		}
	}

	// Add summary bar
	total := pending + inProgress + completed
	if total > 0 {
		html += `<div class="todo-compact-summary">`
		html += `<span class="todo-compact-title">📋 Todo List</span>`

		if completed > 0 {
			html += fmt.Sprintf(`<span class="todo-stat completed">✓ %d</span>`, completed)
		}
		if inProgress > 0 {
			html += fmt.Sprintf(`<span class="todo-stat in-progress">⏳ %d</span>`, inProgress)
		}
		if pending > 0 {
			html += fmt.Sprintf(`<span class="todo-stat pending">○ %d</span>`, pending)
		}
		html += `</div>`

		// Show todo items
		html += `<div class="todo-compact-items">`
		for _, todoInterface := range todos {
			todo, ok := todoInterface.(map[string]interface{})
			if !ok {
				continue
			}

			content := f.extractString(todo, "content")
			status := f.extractString(todo, "status")
			priority := f.extractString(todo, "priority")

			// Determine status icon
			var statusIcon string
			var statusClass string
			switch status {
			case "completed":
				statusIcon = "✓"
				statusClass = "completed"
			case "in_progress":
				statusIcon = "⏳"
				statusClass = "in-progress"
			case "pending":
				statusIcon = "○"
				statusClass = "pending"
			}

			// Priority badge
			var priorityBadge string
			if priority == "high" {
				priorityBadge = ` <span class="todo-priority-badge high">H</span>`
			} else if priority == "medium" {
				priorityBadge = ` <span class="todo-priority-badge medium">M</span>`
			}

			html += fmt.Sprintf(`<div class="todo-compact-item %s"><span class="todo-icon">%s</span> %s%s</div>`,
				statusClass, statusIcon, f.escapeHTML(content), priorityBadge)
		}
		html += `</div>`
	}

	html += `</div>`
	return template.HTML(html)
}
