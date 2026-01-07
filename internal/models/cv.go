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

package models

// CV represents the complete structure of a resume
type CV struct {
	Personal       Personal        `yaml:"personal"`
	Summary        string          `yaml:"summary"`
	Experience     []Experience    `yaml:"experience"`
	Education      []Education     `yaml:"education"`
	Skills         []SkillCategory `yaml:"skills"`
	Languages      []Language      `yaml:"languages"`
	Certifications []Certification `yaml:"certifications"`
	Projects       []Project       `yaml:"projects"`
	Interests      []string        `yaml:"interests"`
}

// Personal contains personal information
type Personal struct {
	FirstName      string `yaml:"firstName"`
	LastName       string `yaml:"lastName"`
	Title          string `yaml:"title"`
	Email          string `yaml:"email"`
	Phone          string `yaml:"phone"`
	Location       string `yaml:"location"`
	LinkedIn       string `yaml:"linkedin"`
	GitHub         string `yaml:"github"`
	Website        string `yaml:"website"`
	Photo          string `yaml:"photo"`
	PhotoGrayscale bool   `yaml:"photoGrayscale"` // Black and white filter
	PhotoShape     string `yaml:"photoShape"`     // "round" (default) or "square"
}

// Experience represents a work experience
type Experience struct {
	Company     string   `yaml:"company"`
	Position    string   `yaml:"position"`
	Location    string   `yaml:"location"`
	StartDate   string   `yaml:"startDate"`
	EndDate     string   `yaml:"endDate"`
	Description string   `yaml:"description"`
	Highlights  []string `yaml:"highlights"`
}

// Education represents an educational background
type Education struct {
	Institution string `yaml:"institution"`
	Degree      string `yaml:"degree"`
	Field       string `yaml:"field"`
	Location    string `yaml:"location"`
	StartDate   string `yaml:"startDate"`
	EndDate     string `yaml:"endDate"`
	Description string `yaml:"description"`
}

// SkillCategory represents a skill category
type SkillCategory struct {
	Category string   `yaml:"category"`
	Items    []string `yaml:"items"`
}

// Language represents a spoken language
type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

// Certification represents a certification
type Certification struct {
	Name   string `yaml:"name"`
	Issuer string `yaml:"issuer"`
	Date   string `yaml:"date"`
}

// Project represents a personal project
type Project struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	URL          string   `yaml:"url"`
	Technologies []string `yaml:"technologies"`
}

// FullName returns the full name
func (p Personal) FullName() string {
	return p.FirstName + " " + p.LastName
}

// FormatDate formats a date for display
func FormatDate(date string) string {
	if date == "present" || date == "présent" {
		return "Présent"
	}
	return date
}
