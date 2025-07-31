package renderer

import (
	"embed"
	"fmt"
	"github.com/brads3290/cclogviewer/internal/constants"
	"html/template"
	"io"
	"io/fs"
	"strings"
)

//go:embed templates/*
var templateFS embed.FS

// LoadTemplates loads embedded HTML templates with custom functions.
func LoadTemplates(funcMap template.FuncMap) (*template.Template, error) {
	tmpl := template.New("").Funcs(funcMap)

	// Walk through the templates directory
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .html files
		if !strings.HasSuffix(path, constants.HTMLFileExtension) {
			return nil
		}

		// Read the file content
		content, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Get the template name (without templates/ prefix and extension)
		name := strings.TrimPrefix(path, constants.TemplateDirectoryPrefix)
		name = strings.TrimSuffix(name, constants.HTMLFileExtension)
		name = strings.ReplaceAll(name, constants.TemplateNameSeparator, "-")

		// Parse the template
		_, err = tmpl.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Define the styles template that includes all CSS files
	stylesTemplate := `{{define "styles"}}` + "\n"

	// Read all CSS files
	cssFiles := []string{
		constants.TemplateDirectoryPrefix + "styles/main.css",
		constants.TemplateDirectoryPrefix + "styles/themes.css",
		constants.TemplateDirectoryPrefix + "styles/components.css",
	}

	for _, cssFile := range cssFiles {
		content, err := templateFS.ReadFile(cssFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CSS file %s: %w", cssFile, err)
		}
		stylesTemplate += string(content) + "\n"
	}

	stylesTemplate += "{{end}}"

	_, err = tmpl.New("styles-template").Parse(stylesTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse styles template: %w", err)
	}

	// Define the scripts template that includes all JS files
	scriptsTemplate := `{{define "scripts"}}` + "\n"

	// Read all JS files
	jsFiles := []string{
		constants.TemplateDirectoryPrefix + "scripts/main.js",
	}

	for _, jsFile := range jsFiles {
		content, err := templateFS.ReadFile(jsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read JS file %s: %w", jsFile, err)
		}
		scriptsTemplate += string(content) + "\n"
	}

	scriptsTemplate += "{{end}}"

	_, err = tmpl.New("scripts-template").Parse(scriptsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scripts template: %w", err)
	}

	// Parse entry and tool-call templates with their original names
	entryContent, err := templateFS.ReadFile(constants.TemplateDirectoryPrefix + "partials/entry.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read entry template: %w", err)
	}
	_, err = tmpl.New("entry").Parse(string(entryContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse entry template: %w", err)
	}

	toolCallContent, err := templateFS.ReadFile(constants.TemplateDirectoryPrefix + "partials/tool-call.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read tool-call template: %w", err)
	}
	_, err = tmpl.New("tool-call").Parse(string(toolCallContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool-call template: %w", err)
	}

	return tmpl, nil
}

// ExecuteTemplate renders the base template with provided data.
func ExecuteTemplate(tmpl *template.Template, wr io.Writer, data interface{}) error {
	return tmpl.ExecuteTemplate(wr, "base", data)
}
