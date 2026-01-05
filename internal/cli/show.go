package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"resumectl/internal/generator"
	"resumectl/internal/models"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	showStyle   string
	showPager   bool
	forceGlow   bool
	forceInline bool
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display CV in the terminal",
	Long: `Display the formatted CV directly in the terminal using
glow (if installed) or glamour as fallback.

Glow offers an interactive experience with built-in pager.
Glamour is used if glow is not available or with --inline.

Available styles: auto, dark, light, dracula, tokyo-night, notty

Examples:
  resumectl show
  resumectl show --style dracula
  resumectl show --pager        # Force glow pager usage
  resumectl show --inline       # Force inline display (glamour)
  resumectl show -d my_cv.yaml`,
	Run: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringVarP(&showStyle, "style", "s", "auto", "Display style (auto, dark, light, dracula, tokyo-night, notty)")
	showCmd.Flags().BoolVarP(&showPager, "pager", "p", false, "Force glow pager usage")
	showCmd.Flags().BoolVar(&forceInline, "inline", false, "Force inline display without pager (glamour)")
}

// glowAvailable checks if glow is installed on the system
func glowAvailable() bool {
	_, err := exec.LookPath("glow")
	return err == nil
}

// renderWithGlow uses glow to display markdown with pager
func renderWithGlow(markdown string) error {
	// Create a temporary file for markdown
	tmpFile, err := os.CreateTemp("", "cv-*.md")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(markdown); err != nil {
		return fmt.Errorf("error writing markdown: %w", err)
	}
	tmpFile.Close()

	// Build the glow command
	args := []string{}
	if showPager {
		args = append(args, "--pager")
	}

	// Map glamour styles to glow
	glowStyle := mapStyleToGlow(showStyle)
	if glowStyle != "" {
		args = append(args, "--style", glowStyle)
	}

	args = append(args, tmpFile.Name())

	cmd := exec.Command("glow", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// mapStyleToGlow converts glamour styles to glow styles
func mapStyleToGlow(glamourStyle string) string {
	switch glamourStyle {
	case "auto":
		return "auto"
	case "dark":
		return "dark"
	case "light":
		return "light"
	case "dracula":
		return "dracula"
	case "tokyo-night":
		return "tokyo-night"
	case "notty":
		return "notty"
	default:
		return "auto"
	}
}

// renderWithGlamour uses glamour for inline display
func renderWithGlamour(markdown string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(showStyle),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return fmt.Errorf("error initializing renderer: %w", err)
	}

	out, err := renderer.Render(markdown)
	if err != nil {
		return fmt.Errorf("error rendering: %w", err)
	}

	fmt.Print(out)
	return nil
}

func runShow(cmd *cobra.Command, args []string) {
	// Resolve absolute path of data file
	absDataPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Fatal("Error resolving path", "path", dataPath, "error", err)
	}

	if _, err := os.Stat(absDataPath); os.IsNotExist(err) {
		log.Fatal("Data file does not exist", "path", absDataPath)
	}

	gen, err := generator.New(absDataPath, theme, outputDir)
	if err != nil {
		log.Fatal("Error", "error", err)
	}

	cv := gen.GetCV()
	markdown := generateMarkdown(cv)

	// Determine rendering mode
	useGlow := glowAvailable() && !forceInline

	if useGlow {
		log.Debug("Using glow for rendering")
		if err := renderWithGlow(markdown); err != nil {
			log.Warn("Error with glow, falling back to glamour", "error", err)
			if err := renderWithGlamour(markdown); err != nil {
				log.Fatal("Error rendering", "error", err)
			}
		}
	} else {
		log.Debug("Using glamour for rendering")
		if err := renderWithGlamour(markdown); err != nil {
			log.Fatal("Erreur lors du rendu", "error", err)
		}
	}
}

func generateMarkdown(cv *models.CV) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("# %s\n\n", cv.Personal.FullName()))
	buf.WriteString(fmt.Sprintf("### %s\n\n", cv.Personal.Title))

	buf.WriteString("---\n\n")
	var contacts []string
	if cv.Personal.Email != "" {
		contacts = append(contacts, cv.Personal.Email)
	}
	if cv.Personal.Phone != "" {
		contacts = append(contacts, cv.Personal.Phone)
	}
	if cv.Personal.Location != "" {
		contacts = append(contacts, cv.Personal.Location)
	}
	if cv.Personal.LinkedIn != "" {
		contacts = append(contacts, cv.Personal.LinkedIn)
	}
	if cv.Personal.GitHub != "" {
		contacts = append(contacts, cv.Personal.GitHub)
	}
	if cv.Personal.Website != "" {
		contacts = append(contacts, cv.Personal.Website)
	}
	buf.WriteString(strings.Join(contacts, " | ") + "\n\n")

	if cv.Summary != "" {
		buf.WriteString("## Summary\n\n")
		buf.WriteString(strings.TrimSpace(cv.Summary) + "\n\n")
	}

	if len(cv.Experience) > 0 {
		buf.WriteString("## Professional Experience\n\n")
		for _, exp := range cv.Experience {
			buf.WriteString(fmt.Sprintf("### %s - *%s*\n", exp.Position, exp.Company))
			buf.WriteString(fmt.Sprintf("%s - %s | %s\n\n", exp.StartDate, exp.EndDate, exp.Location))
			if exp.Description != "" {
				buf.WriteString(strings.TrimSpace(exp.Description) + "\n\n")
			}
			if len(exp.Highlights) > 0 {
				for _, h := range exp.Highlights {
					buf.WriteString(fmt.Sprintf("- %s\n", h))
				}
				buf.WriteString("\n")
			}
		}
	}

	if len(cv.Education) > 0 {
		buf.WriteString("## Education\n\n")
		for _, edu := range cv.Education {
			buf.WriteString(fmt.Sprintf("### %s - %s\n", edu.Degree, edu.Field))
			buf.WriteString(fmt.Sprintf("%s | %s - %s\n\n", edu.Institution, edu.StartDate, edu.EndDate))
			if edu.Description != "" {
				buf.WriteString(strings.TrimSpace(edu.Description) + "\n\n")
			}
		}
	}

	if len(cv.Skills) > 0 {
		buf.WriteString("## Skills\n\n")
		for _, skill := range cv.Skills {
			buf.WriteString(fmt.Sprintf("**%s:** %s\n\n", skill.Category, strings.Join(skill.Items, " | ")))
		}
	}

	if len(cv.Languages) > 0 {
		buf.WriteString("## Languages\n\n")
		for _, lang := range cv.Languages {
			buf.WriteString(fmt.Sprintf("- **%s:** %s\n", lang.Name, lang.Level))
		}
		buf.WriteString("\n")
	}

	if len(cv.Certifications) > 0 {
		buf.WriteString("## Certifications\n\n")
		for _, cert := range cv.Certifications {
			buf.WriteString(fmt.Sprintf("- **%s** - %s (%s)\n", cert.Name, cert.Issuer, cert.Date))
		}
		buf.WriteString("\n")
	}

	if len(cv.Projects) > 0 {
		buf.WriteString("## Projects\n\n")
		for _, proj := range cv.Projects {
			buf.WriteString(fmt.Sprintf("### %s\n", proj.Name))
			buf.WriteString(fmt.Sprintf("%s\n\n", proj.Description))
			if len(proj.Technologies) > 0 {
				buf.WriteString(fmt.Sprintf("*Technologies:* %s\n\n", strings.Join(proj.Technologies, ", ")))
			}
		}
	}

	if len(cv.Interests) > 0 {
		buf.WriteString("## Interests\n\n")
		buf.WriteString(strings.Join(cv.Interests, " | ") + "\n")
	}

	return buf.String()
}
