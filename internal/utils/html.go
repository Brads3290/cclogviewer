package utils

import (
	"html"
	"html/template"
)

// EscapeHTML escapes special HTML characters
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// SafeHTML creates a template.HTML from an escaped string
func SafeHTML(s string) template.HTML {
	return template.HTML(EscapeHTML(s))
}

// RawHTML creates a template.HTML without escaping (use with caution)
func RawHTML(s string) template.HTML {
	return template.HTML(s)
}