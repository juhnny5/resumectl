![](./img/logo-header.gif)

`resumectl` allows you to generate your resume from a YAML file to a PDF, HTML, or directly in your terminal.

## Demonstration

![](./img/demo.gif)

## Available Themes

You can view the available themes by looking at [THEMES.md](./THEMES.md).

## Examples

For reference, you can use the examples in `examples/`.

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

### Initialize a CV from LinkedIn and GitHub

The `init` command allows you to quickly bootstrap a CV YAML file by importing data from your LinkedIn profile and/or GitHub projects.

```bash
# Create an empty CV template
resumectl init

# Initialize from LinkedIn profile (public data only)
resumectl init --linkedin johndoe

# Initialize with full LinkedIn data (requires authentication)
resumectl init --linkedin johndoe --cookie "YOUR_LI_AT_COOKIE"

# Add top GitHub projects
resumectl init --github yourusername

# Combine LinkedIn and GitHub
resumectl init --linkedin johndoe --cookie "AQEDAx..." --github yourusername

# Customize number of GitHub projects (default: 5)
resumectl init --github yourusername --projects 10

# Custom output file
resumectl init -f my-cv.yaml
```

#### Getting your LinkedIn cookie

To access full LinkedIn profile data (experiences, education, skills, etc.), you need to provide your `li_at` session cookie:

1. Log in to LinkedIn in your browser
2. Open Developer Tools (F12)
3. Go to Application > Cookies > linkedin.com
4. Copy the value of the `li_at` cookie

### Generate your CV

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
```

### Other commands

```bash
# Show CV in terminal
resumectl show

# Validate YAML file
resumectl validate

# List available themes
resumectl themes

# Live preview with hot reload
resumectl serve
```

![](./img/logo-footer.png)
