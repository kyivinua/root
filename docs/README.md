# Docgen-Tool Documentation

Welcome to the docgen-tool documentation! This directory contains comprehensive guides and references for using docgen-tool.

## Documentation Index

- **[QUICKSTART.md](QUICKSTART.md)** - Get started in 5 minutes
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed architecture and design documentation

## What is docgen-tool?

Docgen-tool is an enterprise-grade documentation generator for Protocol Buffer services. It automatically generates comprehensive, high-quality documentation with AI-powered enrichment, quality validation, and diagram generation.

## Key Features

- **AI-Powered Enrichment** - Automatic description enhancement using Claude AI
- **Quality Validation** - Comprehensive coverage and quality metrics
- **Diagram Generation** - Mermaid diagrams for services, messages, and sequences
- **Auto-Fix** - Automatic quality issue resolution
- **Parallel Processing** - Fast generation for large projects
- **Security-First** - Built with security best practices

## Quick Links

- [Installation](#installation)
- [Quick Start](QUICKSTART.md)
- [Configuration](#configuration)
- [Examples](#examples)

## Installation

```bash
# Clone the repository
git clone https://github.com/kyivinua/docgen-tool.git
cd docgen-tool

# Build and install
make build
make install
```

## Configuration

Create a configuration file:

```bash
docgen init --output ./docgen.yaml
```

Edit the configuration to match your project structure.

## Examples

### Basic Usage

```bash
# Generate documentation
docgen generate --config ./docgen.yaml
```

### With AI Enrichment

```bash
# Set API key
export ANTHROPIC_API_KEY=your-api-key

# Generate with AI
docgen generate --config ./docgen.yaml --ai
```

### With Auto-Fix

```bash
# Generate with automatic quality fixes
docgen generate --config ./docgen.yaml --auto-fix
```

## Support

For issues and questions, please visit our [GitHub repository](https://github.com/kyivinua/docgen-tool).
