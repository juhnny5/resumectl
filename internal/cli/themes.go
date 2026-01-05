package cli

import (
	"fmt"

	"resumectl/internal/templates"

	"github.com/spf13/cobra"
)

var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Long: `Display the list of available themes for CV generation.

Each theme offers a different visual style for your HTML and PDF CV.

Examples:
  resumectl themes
  resumectl generate --theme elegant`,
	Run: runThemes,
}

func init() {
	rootCmd.AddCommand(themesCmd)
}

func runThemes(cmd *cobra.Command, args []string) {
	fmt.Println("Available themes for resumectl:")
	fmt.Println()

	for name, theme := range templates.AvailableThemes {
		marker := " "
		if name == "modern" {
			marker = "*"
		}
		fmt.Printf("  %s %-10s  %s\n", marker, name, theme.Description)
	}

	fmt.Println()
	fmt.Println("  * = default theme")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  resumectl generate --theme <name>")
}
