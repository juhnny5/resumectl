// Copyright (c) 2026 Julien Briault
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

Usage examples:
  resumectl themes                   # List all available themes
  resumectl generate --theme elegant # Use a specific theme`,
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
