# Advanced AI Enrichment Features

Docgen-tool includes sophisticated AI-powered enrichment capabilities that go beyond basic description generation. These advanced features leverage Claude AI to provide comprehensive, high-quality documentation with minimal manual effort.

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Configuration](#configuration)
4. [Usage Examples](#usage-examples)
5. [Strategies](#strategies)
6. [Best Practices](#best-practices)

## Overview

Advanced AI enrichment transforms basic proto definitions into comprehensive, professional documentation through:

- **Context-Aware Generation** - Understands relationships between services, methods, and messages
- **Example Code Generation** - Automatically creates usage examples in multiple languages
- **Terminology Management** - Extracts and maintains consistent terminology
- **Best Practice Analysis** - Suggests API design improvements
- **Multi-Pass Refinement** - Iteratively improves content quality
- **Batch Processing** - Efficient API usage for large projects
- **Quality Scoring** - AI-powered assessment of documentation quality

## Features

### 1. Context-Aware Enrichment

Generates descriptions that understand the full context of your API:

```yaml
enricher:
  advanced:
    enabled: true
    use_context_aware: true
    project_domain: "E-commerce Microservices"
```

**Benefits:**
- Understands service relationships and dependencies
- Generates contextually relevant descriptions
- Maintains consistency across the entire API
- Leverages project domain knowledge

**Example Output:**
```
Before: "UserService"
After:  "UserService provides comprehensive user account management within the
         e-commerce platform, handling authentication, profile management, and
         user preferences. It serves as the central authority for user identity
         and integrates with OrderService and PaymentService."
```

### 2. Example Code Generation

Automatically generates code examples in multiple programming languages:

```yaml
enricher:
  advanced:
    enabled: true
    use_example_generation: true
    include_code_examples: true
    languages_for_examples:
      - "go"
      - "python"
      - "typescript"
      - "java"
```

**Example Output:**

**Go:**
```go
// Create a new user
client := pb.NewUserServiceClient(conn)
req := &pb.CreateUserRequest{
    Name:  "John Doe",
    Email: "john@example.com",
}
resp, err := client.CreateUser(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created user: %s\n", resp.Id)
```

**Python:**
```python
# Create a new user
stub = user_pb2_grpc.UserServiceStub(channel)
request = user_pb2.CreateUserRequest(
    name="John Doe",
    email="john@example.com"
)
response = stub.CreateUser(request)
print(f"Created user: {response.id}")
```

### 3. Terminology Extraction

Automatically extracts and maintains consistent terminology:

```yaml
enricher:
  advanced:
    enabled: true
    use_terminology: true
```

**Benefits:**
- Identifies key technical terms
- Ensures consistent definitions across documentation
- Warns about terminology inconsistencies
- Builds a project glossary automatically

**Example:**
```yaml
Extracted Terminology:
  - "User": An authenticated account with associated profile data
  - "Session": An active authentication state with expiration
  - "Role": A permission level that determines API access
```

### 4. Best Practice Analysis

Analyzes your API design and suggests improvements:

```yaml
enricher:
  advanced:
    enabled: true
    use_best_practices: true
```

**Example Suggestions:**
- **Naming Convention**: "Consider using `list_users` instead of `getUsers` for consistency with gRPC naming conventions"
- **Error Handling**: "Add explicit error codes in responses for better client error handling"
- **Pagination**: "Consider adding pagination to `ListUsers` method for better scalability"
- **Versioning**: "Include API version in package name (e.g., `user.v1`)"

### 5. Multi-Pass Refinement

Iteratively improves content quality through multiple AI passes:

```yaml
enricher:
  advanced:
    enabled: true
    use_multi_pass: true
    multi_pass_count: 3
```

**How it Works:**
1. **First Pass**: Generate initial comprehensive description
2. **Second Pass**: Refine for clarity and technical accuracy
3. **Third Pass**: Polish for professional tone and completeness

**Quality Improvement:**
- Initial: 70% quality score
- After Pass 2: 85% quality score
- After Pass 3: 92% quality score

### 6. Batch Processing

Efficiently processes multiple items in batches:

```yaml
enricher:
  advanced:
    enabled: true
    batch_size: 10
```

**Benefits:**
- Reduced API calls
- Better rate limit utilization
- Faster overall processing
- Lower costs

### 7. Quality Scoring

AI-powered assessment of documentation quality:

```yaml
enricher:
  advanced:
    enabled: true
    enable_quality_scoring: true
    min_quality_score: 80.0
```

**Scoring Criteria:**
- **Clarity** (25%): Easy to understand
- **Completeness** (25%): All aspects covered
- **Technical Accuracy** (25%): Correct technical details
- **Professional Tone** (15%): Appropriate writing style
- **Examples & Details** (10%): Helpful additional information

## Configuration

### Basic Advanced Enrichment

```yaml
enricher:
  provider: "claude"
  api_key: "${ANTHROPIC_API_KEY}"
  enabled: true

  advanced:
    enabled: true
    use_context_aware: true
    use_terminology: true
    detail_level: "standard"
    tone: "technical"
```

### Full Advanced Configuration

```yaml
enricher:
  advanced:
    enabled: true

    # Core Features
    use_context_aware: true
    use_example_generation: true
    use_terminology: true
    use_best_practices: true
    use_multi_pass: true

    # Options
    multi_pass_count: 2
    batch_size: 5
    detail_level: "comprehensive"  # minimal, standard, detailed, comprehensive
    tone: "technical"               # technical, conversational, formal

    # Examples
    include_code_examples: true
    languages_for_examples:
      - "go"
      - "python"
      - "typescript"
      - "java"
      - "rust"

    # Context
    project_domain: "Microservices E-commerce Platform"

    # Quality
    min_quality_score: 80.0
    enable_quality_scoring: true
```

## Usage Examples

### Example 1: E-commerce Platform

```yaml
enricher:
  advanced:
    enabled: true
    use_context_aware: true
    use_example_generation: true
    project_domain: "E-commerce Platform with Order, Payment, and Shipping Services"
    detail_level: "comprehensive"
    include_code_examples: true
```

**Result:**
- Context-aware descriptions mentioning related services
- Code examples showing typical workflows
- Comprehensive documentation with business context

### Example 2: Internal Microservices

```yaml
enricher:
  advanced:
    enabled: true
    use_context_aware: true
    use_terminology: true
    use_best_practices: true
    project_domain: "Internal Microservices for Data Processing Pipeline"
    tone: "technical"
```

**Result:**
- Consistent technical terminology
- Best practice recommendations
- Technical documentation style

### Example 3: Public API

```yaml
enricher:
  advanced:
    enabled: true
    use_context_aware: true
    use_example_generation: true
    use_multi_pass: true
    multi_pass_count: 3
    include_code_examples: true
    languages_for_examples:
      - "python"
      - "javascript"
      - "curl"
    detail_level: "comprehensive"
    tone: "conversational"
    enable_quality_scoring: true
    min_quality_score: 90.0
```

**Result:**
- Developer-friendly, conversational tone
- Examples in popular languages
- Very high quality documentation
- Comprehensive coverage

## Strategies

### Context-Aware Strategy

Best for understanding service relationships and dependencies.

**When to Use:**
- Multi-service architectures
- Services with complex interactions
- Need for consistent cross-service documentation

### Example Generation Strategy

Best for developer-focused documentation.

**When to Use:**
- Public APIs
- SDKs and client libraries
- Developer onboarding materials

### Terminology Strategy

Best for maintaining consistency.

**When to Use:**
- Large projects with multiple contributors
- Domain-specific terminology
- Compliance requirements

### Best Practice Strategy

Best for improving API design.

**When to Use:**
- New API development
- API reviews
- Refactoring projects

### Multi-Pass Strategy

Best for highest quality output.

**When to Use:**
- Public-facing APIs
- Critical documentation
- High standards for quality

## Best Practices

### 1. Start Simple

Begin with basic advanced features and add more as needed:

```yaml
enricher:
  advanced:
    enabled: true
    use_context_aware: true  # Start with this
```

### 2. Use Project Domain

Always set the project domain for better context:

```yaml
enricher:
  advanced:
    project_domain: "Healthcare Data Management Platform"
```

### 3. Choose Appropriate Detail Level

Match detail level to your audience:

- **minimal**: Internal APIs, quick references
- **standard**: Most use cases
- **detailed**: Complex APIs
- **comprehensive**: Public APIs, external documentation

### 4. Enable Quality Scoring for Critical APIs

```yaml
enricher:
  advanced:
    enable_quality_scoring: true
    min_quality_score: 85.0
```

### 5. Use Example Generation Wisely

Generate examples for public APIs but skip for internal services:

```yaml
# For public API
enricher:
  advanced:
    use_example_generation: true
    languages_for_examples: ["python", "javascript", "go"]

# For internal services
enricher:
  advanced:
    use_example_generation: false
```

### 6. Monitor API Costs

Use batch processing and caching to optimize costs:

```yaml
enricher:
  cache_enabled: true
  advanced:
    batch_size: 10  # Larger batches = fewer API calls
```

### 7. Combine Features

Use multiple strategies together for best results:

```yaml
enricher:
  advanced:
    use_context_aware: true      # For context
    use_terminology: true         # For consistency
    use_multi_pass: true          # For quality
    multi_pass_count: 2
```

## Performance Considerations

### API Costs

- Context-Aware: ~500 tokens per item
- Example Generation: ~1000 tokens per example
- Multi-Pass (2 passes): 2x base cost
- Best Practices: ~800 tokens per service

### Processing Time

- Basic Enrichment: ~1s per item
- Context-Aware: ~2s per item
- With Examples: ~3-4s per item
- Multi-Pass (2): 2-3x base time

### Optimization Tips

1. Use caching for repeated content
2. Process in batches
3. Enable only needed features
4. Use appropriate rate limits

## Troubleshooting

### Low Quality Scores

**Problem:** Generated content not meeting quality threshold

**Solutions:**
- Enable multi-pass refinement
- Lower quality threshold temporarily
- Provide better project domain context
- Use more detailed detail level

### Slow Performance

**Problem:** Enrichment taking too long

**Solutions:**
- Increase batch size
- Reduce multi-pass count
- Disable example generation for internal APIs
- Increase parallel requests

### Inconsistent Terminology

**Problem:** Different terms used for same concepts

**Solutions:**
- Enable terminology extraction
- Provide initial terminology in config
- Use multi-pass refinement
- Review and standardize manually

## Summary

Advanced AI enrichment provides powerful capabilities for generating high-quality documentation:

✅ **Context-Aware** - Understands your entire API
✅ **Code Examples** - Automatic generation in multiple languages
✅ **Terminology** - Consistent definitions across docs
✅ **Best Practices** - AI-powered design suggestions
✅ **Multi-Pass** - Iterative quality improvement
✅ **Batch Processing** - Efficient API usage
✅ **Quality Scoring** - Measurable documentation quality

Start with basic features and gradually enable advanced capabilities as your needs grow!
