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

![](./img/logo-footer.png)
