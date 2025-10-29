# Shared Resources & Template Guide

This guide explains how docgen-tool handles shared proto files, resolves dependencies, and uses templates for documentation generation.

## Table of Contents

1. [Shared Resources Overview](#shared-resources-overview)
2. [Configuration](#configuration)
3. [Features](#features)
4. [Use Cases](#use-cases)
5. [Templates](#templates)
6. [Best Practices](#best-practices)

## Shared Resources Overview

Modern microservice architectures often share common proto definitions across services. Docgen-tool automatically discovers and resolves these shared resources, ensuring comprehensive and consistent documentation.

### Problems Solved

- **Shared Message Definitions** - Common messages used across multiple services
- **Import Dependencies** - Resolving proto import statements
- **Services Without Proto Files** - Generating documentation for services lacking proto definitions
- **Duplicate Messages** - Deduplicating messages that appear in multiple services
- **Virtual Services** - Creating documentation for shared proto files without service definitions

## Configuration

### Basic Configuration

```yaml
shared_resources:
  enabled: true
  auto_discover: true
  deduplicate_messages: true
```

### Complete Configuration

```yaml
shared_resources:
  # Enable shared resource resolution
  enabled: true

  # Paths to shared proto files
  shared_proto_paths:
    - "./proto/common"
    - "./proto/shared"
    - "./api/common"

  # Common directory patterns to auto-discover
  common_dir_patterns:
    - "common"
    - "shared"
    - "proto/common"
    - "proto/shared"
    - "api/common"
    - "api/shared"

  # Auto-discover shared proto files
  auto_discover: true

  # Import resolution strategy: local, global, hybrid
  import_strategy: "hybrid"

  # Generate virtual services for shared protos without services
  generate_virtual_services: false

  # Deduplicate shared messages across services
  deduplicate_messages: true
```

## Features

### 1. Auto-Discovery

Automatically discovers shared proto files in common directories.

**How It Works:**
1. Scans configured `shared_proto_paths`
2. Looks for common directory patterns (`common/`, `shared/`, etc.)
3. Discovers proto files referenced in imports
4. Builds a complete dependency graph

**Example:**
```
proto/
├── user/
│   └── v1/
│       └── service.proto    # imports "common/types.proto"
├── product/
│   └── v1/
│       └── service.proto    # imports "common/types.proto"
└── common/
    └── types.proto          # Automatically discovered
```

### 2. Import Resolution

Resolves proto import statements and loads dependencies.

**Strategies:**
- **local**: Resolve imports relative to proto file location
- **global**: Resolve imports from configured proto paths
- **hybrid** (default): Try local first, then global

**Example:**
```protobuf
// user/v1/service.proto
import "common/types.proto";
import "google/protobuf/timestamp.proto";

// Both imports are automatically resolved and loaded
```

### 3. Message Deduplication

Removes duplicate message definitions that appear across services.

**Before:**
```
UserService:
  - User
  - CommonRequest
  - CommonResponse

ProductService:
  - Product
  - CommonRequest    # Duplicate
  - CommonResponse   # Duplicate
```

**After:**
```
UserService:
  - User
  - CommonRequest
  - CommonResponse

ProductService:
  - Product
  # Duplicates removed
```

### 4. Virtual Services

Generates documentation for shared proto files that don't define services.

**Use Case:**
```
proto/common/types.proto defines messages but no services
↓
Virtual service "Shared_common" is created
↓
Documentation generated for common types
```

**Configuration:**
```yaml
shared_resources:
  generate_virtual_services: true
```

### 5. Services Without Proto Files

Handles services that don't have proto files by generating minimal definitions.

**Example:**
```go
// Service name provided but no proto files
service, err := resolver.HandleServiceWithoutProtos("MyService")

// Generates:
// - Minimal service definition
// - Basic CRUD methods
// - Request/Response messages
```

## Use Cases

### Use Case 1: Microservices with Shared Types

**Scenario:**
Multiple services share common message types (User, Address, Timestamp, etc.)

**Solution:**
```yaml
shared_resources:
  enabled: true
  shared_proto_paths:
    - "./proto/common"
  auto_discover: true
  deduplicate_messages: true
```

**Result:**
- Common types automatically discovered
- Included in relevant service documentation
- No duplicates in generated docs

### Use Case 2: Legacy Services Without Proto Files

**Scenario:**
Service exists but proto files are missing or not accessible

**Solution:**
```yaml
services:
  - name: "LegacyService"
    proto_files: []  # No proto files
    enabled: true

shared_resources:
  enabled: true
```

**Result:**
- Minimal service definition generated
- Basic methods created (Get, List, Create, Update, Delete)
- Can be enhanced with AI enrichment

### Use Case 3: Documentation for Shared Libraries

**Scenario:**
Proto library with common types but no services

**Solution:**
```yaml
shared_resources:
  enabled: true
  generate_virtual_services: true
  shared_proto_paths:
    - "./proto-lib/common"
```

**Result:**
- Virtual service "Shared_common" created
- All common types documented
- Searchable and navigable documentation

### Use Case 4: Complex Import Dependencies

**Scenario:**
Services with nested import dependencies

**Solution:**
```yaml
shared_resources:
  enabled: true
  auto_discover: true
  import_strategy: "hybrid"
```

**Result:**
- All imports resolved automatically
- Dependency graph built
- Complete documentation generated

## Templates

Docgen-tool uses Go templates for flexible documentation generation.

### Built-in Templates

Located in `templates/` directory:

- **service.md.tmpl** - Service documentation template
- **method.md.tmpl** - Method documentation template
- **message.md.tmpl** - Message documentation template

### Template Functions

Available functions in templates:

- `title` - Title case string
- `lower` - Lowercase string
- `upper` - Uppercase string
- `replace` - Replace all occurrences
- `contains` - Check if string contains substring
- `hasPrefix` - Check if string starts with prefix
- `hasSuffix` - Check if string ends with suffix
- `trim` - Trim whitespace
- `goType` - Convert proto type to Go type
- `pythonType` - Convert proto type to Python type
- `jsonType` - Convert proto type to JSON type
- `join` - Join strings with separator

### Custom Templates

#### 1. Create Custom Template

```markdown
# {{.Name}} Service

Package: {{.Package}}

## Methods

{{range .Methods}}
### {{.Name}}

{{.Description}}

**Request:** `{{.InputType}}`
**Response:** `{{.OutputType}}`
{{end}}
```

#### 2. Configure Template Path

```yaml
templates:
  dir: "./my-templates"
  service_template: "my-service.md.tmpl"
```

#### 3. Use Custom Template Functions

```go
import "github.com/kyivinua/docgen-tool/internal/generator"

funcMap := generator.GetTemplateFunctions()
// Add custom functions
funcMap["myFunc"] = func(s string) string {
    return "custom: " + s
}
```

### Template Examples

#### Example 1: Minimal Service Template

```markdown
# {{.Name}}

{{.Description}}

Methods: {{len .Methods}}
Messages: {{len .Messages}}
```

#### Example 2: Detailed Method Template

```markdown
## {{.Name}}

{{.Description}}

### Request

Type: `{{.InputType | goType}}`

{{if .ClientStreaming}}
This is a client streaming method.
{{end}}

### Response

Type: `{{.OutputType | goType}}`

{{if .ServerStreaming}}
This is a server streaming method.
{{end}}

### Example

```go
req := &pb.{{.InputType}}{}
resp, err := client.{{.Name}}(ctx, req)
```
```

#### Example 3: Message with Type Conversions

```markdown
# {{.Name}}

## Fields

{{range .Fields}}
- **{{.Name}}**
  - Proto: `{{.Type}}`
  - Go: `{{.Type | goType}}`
  - Python: `{{.Type | pythonType}}`
  - JSON: `{{.Type | jsonType}}`
{{end}}
```

## Best Practices

### 1. Organize Proto Files

```
proto/
├── common/           # Shared types
│   ├── types.proto
│   └── errors.proto
├── user/
│   └── v1/
│       └── service.proto
└── product/
    └── v1/
        └── service.proto
```

### 2. Use Consistent Import Paths

```protobuf
// Good: Consistent relative imports
import "common/types.proto";
import "common/errors.proto";

// Avoid: Inconsistent paths
import "../common/types.proto";
import "proto/common/errors.proto";
```

### 3. Configure Shared Paths

```yaml
shared_resources:
  shared_proto_paths:
    - "./proto/common"
    - "./proto/shared"
    - "./api/v1/common"
```

### 4. Enable Auto-Discovery

```yaml
shared_resources:
  auto_discover: true  # Recommended
```

### 5. Use Message Deduplication

```yaml
shared_resources:
  deduplicate_messages: true  # Prevents duplicate documentation
```

### 6. Handle Missing Proto Files

```yaml
# Allow services without proto files
services:
  - name: "LegacyService"
    proto_files: []
    enabled: true

shared_resources:
  enabled: true
```

### 7. Custom Templates for Consistency

- Create organization-specific templates
- Include branding and style guidelines
- Add custom metadata and badges

## Troubleshooting

### Issue: Shared Messages Not Found

**Problem:** Shared messages aren't included in documentation

**Solution:**
```yaml
shared_resources:
  enabled: true
  auto_discover: true
  shared_proto_paths:
    - "./proto/common"  # Add explicit paths
```

### Issue: Import Resolution Fails

**Problem:** Proto imports can't be resolved

**Solution:**
```yaml
shared_resources:
  import_strategy: "hybrid"  # Try local then global
```

### Issue: Duplicate Messages

**Problem:** Same message appears multiple times

**Solution:**
```yaml
shared_resources:
  deduplicate_messages: true
```

### Issue: Service Without Proto Files

**Problem:** Can't generate docs for service without protos

**Solution:**
```yaml
shared_resources:
  enabled: true  # Enables minimal service generation
```

### Issue: Template Errors

**Problem:** Template rendering fails

**Solution:**
1. Check template syntax
2. Verify field names match data structure
3. Use default templates as reference

## Summary

Shared resource resolution in docgen-tool provides:

✅ **Auto-Discovery** - Finds shared proto files automatically
✅ **Import Resolution** - Resolves all dependencies
✅ **Deduplication** - Removes duplicate messages
✅ **Virtual Services** - Documents shared proto libraries
✅ **Flexible Templates** - Customizable documentation format
✅ **Missing Proto Handling** - Generates docs for services without protos

Configure once, document everything!
