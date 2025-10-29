# Docgen-Tool

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/dl/)
[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/kyivinua/docgen-tool/releases)

**Enterprise-grade documentation generator for Protocol Buffer services**

Docgen-tool automatically generates comprehensive, high-quality documentation for your Protocol Buffer services with AI-powered enrichment, quality validation, and beautiful diagram generation.

## Features

- **🤖 AI-Powered Enrichment** - Automatic description enhancement using Claude AI
- **✅ Quality Validation** - Comprehensive coverage and quality metrics with configurable thresholds
- **📊 Diagram Generation** - Beautiful Mermaid diagrams for services, messages, and sequences
- **🔧 Auto-Fix** - Automatic quality issue resolution
- **⚡ High Performance** - Parallel processing for fast generation
- **🔒 Security-First** - Built with security best practices and input validation
- **🎨 Customizable** - Flexible templates and configuration options
- **📈 Quality Gates** - Enforce documentation standards across your team

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/kyivinua/docgen-tool.git
cd docgen-tool
make build
make install

# Or use the installation script
curl -sSL https://raw.githubusercontent.com/kyivinua/docgen-tool/main/scripts/install.sh | bash
```

### Usage

```bash
# Initialize configuration
docgen init

# Generate documentation
docgen generate --config ./docgen.yaml

# With AI enrichment
export ANTHROPIC_API_KEY=your-api-key
docgen generate --config ./docgen.yaml --ai

# With auto-fix
docgen generate --config ./docgen.yaml --auto-fix
```

## Example Output

After running docgen, you'll see:

```
✅ Documentation generated successfully!

📊 Summary:
  Services:           3
  Coverage Score:     95.5%
  Quality Score:      92.3%
  Quality Gate:       ✅ Passed
  Generation Time:    2.5s
  Output Directory:   ./docs/generated
```

## Configuration

Create a `docgen.yaml` file:

```yaml
version: "1.0"
project_name: "my-project"
output_dir: "./docs/generated"

services:
  - name: "UserService"
    proto_files:
      - "./proto/user/v1/service.proto"
    enabled: true

quality:
  coverage_target: 80
  description_quality_target: 85
  auto_fix: true

enricher:
  provider: "claude"
  api_key: "${ANTHROPIC_API_KEY}"
  enabled: false

diagrams:
  enabled: true
  service_graph: true
  message_graph: true
```

See [configs/example.yaml](configs/example.yaml) for all options.

## Documentation

- **[Quick Start Guide](docs/QUICKSTART.md)** - Get started in 5 minutes
- **[Architecture](docs/ARCHITECTURE.md)** - Detailed architecture documentation
- **[Example Configuration](configs/example.yaml)** - Complete configuration reference

## Key Components

### Generator
Orchestrates the entire documentation generation process with parallel processing and error handling.

### Validator
Validates documentation quality with configurable metrics:
- Coverage score (% of documented items)
- Description quality score (0-100)
- Method and field coverage
- Quality gates with auto-fix

### Enricher
AI-powered content enrichment using Claude:
- Intelligent caching
- Rate limiting
- Parallel requests
- Fallback descriptions

### Diagram Generator
Creates Mermaid diagrams:
- Service architecture
- Message structures
- Sequence diagrams
- Multi-service overviews

## Project Structure

```
docgen-tool/
├── cmd/docgen/          # CLI entry point
├── internal/            # Private application code
│   ├── docgen/          # Core data models
│   ├── config/          # Configuration management
│   ├── generator/       # Main generator
│   ├── validator/       # Quality validation
│   ├── enricher/        # AI enrichment
│   ├── fileutil/        # File operations
│   └── diagrams/        # Diagram generation
├── configs/             # Example configurations
├── docs/                # Documentation
├── scripts/             # Utility scripts
└── Makefile             # Build automation
```

## Quality Metrics

Docgen-tool provides comprehensive quality metrics:

- **Coverage Score** - Percentage of documented items
- **Description Quality** - Quality score based on length, structure, and detail
- **Method Coverage** - Percentage of documented methods
- **Field Coverage** - Percentage of documented fields

## Build Commands

```bash
make help           # Show all available commands
make build          # Build the binary
make install        # Install to system
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make fmt            # Format code
make clean          # Clean artifacts
make release        # Create release builds
```

## Requirements

- Go 1.22 or later
- (Optional) Anthropic API key for AI enrichment

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Zerolog](https://github.com/rs/zerolog) - Structured logging
- [Anthropic Claude](https://www.anthropic.com) - AI enrichment

## Support

For issues and questions:
- GitHub Issues: https://github.com/kyivinua/docgen-tool/issues
- Documentation: https://github.com/kyivinua/docgen-tool/tree/main/docs

---

**Built with ❤️ for better documentation**
