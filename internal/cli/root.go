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
	"os"

	"resumectl/internal/templates"

	"github.com/spf13/cobra"
)

var (
	Version      = "dev"
	dataPath     string
	outputDir    string
	theme        string
	primaryColor string
)

var rootCmd = &cobra.Command{
	Use:   "resumectl",
	Short: "HTML and PDF resume generator",
	Long: `resumectl - HTML and PDF resume generator from a YAML file

This program generates a professional resume in HTML and PDF formats
from a YAML configuration file containing your information.

Templates are embedded in the binary with multiple themes available.

Usage examples:
  resumectl generate                        # Generate HTML and PDF (modern theme)
  resumectl generate --theme elegant        # Use the elegant theme
  resumectl generate --color #ff5733        # Custom color (any theme)
  resumectl generate --theme tech --color #8b5cf6  # Tech theme with purple
  resumectl generate --html                 # Generate HTML only
  resumectl themes                          # List available themes`,
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
	rootCmd.PersistentFlags().StringVar(&primaryColor, "color", "", "Custom primary color for any theme (hex, e.g. #ff5733)")
}
