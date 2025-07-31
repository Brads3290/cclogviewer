package diff

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

// FormatDiffHTML formats diff lines as HTML
func FormatDiffHTML(lines []DiffLine) template.HTML {
	var result strings.Builder

	result.WriteString(`<div class="diff-content unified">`)
	result.WriteString(`<div class="diff-code">`)

	for _, line := range lines {
		result.WriteString(fmt.Sprintf(`<div class="diff-line %s">`, line.Type.CSSClass()))
		result.WriteString(fmt.Sprintf(`<span class="line-number">%3d</span>`, line.LineNum))
		result.WriteString(fmt.Sprintf(`<span class="line-prefix">%s</span>`, line.Type.Prefix()))
		result.WriteString(`<span class="line-content">`)
		result.WriteString(html.EscapeString(line.Content))
		result.WriteString(`</span>`)
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	result.WriteString(`</div>`)

	return template.HTML(result.String())
}

// FormatDiffInline formats a single diff line
func FormatDiffInline(line DiffLine) string {
	return fmt.Sprintf(
		`<div class="diff-line %s"><span class="line-number">%3d</span><span class="line-prefix">%s</span><span class="line-content">%s</span></div>`,
		line.Type.CSSClass(),
		line.LineNum,
		line.Type.Prefix(),
		html.EscapeString(line.Content),
	)
}
