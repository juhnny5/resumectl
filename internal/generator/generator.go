package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"resumectl/internal/models"
	"resumectl/internal/templates"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Generator handles CV generation
type Generator struct {
	cv          *models.CV
	theme       string
	customColor string
	outputDir   string
}

// New creates a new generator with the specified theme
func New(yamlPath, theme, outputDir string) (*Generator, error) {
	return NewWithColor(yamlPath, theme, "", outputDir)
}

// NewWithColor creates a new generator with theme and custom color
func NewWithColor(yamlPath, theme, customColor, outputDir string) (*Generator, error) {
	cv, err := loadCV(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("error loading CV: %w", err)
	}

	// Validate theme
	if theme == "" {
		theme = "modern"
	}
	if _, err := templates.GetThemeCSS(theme); err != nil {
		return nil, err
	}

	// Validate custom color
	if customColor != "" && !templates.ValidateHexColor(customColor) {
		return nil, fmt.Errorf("invalid hex color: %s (use format #RRGGBB)", customColor)
	}

	return &Generator{
		cv:          cv,
		theme:       theme,
		customColor: customColor,
		outputDir:   outputDir,
	}, nil
}

// loadCV loads the CV from a YAML file
func loadCV(path string) (*models.CV, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cv models.CV
	if err := yaml.Unmarshal(data, &cv); err != nil {
		return nil, err
	}

	return &cv, nil
}

// GenerateHTML generates the HTML file
func (g *Generator) GenerateHTML(outputPath string) error {
	// Load template with theme
	funcMap := template.FuncMap{
		"formatDate": models.FormatDate,
		"join":       strings.Join,
	}

	tmpl, err := templates.GetParsedTemplateWithColor(g.theme, g.customColor, funcMap)
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Generate HTML
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.cv); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// GeneratePDF generates the PDF file from HTML
func (g *Generator) GeneratePDF(htmlPath, pdfPath string) error {
	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(pdfPath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Try different methods to generate PDF
	methods := []struct {
		name string
		fn   func(string, string) error
	}{
		{"wkhtmltopdf", g.generateWithWkhtmltopdf},
		{"chromium", g.generateWithChromium},
		{"weasyprint", g.generateWithWeasyprint},
	}

	var lastErr error
	for _, method := range methods {
		if err := method.fn(htmlPath, pdfPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("no PDF generator available. Install wkhtmltopdf, chromium or weasyprint. Last error: %w", lastErr)
}

// generateWithWkhtmltopdf uses wkhtmltopdf to generate the PDF
func (g *Generator) generateWithWkhtmltopdf(htmlPath, pdfPath string) error {
	cmd := exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--page-size", "A4",
		"--margin-top", "0",
		"--margin-right", "0",
		"--margin-bottom", "0",
		"--margin-left", "0",
		"--encoding", "UTF-8",
		htmlPath, pdfPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wkhtmltopdf: %w - %s", err, string(output))
	}
	return nil
}

// generateWithChromium uses Chrome/Chromium to generate the PDF
func (g *Generator) generateWithChromium(htmlPath, pdfPath string) error {
	// Find Chrome/Chromium
	chromePaths := []string{
		"chromium",
		"chromium-browser",
		"google-chrome",
		"google-chrome-stable",
	}

	// Add macOS paths
	if runtime.GOOS == "darwin" {
		chromePaths = append(chromePaths,
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		)
	}

	var chromePath string
	for _, p := range chromePaths {
		if _, err := exec.LookPath(p); err == nil {
			chromePath = p
			break
		}
	}

	if chromePath == "" {
		return fmt.Errorf("chrome/chromium not found")
	}

	absHTMLPath, _ := filepath.Abs(htmlPath)
	absPDFPath, _ := filepath.Abs(pdfPath)

	cmd := exec.Command(chromePath,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--print-to-pdf="+absPDFPath,
		"--print-to-pdf-no-header",
		"file://"+absHTMLPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chromium: %w - %s", err, string(output))
	}
	return nil
}

// generateWithWeasyprint uses WeasyPrint to generate the PDF
func (g *Generator) generateWithWeasyprint(htmlPath, pdfPath string) error {
	cmd := exec.Command("weasyprint", htmlPath, pdfPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("weasyprint: %w - %s", err, string(output))
	}
	return nil
}

// GetCV returns the loaded CV
func (g *Generator) GetCV() *models.CV {
	return g.cv
}

// GetTheme returns the theme being used
func (g *Generator) GetTheme() string {
	return g.theme
}
