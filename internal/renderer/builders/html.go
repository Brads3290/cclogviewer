package builders

import (
	"fmt"
	"github.com/Brads3290/cclogviewer/internal/processor"
	"html"
	"strings"
)

// HTMLBuilder provides methods for building HTML content
type HTMLBuilder struct {
	parts []string
}

// NewHTMLBuilder creates a new HTML builder
func NewHTMLBuilder() *HTMLBuilder {
	return &HTMLBuilder{
		parts: make([]string, 0, processor.HTMLBuilderInitialCapacity),
	}
}

// StartSpan starts a new span element with optional classes and styles
func (b *HTMLBuilder) StartSpan(classes []string, styles map[string]string) {
	var attrs []string

	if len(classes) > 0 {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, strings.Join(classes, " ")))
	}

	if len(styles) > 0 {
		var styleStrs []string
		for key, value := range styles {
			styleStrs = append(styleStrs, fmt.Sprintf("%s: %s", key, value))
		}
		attrs = append(attrs, fmt.Sprintf(`style="%s"`, strings.Join(styleStrs, "; ")))
	}

	if len(attrs) > 0 {
		b.parts = append(b.parts, fmt.Sprintf("<span %s>", strings.Join(attrs, " ")))
	} else {
		b.parts = append(b.parts, "<span>")
	}
}

// EndSpan ends the current span element
func (b *HTMLBuilder) EndSpan() {
	b.parts = append(b.parts, "</span>")
}

// Text adds escaped text content
func (b *HTMLBuilder) Text(content string) {
	b.parts = append(b.parts, html.EscapeString(content))
}

// Raw adds raw HTML content (use with caution)
func (b *HTMLBuilder) Raw(htmlContent string) {
	b.parts = append(b.parts, htmlContent)
}

// StartElement starts a new HTML element with optional attributes
func (b *HTMLBuilder) StartElement(tag string, attrs map[string]string) {
	if len(attrs) == 0 {
		b.parts = append(b.parts, fmt.Sprintf("<%s>", tag))
		return
	}

	var attrStrs []string
	for key, value := range attrs {
		attrStrs = append(attrStrs, fmt.Sprintf(`%s="%s"`, key, html.EscapeString(value)))
	}

	b.parts = append(b.parts, fmt.Sprintf("<%s %s>", tag, strings.Join(attrStrs, " ")))
}

// EndElement ends the current HTML element
func (b *HTMLBuilder) EndElement(tag string) {
	b.parts = append(b.parts, fmt.Sprintf("</%s>", tag))
}

// Build returns the built HTML string
func (b *HTMLBuilder) Build() string {
	return strings.Join(b.parts, "")
}

// Reset clears the builder for reuse
func (b *HTMLBuilder) Reset() {
	b.parts = b.parts[:0]
}

// Len returns the current number of parts in the builder
func (b *HTMLBuilder) Len() int {
	return len(b.parts)
}

// EscapeHTML is a utility function to escape HTML content
func EscapeHTML(content string) string {
	return html.EscapeString(content)
}

// FormatWithLineBreaks converts newlines to HTML line breaks
func FormatWithLineBreaks(content string) string {
	escaped := html.EscapeString(content)
	return strings.ReplaceAll(escaped, "\n", "<br>")
}
