# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go React Generator CLI - A project scaffolding tool that generates Go backend + React frontend applications from reusable templates. The CLI uses Go's `embed.FS` to bundle templates directly into the binary, allowing users to create new projects with a single command.

## Build and Development Commands

**Version Management:**
The version is automatically set from git tags during build using `-ldflags`. Format: `{git-tag}-{commit-hash}` (e.g., `v1.0.0-abc123f`). If no git tag exists, defaults to `v0.0.0-{commit}`. Without git, shows `v0.0.0-unknown`.

```bash
# Build the generator binary (version auto-injected from git)
make build

# Run the generator with arguments
make run ARGS="--help"
make run ARGS="--name myproject --template basic"
make run ARGS="--interactive"

# Install the generator to $GOBIN
make install

# Run tests
make test

# Format code
make fmt

# Run go vet
make vet

# Clean build artifacts
make clean
```

## CLI Usage Patterns

```bash
# Interactive mode (recommended for first-time users)
./generator --interactive

# Direct project creation
./generator --name myproject --template basic

# List available templates
./generator --list

# Install custom template (not yet implemented for remote)
./generator --install /path/to/template

# Show version
./generator --version
```

## Architecture

### Core Components

**Template Manager** ([internal/template/manaager.go](internal/template/manaager.go))
- Loads embedded templates from `internal/template/templates/` using Go 1.16+ `embed.FS`
- Loads user templates from `~/.go-react-generator/templates/`
- Priority: user templates override built-in templates with the same name
- Each template must have a `template.yaml` configuration file

**Generator** ([internal/template/generator.go](internal/template/generator.go))
- Processes template files and generates project structure
- Files with `.tmpl` extension are processed as Go templates with variable substitution
- Non-template files are copied directly
- Supports file mapping rules (source → target path transformations)
- Executes post-generation commands (e.g., `go mod init`, `npm install`)

**Template Configuration** ([internal/template/config.go](internal/template/config.go))
- Defines template metadata (name, version, author, tags)
- Variables: can be required, have defaults, or be select options
- File rules: map source paths to target paths (directory or file level)
- Post-generate commands: executed in specified working directories

### Template Structure

Templates live in `internal/template/templates/{template-name}/`:
```
templates/
├── basic/
│   ├── template.yaml        # Template configuration
│   ├── backend/             # Go backend files
│   │   ├── main.go.tmpl     # Template files (processed)
│   │   └── Makefile.tmpl
│   └── frontend/            # React frontend files
└── advance/
    ├── template.yaml
    └── ...
```

**Template File Processing:**
- Files ending in `.tmpl` are processed as Go templates
- Available variables: `{{.ProjectName}}`, `{{.ModuleName}}`, `{{.Port}}`, etc.
- The `.tmpl` suffix is removed in the output filename
- Non-`.tmpl` files are copied as-is

**File Mapping Rules:**
The `files` section in `template.yaml` controls which template files are copied and where:
- `type: "directory"` - copies entire directory tree
- `type: "file"` - copies single file
- `source` and `target` define the path transformation
- Files not matching any rule are skipped (when rules are defined)

### Entry Point

**Main CLI** ([cmd/generator/main.go](cmd/generator/main.go))
- Uses `spf13/cobra` for command-line interface
- Validates environment (checks for `go` and `node` executables)
- Supports interactive mode with template selection and project naming prompts
- Shows "next steps" after generation (cd, make install, make dev, make build)

## Important Implementation Details

### Template Variable Collection
When generating a project, variables are collected in this order:
1. Default built-in variables: `ProjectName`, `ModuleName`
2. Variables from `template.yaml` with defaults
3. Interactive prompts for required variables without defaults
4. Validation for `select` type variables against defined options

### Post-Generation Commands
- Commands in `postGenerate` are executed sequentially
- Use `{{.VariableName}}` syntax for variable substitution
- Commands run in context of `workDir` (relative to project root)
- Failures are logged as warnings but don't stop generation
- Note: Uses `sh -c` which won't work on Windows without WSL/Git Bash

### User Template Installation
User can install custom templates to `~/.go-react-generator/templates/`:
- Local installation: copies template directory to user templates folder
- Remote installation: not yet implemented (returns error)
- User templates override built-in templates with the same name

## Module and Dependencies

- Go 1.24.4
- `gopkg.in/yaml.v3` - YAML parsing for template.yaml
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/pflag` - POSIX/GNU-style flags

## Known Limitations

- Remote template installation (`--install <url>`) is not implemented
- Post-generate commands use `sh -c` which requires Unix shell on Windows
- No template validation beyond YAML parsing
- No rollback mechanism if generation fails partway through
