![](./assets/logo-header.gif)

`resumectl` allows you to generate your resume from a YAML file to a PDF, HTML, or directly in your terminal.

## Demonstration

![](./assets/demo.gif)

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

## Shell Completion

`resumectl` supports shell completion for **Bash**, **Zsh**, **Fish**, and **PowerShell**.

### Bash

```bash
# Load completions in current session
source <(resumectl completion bash)

# Load completions for each session (Linux)
resumectl completion bash > /etc/bash_completion.d/resumectl

# Load completions for each session (macOS with Homebrew)
resumectl completion bash > $(brew --prefix)/etc/bash_completion.d/resumectl
```

### Zsh

```bash
# Enable shell completion if not already done
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Load completions for each session
resumectl completion zsh > "${fpath[1]}/_resumectl"

# Restart your shell or run: source ~/.zshrc
```

### Fish

```bash
# Load completions in current session
resumectl completion fish | source

# Load completions for each session
resumectl completion fish > ~/.config/fish/completions/resumectl.fish
```

### PowerShell

```powershell
# Load completions in current session
resumectl completion powershell | Out-String | Invoke-Expression

# Load completions for each session (add to your profile)
resumectl completion powershell > resumectl.ps1
```

## YAML Auto-Completion (Editor)

A JSON Schema is provided to enable auto-completion and validation in your editor when editing CV YAML files.

### VS Code

Add the following to your `.vscode/settings.json`:

```json
{
  "yaml.schemas": {
    "./resumectl.schema.json": ["cv.yaml", "**/cv.yaml", "examples/*.yaml"]
  }
}
```

Make sure you have the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) installed.

### IntelliJ IDEA / JetBrains IDEs

1. Open Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema Mappings
2. Add a new mapping:
   - Schema file: `resumectl.schema.json`
   - File pattern: `cv.yaml` or `*.cv.yaml`

### Neovim (with yaml-language-server)

Add to your YAML file header:

```yaml
# yaml-language-server: $schema=./resumectl.schema.json
---
personal:
  ...
```

Or configure in your LSP settings.

![](./assets/logo-footer.png)
