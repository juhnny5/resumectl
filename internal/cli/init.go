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
	"path/filepath"
	"strings"

	"resumectl/internal/github"
	"resumectl/internal/linkedin"
	"resumectl/internal/models"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	linkedinURL    string
	outputFile     string
	forceOverwrite bool
	linkedinCookie string
	githubUsername string
	githubProjects int
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new CV YAML file",
	Long: `Initialize a new CV YAML file with optional LinkedIn and GitHub profile import.

This command creates a new cv.yaml file that you can edit with your information.
If a LinkedIn profile URL is provided, it will fetch public information
from the profile to pre-populate the CV.

If a GitHub username is provided, it will fetch your top projects by stars
and add them to the projects section.

To get ALL LinkedIn profile data (experiences, education, etc.), you need to provide your
LinkedIn session cookie (li_at). To get it:
  1. Log in to LinkedIn in your browser
  2. Open Developer Tools (F12) > Application > Cookies > linkedin.com
  3. Copy the value of the 'li_at' cookie

Usage examples:
  resumectl init                                     # Create an empty template
  resumectl init --linkedin https://linkedin.com/in/johndoe
  resumectl init --linkedin johndoe --cookie "AQEDAx..."  # With auth for full data
  resumectl init --github juhnny5                    # Add top GitHub projects
  resumectl init --github juhnny5 --projects 10     # Add top 10 projects
  resumectl init -f my-cv.yaml                       # Custom output file
  resumectl init --force                             # Overwrite existing file`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&linkedinURL, "linkedin", "l", "", "LinkedIn profile URL or username")
	initCmd.Flags().StringVarP(&outputFile, "file", "f", "cv.yaml", "Output file name")
	initCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing file without confirmation")
	initCmd.Flags().StringVarP(&linkedinCookie, "cookie", "c", "", "LinkedIn session cookie (li_at) for full data access")
	initCmd.Flags().StringVarP(&githubUsername, "github", "g", "", "GitHub username to fetch top projects")
	initCmd.Flags().IntVarP(&githubProjects, "projects", "p", 5, "Number of top GitHub projects to fetch (default: 5)")
}

func runInit(cmd *cobra.Command, args []string) {
	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil && !forceOverwrite {
		log.Fatal("File already exists. Use --force to overwrite", "file", outputFile)
	}

	var cv *models.CV

	if linkedinURL != "" {
		log.Info("Fetching LinkedIn profile...")

		// Extract the username
		username, err := linkedin.ExtractUsernameFromURL(linkedinURL)
		if err != nil {
			log.Warn("Could not extract username, using as-is", "error", err)
			username = linkedinURL
		}

		log.Info("Looking up profile", "username", username)

		// Fetch LinkedIn profile (with or without authentication)
		var profile *linkedin.LinkedInProfile
		if linkedinCookie != "" {
			log.Info("Using authenticated session for full data access...")
			linkedin.DebugMode = DebugMode
			profile, err = linkedin.FetchProfileWithAuth(username, linkedinCookie)
		} else {
			profile, err = linkedin.FetchProfile(username)
		}

		if err != nil {
			log.Warn("Could not fetch LinkedIn profile, creating template instead", "error", err)
			cv = createEmptyCV()
		} else {
			log.Info("Profile found", "name", profile.FirstName+" "+profile.LastName)
			cv = profile.ToCV(linkedinURL)

			// Warn about LinkedIn limitations if no cookie
			if linkedinCookie == "" {
				log.Warn("LinkedIn limits public data access. Some information may be missing or incomplete.")
				fmt.Println("")
				fmt.Println("  ⚠️  Note: LinkedIn masks most profile data for non-authenticated visitors.")
				fmt.Println("      To get ALL data, use: --cookie <your_li_at_cookie>")
				fmt.Println("")
				fmt.Println("      How to get your cookie:")
				fmt.Println("      1. Log in to LinkedIn in your browser")
				fmt.Println("      2. Open DevTools (F12) > Application > Cookies > linkedin.com")
				fmt.Println("      3. Copy the value of 'li_at' cookie")
				fmt.Println("")
			} else {
				log.Info("Full profile data retrieved successfully!")
			}
		}
	} else {
		cv = createEmptyCV()
	}

	// Fetch GitHub projects if a username is provided
	if githubUsername != "" {
		log.Info("Fetching GitHub projects...")

		username := github.ExtractUsernameFromURL(githubUsername)
		log.Info("Looking up GitHub profile", "username", username)

		github.DebugMode = DebugMode
		projects, err := github.FetchTopProjects(username, githubProjects)
		if err != nil {
			log.Warn("Could not fetch GitHub projects", "error", err)
		} else {
			log.Info("GitHub projects fetched successfully!", "count", len(projects))

			// Convert and add projects to CV
			for _, proj := range projects {
				cvProject := models.Project{
					Name:         proj.Name,
					Description:  proj.Description,
					URL:          proj.URL,
					Technologies: proj.Technologies,
				}
				cv.Projects = append(cv.Projects, cvProject)
			}

			// Update GitHub link in personal info
			if cv.Personal.GitHub == "" || cv.Personal.GitHub == "github.com/yourusername" {
				cv.Personal.GitHub = "github.com/" + username
			}
		}
	}

	// Generate YAML file
	if err := writeCV(cv, outputFile); err != nil {
		log.Fatal("Error writing CV file", "error", err)
	}

	absPath, _ := filepath.Abs(outputFile)
	log.Info("CV file created successfully!", "path", absPath)
	log.Info("Next steps:")
	fmt.Println("  1. Edit the file with your information: " + outputFile)
	fmt.Println("  2. Generate your CV: resumectl generate")
	fmt.Println("  3. Preview in browser: resumectl serve")
}

// createEmptyCV creates an empty CV template
func createEmptyCV() *models.CV {
	return &models.CV{
		Personal: models.Personal{
			FirstName:      "John",
			LastName:       "Doe",
			Title:          "Your Professional Title",
			Email:          "your.email@example.com",
			Phone:          "+1 000 000 0000",
			Location:       "City, Country",
			LinkedIn:       "linkedin.com/in/yourprofile",
			GitHub:         "github.com/yourusername",
			Website:        "yourwebsite.com",
			Photo:          "",
			PhotoGrayscale: false,
			PhotoShape:     "round",
		},
		Summary: "Write a brief professional summary highlighting your key skills, experience, and career objectives. " +
			"This section should give employers a quick overview of who you are and what you bring to the table.",
		Experience: []models.Experience{
			{
				Company:   "Company Name",
				Position:  "Senior Position",
				Location:  "City, Country",
				StartDate: "2022-01",
				EndDate:   "present",
				Description: "Brief description of your role and main responsibilities " +
					"in this position.",
				Highlights: []string{
					"Key achievement or responsibility 1",
					"Key achievement or responsibility 2",
					"Key achievement or responsibility 3",
				},
			},
			{
				Company:   "Previous Company",
				Position:  "Position Title",
				Location:  "City, Country",
				StartDate: "2019-06",
				EndDate:   "2021-12",
				Description: "Brief description of your role and main responsibilities " +
					"in this position.",
				Highlights: []string{
					"Key achievement or responsibility 1",
					"Key achievement or responsibility 2",
				},
			},
		},
		Education: []models.Education{
			{
				Institution: "University Name",
				Degree:      "Master's Degree",
				Field:       "Field of Study",
				Location:    "City, Country",
				StartDate:   "2015",
				EndDate:     "2019",
				Description: "Relevant coursework, honors, or achievements",
			},
		},
		Skills: []models.SkillCategory{
			{
				Category: "Programming Languages",
				Items:    []string{"Language 1", "Language 2", "Language 3"},
			},
			{
				Category: "Frameworks & Tools",
				Items:    []string{"Framework 1", "Framework 2", "Tool 1"},
			},
			{
				Category: "Soft Skills",
				Items:    []string{"Communication", "Leadership", "Problem Solving"},
			},
		},
		Languages: []models.Language{
			{Name: "English", Level: "Native"},
			{Name: "French", Level: "Fluent (C1)"},
		},
		Certifications: []models.Certification{
			{
				Name:   "Certification Name",
				Issuer: "Issuing Organization",
				Date:   "2023",
			},
		},
		Projects: []models.Project{
			{
				Name:         "Project Name",
				Description:  "Brief description of the project and your role",
				URL:          "github.com/username/project",
				Technologies: []string{"Tech 1", "Tech 2", "Tech 3"},
			},
		},
		Interests: []string{"Interest 1", "Interest 2", "Interest 3"},
	}
}

// writeCV writes the CV to a YAML file
func writeCV(cv *models.CV, filename string) error {
	// Create parent directory if needed
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Serialize to YAML
	data, err := yaml.Marshal(cv)
	if err != nil {
		return fmt.Errorf("failed to marshal CV: %w", err)
	}

	// Ajouter le header de licence et les commentaires
	header := `# CV Configuration File
# Generated by resumectl - https://github.com/juhnny5/resumectl
#
# Edit this file with your personal information, then run:
#   resumectl generate          # Generate HTML and PDF
#   resumectl serve             # Preview in browser with live reload
#   resumectl themes            # List available themes
#
# For more information, see: https://github.com/juhnny5/resumectl

`
	// Ajouter des commentaires explicatifs
	content := string(data)
	content = addYAMLComments(content)

	finalContent := header + content

	// Écrire le fichier
	if err := os.WriteFile(filename, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// addYAMLComments adds explanatory comments to the YAML
func addYAMLComments(content string) string {
	replacements := []struct {
		pattern string
		replace string
	}{
		{"personal:", "# Personal Information\npersonal:"},
		{"summary:", "\n# Professional Summary\nsummary:"},
		{"experience:", "\n# Work Experience\nexperience:"},
		{"education:", "\n# Education\neducation:"},
		{"skills:", "\n# Skills (grouped by category)\nskills:"},
		{"languages:", "\n# Languages\nlanguages:"},
		{"certifications:", "\n# Certifications\ncertifications:"},
		{"projects:", "\n# Personal Projects\nprojects:"},
		{"interests:", "\n# Interests (optional)\ninterests:"},
	}

	result := content
	for _, r := range replacements {
		result = strings.Replace(result, r.pattern, r.replace, 1)
	}

	return result
}
