package templates

import (
	"embed"
	"fmt"
	"html/template"
	"regexp"
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

// ColorScheme holds derived colors from a primary color
type ColorScheme struct {
	Primary   string
	Secondary string
	Accent    string
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

// ValidateHexColor checks if a string is a valid hex color
func ValidateHexColor(color string) bool {
	if color == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{3})$`, color)
	return matched
}

// DeriveColorScheme generates secondary and accent colors from primary
func DeriveColorScheme(primary string) ColorScheme {
	// Darken for secondary, lighten for accent
	return ColorScheme{
		Primary:   primary,
		Secondary: darkenColor(primary, 0.2),
		Accent:    lightenColor(primary, 0.15),
	}
}

// darkenColor darkens a hex color by a factor (0-1)
func darkenColor(hex string, factor float64) string {
	r, g, b := hexToRGB(hex)
	r = int(float64(r) * (1 - factor))
	g = int(float64(g) * (1 - factor))
	b = int(float64(b) * (1 - factor))
	return rgbToHex(r, g, b)
}

// lightenColor lightens a hex color by a factor (0-1)
func lightenColor(hex string, factor float64) string {
	r, g, b := hexToRGB(hex)
	r = r + int(float64(255-r)*factor)
	g = g + int(float64(255-g)*factor)
	b = b + int(float64(255-b)*factor)
	return rgbToHex(r, g, b)
}

// hexToRGB converts hex color to RGB
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// rgbToHex converts RGB to hex color
func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", clamp(r), clamp(g), clamp(b))
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
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

// GetThemeCSSWithColor returns the CSS with custom primary color
func GetThemeCSSWithColor(themeName, customColor string) (string, error) {
	css, err := GetThemeCSS(themeName)
	if err != nil {
		return "", err
	}

	// Apply custom color to any theme
	if customColor != "" {
		if !ValidateHexColor(customColor) {
			return "", fmt.Errorf("invalid hex color: %s (use format #RRGGBB)", customColor)
		}
		scheme := DeriveColorScheme(customColor)

		// Use regex to replace color values regardless of original color
		primaryRe := regexp.MustCompile(`--primary-color:\s*#[0-9A-Fa-f]{6};`)
		secondaryRe := regexp.MustCompile(`--secondary-color:\s*#[0-9A-Fa-f]{6};`)
		accentRe := regexp.MustCompile(`--accent-color:\s*#[0-9A-Fa-f]{6};`)

		css = primaryRe.ReplaceAllString(css, fmt.Sprintf("--primary-color: %s;", scheme.Primary))
		css = secondaryRe.ReplaceAllString(css, fmt.Sprintf("--secondary-color: %s;", scheme.Secondary))
		css = accentRe.ReplaceAllString(css, fmt.Sprintf("--accent-color: %s;", scheme.Accent))
	}

	return css, nil
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
	return GetCompleteTemplateWithColor(themeName, "")
}

// GetCompleteTemplateWithColor returns the complete template with theme CSS and custom color
func GetCompleteTemplateWithColor(themeName, customColor string) (string, error) {
	baseHTML, err := GetBaseTemplate()
	if err != nil {
		return "", err
	}

	themeCSS, err := GetThemeCSSWithColor(themeName, customColor)
	if err != nil {
		return "", err
	}

	// Replace CSS placeholder with theme CSS
	result := strings.Replace(baseHTML, "{{THEME_CSS}}", themeCSS, 1)
	return result, nil
}

// GetParsedTemplate returns a parsed Go template with the theme
func GetParsedTemplate(themeName string, funcMap template.FuncMap) (*template.Template, error) {
	return GetParsedTemplateWithColor(themeName, "", funcMap)
}

// GetParsedTemplateWithColor returns a parsed Go template with the theme and custom color
func GetParsedTemplateWithColor(themeName, customColor string, funcMap template.FuncMap) (*template.Template, error) {
	tmplContent, err := GetCompleteTemplateWithColor(themeName, customColor)
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
