package cli

import (
	"fmt"
	"os"

	"resumectl/internal/templates"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	dataPath  string
	outputDir string
	theme     string
)

var rootCmd = &cobra.Command{
	Use:   "resumectl",
	Short: "HTML and PDF resume generator",
	Long: `resumectl - HTML and PDF resume generator from a YAML file

This program generates a professional resume in HTML and PDF formats
from a YAML configuration file containing your information.

Templates are embedded in the binary with multiple themes available.

Usage examples:
  resumectl generate                   # Generate HTML and PDF (modern theme)
  resumectl generate --theme elegant   # Use the elegant theme
  resumectl generate --html            # Generate HTML only
  resumectl themes                     # List available themes`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SetVersion(v string) {
	Version = v
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dataPath, "data", "d", "cv.yaml", "Path to the CV YAML file")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "output", "Output directory")
	rootCmd.PersistentFlags().StringVar(&theme, "theme", "modern", "CV theme ("+templates.GetThemeNames()+")")
}
