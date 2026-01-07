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
	"os"

	"resumectl/internal/generator"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the CV YAML file",
	Long: `Validate the syntax and structure of the CV YAML file
without generating output files.

Examples:
  resumectl validate
  resumectl validate -d my_cv.yaml`,
	Run: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) {
	log.Info("Validating file", "path", dataPath)

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatal("File does not exist", "path", dataPath)
	}

	gen, err := generator.New(dataPath, theme, "")
	if err != nil {
		log.Fatal("Validation error", "error", err)
	}

	cv := gen.GetCV()

	log.Info("YAML file is valid")

	log.Info("CV summary",
		"name", cv.Personal.FullName(),
		"title", cv.Personal.Title,
		"email", cv.Personal.Email,
		"experiences", len(cv.Experience),
		"education", len(cv.Education),
		"skills", len(cv.Skills),
		"languages", len(cv.Languages),
		"certifications", len(cv.Certifications),
		"projects", len(cv.Projects),
	)
}
