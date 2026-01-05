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
	FirstName string `yaml:"firstName"`
	LastName  string `yaml:"lastName"`
	Title     string `yaml:"title"`
	Email     string `yaml:"email"`
	Phone     string `yaml:"phone"`
	Location  string `yaml:"location"`
	LinkedIn  string `yaml:"linkedin"`
	GitHub    string `yaml:"github"`
	Website   string `yaml:"website"`
	Photo     string `yaml:"photo"`
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
