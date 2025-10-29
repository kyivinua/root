# Docgen-Tool Quick Start Guide

Get started with docgen-tool in 5 minutes! This guide will walk you through installation, configuration, and generating your first documentation.

## Prerequisites

- Go 1.22 or later
- Protocol Buffer files (.proto)
- (Optional) Anthropic API key for AI enrichment

## Step 1: Installation

### Option A: From Source

```bash
# Clone the repository
git clone https://github.com/kyivinua/docgen-tool.git
cd docgen-tool

# Build and install
make build
make install
```

### Option B: Using Installation Script

```bash
curl -sSL https://raw.githubusercontent.com/kyivinua/docgen-tool/main/scripts/install.sh | bash
```

### Verify Installation

```bash
docgen version
# Output: docgen-tool version 1.0.0
```

## Step 2: Initialize Configuration

Create a default configuration file:

```bash
docgen init --output ./docgen.yaml
```

This creates a `docgen.yaml` file with default settings.

## Step 3: Configure Your Project

Edit `docgen.yaml` to match your project:

```yaml
version: "1.0"
project_name: "my-awesome-project"
output_dir: "./docs/generated"

services:
  - name: "UserService"
    proto_files:
      - "./proto/user/v1/service.proto"
    enabled: true
    package: "user.v1"
    version: "1.0"

quality:
  coverage_target: 80
  description_quality_target: 85
  auto_fix: true

enricher:
  provider: "claude"
  api_key: "${ANTHROPIC_API_KEY}"
  enabled: false  # Set to true to enable AI

diagrams:
  enabled: true
  service_graph: true
  message_graph: true
```

## Step 4: Generate Documentation

### Basic Generation

```bash
docgen generate --config ./docgen.yaml
```

### With AI Enrichment (Recommended)

```bash
# Set your Anthropic API key
export ANTHROPIC_API_KEY=sk-ant-...

# Generate with AI
docgen generate --config ./docgen.yaml --ai
```

### With Auto-Fix

```bash
# Automatically fix quality issues
docgen generate --config ./docgen.yaml --auto-fix
```

## Step 5: View Your Documentation

Your documentation is generated in the `output_dir` specified in your config:

```bash
cd ./docs/generated
ls -la
# README.md
# userservice.md
# diagrams/
```

Open `README.md` to see the overview and quality metrics!

## Common Use Cases

### 1. Multiple Services

```yaml
services:
  - name: "UserService"
    proto_files:
      - "./proto/user/v1/service.proto"
    enabled: true

  - name: "ProductService"
    proto_files:
      - "./proto/product/v1/service.proto"
    enabled: true

  - name: "OrderService"
    proto_files:
      - "./proto/order/v1/service.proto"
    enabled: true
```

### 2. High-Quality Documentation

```yaml
quality:
  coverage_target: 95
  description_quality_target: 90
  auto_fix: true
  fail_on_quality_gate: true
```

### 3. Custom Output Structure

```yaml
output_dir: "./docs/api"

templates:
  dir: "./templates/custom"
  service_template: "my-service.md.tmpl"
```

## Validation

Validate your configuration without generating:

```bash
docgen validate --config ./docgen.yaml
```

## Tips for Best Results

1. **Enable AI Enrichment** - Provides the best quality descriptions
2. **Use Auto-Fix** - Automatically fills in missing descriptions
3. **Set Realistic Quality Targets** - Start with 80% and increase gradually
4. **Enable All Diagrams** - Visual documentation is easier to understand
5. **Run Regularly** - Keep documentation in sync with code

## Example Output

After generation, you'll see:

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

## Troubleshooting

### Missing Proto Files

```bash
Error: failed to read proto file: no such file or directory
```

**Solution:** Ensure proto file paths in config are correct and relative to where you run docgen.

### Low Quality Score

```
⚠️  Quality Issues: 15
  - [warning] Method coverage 60% is below target 80%
```

**Solution:** Enable auto-fix or add descriptions manually to your proto files.

### API Key Issues

```bash
Error: enricher.api_key is required when enricher is enabled
```

**Solution:** Set your API key in environment variable:
```bash
export ANTHROPIC_API_KEY=your-key-here
```

## Next Steps

- Read the [Architecture Documentation](ARCHITECTURE.md) to understand how docgen-tool works
- Customize templates for your documentation style
- Set up CI/CD integration for automatic documentation updates
- Explore advanced configuration options

## Getting Help

- GitHub Issues: https://github.com/kyivinua/docgen-tool/issues
- Documentation: https://github.com/kyivinua/docgen-tool/tree/main/docs

Happy documenting!
