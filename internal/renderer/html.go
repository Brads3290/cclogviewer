package renderer

import (
	"cclogviewer/internal/models"
	"fmt"
	"html/template"
	"os"
)

// GenerateHTML generates an HTML file from processed entries
func GenerateHTML(entries []*models.ProcessedEntry, outputFile string) error {
	// Create custom function map
	funcMap := template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
		"formatNumber": func(n int) string {
			if n < 1000 {
				return fmt.Sprintf("%d", n)
			}
			// Format with thousands separators
			str := fmt.Sprintf("%d", n)
			result := ""
			for i, digit := range str {
				if i > 0 && (len(str)-i)%3 == 0 {
					result += ","
				}
				result += string(digit)
			}
			return result
		},
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, entries)
}