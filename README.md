# AAA-Generator

A application template generator for quickly scaffolding new projects with customizable templates.

## Features

- 🚀 Quick project scaffolding with interactive mode
- 📦 Multiple built-in templates (basic, advanced)
- 🔧 Customizable template system
- 📝 Automatic git initialization
- ✨ Beautiful CLI interface with progress feedback

## Installation

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd aaa-generator

# Build the binary
make build

# Install to $GOBIN
make install
```

## Usage

### Interactive Mode (Recommended)

```bash
./generator --interactive
```

### Direct Project Creation

```bash
./generator --name myproject --template basic
```

### List Available Templates

```bash
./generator --list
```

### Show Version

```bash
./generator --version
```

## Available Commands

```bash
make build         # Build the generator binary for current platform
make build-all     # Build binaries for all platforms (Linux, macOS, Windows)
make build-linux   # Build for Linux (amd64)
make build-darwin  # Build for macOS (amd64 + arm64)
make build-windows # Build for Windows (amd64)
make install       # Install to $GOBIN
make run           # Run with arguments: make run ARGS="--help"
make test          # Run tests
make fmt           # Format code
make vet           # Run go vet
make clean         # Clean build artifacts
```

### Cross-Platform Builds

Build binaries for all platforms at once:

```bash
make build-all
```

This creates binaries in the `dist/` directory:
- `generator-linux-amd64` - Linux 64-bit
- `generator-darwin-amd64` - macOS Intel
- `generator-darwin-arm64` - macOS Apple Silicon (M1/M2)
- `generator-windows-amd64.exe` - Windows 64-bit

Or build for specific platforms:

```bash
make build-linux    # Linux only
make build-darwin   # macOS only (both Intel and ARM)
make build-windows  # Windows only
```

## Version Management

The generator uses git tags for versioning. The version is automatically injected during build in the format: `{git-tag}-{commit-hash}` (e.g., `v1.0.0-abc123f`).

### Creating a New Version

**Automated Release (Recommended):**

```bash
# 1. Commit your changes
git add .
git commit -m "feat: new feature"

# 2. Create and push a version tag
git tag v1.0.0
git push origin main
git push origin v1.0.0

# 3. GitHub Actions will automatically:
#    - Build binaries for all platforms
#    - Create release archives (.tar.gz, .zip)
#    - Generate checksums
#    - Create a GitHub Release with all artifacts
```

**Manual Local Build:**

```bash
# Build with the new version locally
make build

# Verify version
./generator --version
```

### Version Tag Guidelines

- Use semantic versioning: `vMAJOR.MINOR.PATCH`
- Examples: `v1.0.0`, `v1.2.3`, `v2.0.0-beta`
- The commit hash is automatically appended during build
- If no tag exists, defaults to `v0.0.0-{commit-hash}`

## Development

### Project Structure

```
aaa-generator/
├── cmd/generator/          # CLI entry point
├── internal/
│   ├── template/          # Template management
│   │   ├── generator.go   # Project generation logic
│   │   ├── manaager.go    # Template loading
│   │   ├── config.go      # Template configuration
│   │   └── templates/     # Built-in templates
│   │       ├── basic/     # Basic Go + React template
│   │       └── advance/   # Advanced template with DB/Auth
│   └── utils/             # Utility functions
├── Makefile               # Build commands
└── CLAUDE.md              # AI assistant guidance
```

### Creating Custom Templates

Templates are stored in `internal/template/templates/`. Each template directory must contain:

1. `template.yaml` - Template configuration
2. Template files (with `.tmpl` extension for templating)
3. Static files (copied as-is)

See existing templates for examples.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## License

[Your License Here]
