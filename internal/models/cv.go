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

// CV représente la structure complète d'un CV
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

// Personal contient les informations personnelles
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

// Experience représente une expérience professionnelle
type Experience struct {
	Company     string   `yaml:"company"`
	Position    string   `yaml:"position"`
	Location    string   `yaml:"location"`
	StartDate   string   `yaml:"startDate"`
	EndDate     string   `yaml:"endDate"`
	Description string   `yaml:"description"`
	Highlights  []string `yaml:"highlights"`
}

// Education représente une formation
type Education struct {
	Institution string `yaml:"institution"`
	Degree      string `yaml:"degree"`
	Field       string `yaml:"field"`
	Location    string `yaml:"location"`
	StartDate   string `yaml:"startDate"`
	EndDate     string `yaml:"endDate"`
	Description string `yaml:"description"`
}

// SkillCategory représente une catégorie de compétences
type SkillCategory struct {
	Category string   `yaml:"category"`
	Items    []string `yaml:"items"`
}

// Language représente une langue parlée
type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

// Certification représente une certification
type Certification struct {
	Name   string `yaml:"name"`
	Issuer string `yaml:"issuer"`
	Date   string `yaml:"date"`
}

// Project représente un projet personnel
type Project struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	URL          string   `yaml:"url"`
	Technologies []string `yaml:"technologies"`
}

// FullName retourne le nom complet
func (p Personal) FullName() string {
	return p.FirstName + " " + p.LastName
}

// FormatDate formate une date pour l'affichage
func FormatDate(date string) string {
	if date == "present" || date == "présent" {
		return "Présent"
	}
	return date
}
