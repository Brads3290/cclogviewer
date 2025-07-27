package renderer

import (
	"cclogviewer/internal/models"
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