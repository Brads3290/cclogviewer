package ansi

import (
	"github.com/Brads3290/cclogviewer/internal/renderer/builders"
)

// ANSIConverter converts ANSI escape sequences to HTML
type ANSIConverter struct {
	parser      *ANSIParser
	colorMapper *ColorMapper
}

// NewANSIConverter creates a new ANSI to HTML converter
func NewANSIConverter() *ANSIConverter {
	return &ANSIConverter{
		parser:      NewANSIParser(),
		colorMapper: NewColorMapper(),
	}
}

// ConvertToHTML converts ANSI-formatted text to HTML
func (c *ANSIConverter) ConvertToHTML(input string) (string, error) {
	// Parse the input into tokens
	tokens, err := c.parser.Parse(input)
	if err != nil {
		return "", err
	}

	// Build HTML from tokens
	builder := builders.NewHTMLBuilder()
	state := NewANSIState()
	openSpan := false

	for _, token := range tokens {
		switch token.Type {
		case TokenEscapeSequence:
			// Close current span if open
			if openSpan {
				builder.EndSpan()
				openSpan = false
			}

			// Apply codes to state
			state.ApplyCodes(token.Codes, c.colorMapper)

		case TokenText:
			if token.Content == "" {
				continue
			}

			// Check if we need to open a new span
			classes := state.GetClasses()
			styles := state.GetStyles()

			if len(classes) > 0 || len(styles) > 0 {
				// Only open a new span if we don't have one open
				if !openSpan {
					builder.StartSpan(classes, styles)
					openSpan = true
				}
			} else if openSpan {
				// Close the span if we have no formatting
				builder.EndSpan()
				openSpan = false
			}

			// Add the text content
			builder.Text(token.Content)
		}
	}

	// Close final span if open
	if openSpan {
		builder.EndSpan()
	}

	return builder.Build(), nil
}

// ConvertToPlainText removes ANSI escape sequences and returns plain text
func (c *ANSIConverter) ConvertToPlainText(input string) string {
	return c.parser.ParseSimple(input)
}

