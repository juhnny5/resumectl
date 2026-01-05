package templates

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed themes/*.css base.html
var content embed.FS

// Theme represents an available theme
type Theme struct {
	Name        string
	Description string
	CSS         string
}

// AvailableThemes returns the list of available themes
var AvailableThemes = map[string]Theme{
	"modern": {
		Name:        "modern",
		Description: "Modern theme with blue gradient (default)",
	},
	"classic": {
		Name:        "classic",
		Description: "Classic professional theme in black",
	},
	"minimal": {
		Name:        "minimal",
		Description: "Clean minimalist theme",
	},
	"elegant": {
		Name:        "elegant",
		Description: "Elegant theme with burgundy colors",
	},
	"tech": {
		Name:        "tech",
		Description: "Tech theme with green/cyan",
	},
}

// GetThemeCSS returns the CSS for a theme
func GetThemeCSS(themeName string) (string, error) {
	if _, ok := AvailableThemes[themeName]; !ok {
		return "", fmt.Errorf("theme '%s' not found. Available themes: %s", themeName, GetThemeNames())
	}

	cssPath := fmt.Sprintf("themes/%s.css", themeName)
	data, err := content.ReadFile(cssPath)
	if err != nil {
		return "", fmt.Errorf("error reading theme: %w", err)
	}

	return string(data), nil
}

// GetBaseTemplate returns the base HTML template
func GetBaseTemplate() (string, error) {
	data, err := content.ReadFile("base.html")
	if err != nil {
		return "", fmt.Errorf("error reading template: %w", err)
	}
	return string(data), nil
}

// GetCompleteTemplate returns the complete template with theme CSS
func GetCompleteTemplate(themeName string) (string, error) {
	baseHTML, err := GetBaseTemplate()
	if err != nil {
		return "", err
	}

	themeCSS, err := GetThemeCSS(themeName)
	if err != nil {
		return "", err
	}

	// Replace CSS placeholder with theme CSS
	result := strings.Replace(baseHTML, "{{THEME_CSS}}", themeCSS, 1)
	return result, nil
}

// GetParsedTemplate returns a parsed Go template with the theme
func GetParsedTemplate(themeName string, funcMap template.FuncMap) (*template.Template, error) {
	tmplContent, err := GetCompleteTemplate(themeName)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("cv").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	return tmpl, nil
}

// GetThemeNames returns theme names separated by commas
func GetThemeNames() string {
	names := make([]string, 0, len(AvailableThemes))
	for name := range AvailableThemes {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// ListThemes returns the formatted list of themes
func ListThemes() string {
	var sb strings.Builder
	sb.WriteString("Available themes:\n")
	for name, theme := range AvailableThemes {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", name, theme.Description))
	}
	return sb.String()
}
