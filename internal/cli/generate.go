package cli

import (
	"os"
	"path/filepath"

	"resumectl/internal/generator"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	htmlOnly bool
	pdfOnly  bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate CV in HTML and/or PDF",
	Long: `Generate the CV from the specified YAML file.

By default, generates both formats (HTML and PDF).
Use --html or --pdf to generate a single format.
Use --theme to choose a theme (modern, classic, minimal, elegant, tech).
Use --color to customize the primary color of any theme.

Examples:
  resumectl generate                         # Generate HTML and PDF (modern theme)
  resumectl generate --theme elegant         # Use the elegant theme
  resumectl generate --color #ff5733         # Custom orange color
  resumectl generate --theme tech --color #8b5cf6  # Tech theme with purple
  resumectl generate --html                  # Generate HTML only
  resumectl generate --pdf                   # Generate PDF only
  resumectl generate -d my_cv.yaml           # Use a custom YAML file`,
	Run: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().BoolVar(&htmlOnly, "html", false, "Generate HTML file only")
	generateCmd.Flags().BoolVar(&pdfOnly, "pdf", false, "Generate PDF file only")
}

func runGenerate(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatal("Data file does not exist", "path", dataPath)
	}

	gen, err := generator.NewWithColor(dataPath, theme, primaryColor, outputDir)
	if err != nil {
		log.Fatal("Error", "error", err)
	}

	cv := gen.GetCV()
	if primaryColor != "" {
		log.Info("Generating CV", "name", cv.Personal.FullName(), "theme", theme, "color", primaryColor)
	} else {
		log.Info("Generating CV", "name", cv.Personal.FullName(), "theme", theme)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Error creating output directory", "error", err)
	}

	htmlPath := filepath.Join(outputDir, "cv.html")
	pdfPath := filepath.Join(outputDir, "cv.pdf")

	generateBoth := !htmlOnly && !pdfOnly

	if generateBoth || htmlOnly {
		log.Info("Generating HTML...")
		if err := gen.GenerateHTML(htmlPath); err != nil {
			log.Fatal("Error generating HTML", "error", err)
		}
		log.Info("HTML generated", "path", htmlPath)
	}

	if generateBoth || pdfOnly {
		if pdfOnly {
			if err := gen.GenerateHTML(htmlPath); err != nil {
				log.Fatal("Error generating intermediate HTML", "error", err)
			}
		}

		log.Info("Generating PDF...")
		if err := gen.GeneratePDF(htmlPath, pdfPath); err != nil {
			log.Error("Error generating PDF", "error", err)
			log.Info("Tip: Install Google Chrome or WeasyPrint")
			log.Info("  - macOS/Linux: Chrome is often already installed")
			log.Info("  - pip install weasyprint")
			os.Exit(1)
		}
		log.Info("PDF generated", "path", pdfPath)
	}

	log.Info("Generation completed successfully")
}
