# resumectl

Generate HTML and PDF resumes from a YAML file.

## Installation

### Requirements

- Go 1.21+
- PDF generator: Google Chrome, Chromium, or WeasyPrint

### Build

```bash
go mod tidy
make build
```

## Usage

```bash
# Generate HTML and PDF
resumectl generate

# Generate HTML only
resumectl generate --html

# Generate PDF only
resumectl generate --pdf

# Use custom YAML file
resumectl generate -d my_cv.yaml

# Use specific theme
resumectl generate --theme elegant

# Show CV in terminal
resumectl show

# Validate YAML file
resumectl validate

# List available themes
resumectl themes
```

## Available Themes

- `modern` - Blue gradient (default)
- `classic` - Professional black
- `minimal` - Clean and simple
- `elegant` - Burgundy colors
- `tech` - Green/cyan developer style

## YAML Structure

```yaml
personal:
  firstName: John
  lastName: Doe
  title: Software Engineer
  email: john@example.com
  phone: "+1 555 123 4567"
  location: New York, USA
  linkedin: linkedin.com/in/johndoe
  github: github.com/johndoe

summary: |
  Professional summary...

experience:
  - company: Company Name
    position: Job Title
    location: City
    startDate: "2022-01"
    endDate: "present"
    highlights:
      - Achievement 1
      - Achievement 2

education:
  - institution: University
    degree: Bachelor
    field: Computer Science
    startDate: "2018"
    endDate: "2022"

skills:
  - category: Languages
    items: [Go, Python, JavaScript]

languages:
  - name: English
    level: Native

certifications:
  - name: AWS Solutions Architect
    issuer: Amazon
    date: "2023"

projects:
  - name: Project Name
    description: Description
    technologies: [Go, Docker]

interests:
  - Open Source
  - Photography
```
