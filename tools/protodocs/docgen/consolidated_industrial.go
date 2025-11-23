package docgen

import (
	"fmt"
	"strings"
)

// IndustrialConfig holds industrial-grade documentation configuration
type IndustrialConfig struct {
	// Authentication
	EnableAuthenticationSection bool
	AuthMethods                 []AuthMethod

	// Rate Limiting
	EnableRateLimitingSection bool
	RateLimits                []RateLimitTier

	// SLA/SLO
	EnableSLASection bool
	SLATiers         []SLATier

	// Enhanced Error Handling
	EnableEnhancedErrors bool

	// Versioning & Deprecation
	EnableVersioningSection bool
	VersioningPolicy        VersioningPolicy
}

// AuthMethod represents an authentication method
type AuthMethod struct {
	Name        string
	Description string
	Example     string
}

// RateLimitTier represents a rate limiting tier
type RateLimitTier struct {
	Name           string
	RequestsPerSec int
	RequestsPerDay int
	Burst          int
}

// SLATier represents an SLA tier
type SLATier struct {
	Name              string
	UptimePercentage  string
	MonthlyDowntime   string
	LatencyP95        string
}

// VersioningPolicy represents the versioning policy
type VersioningPolicy struct {
	Strategy       string
	SupportPeriod  string
	DeprecationNotice string
}

// DefaultIndustrialConfig returns default industrial configuration
func DefaultIndustrialConfig() IndustrialConfig {
	return IndustrialConfig{
		EnableAuthenticationSection: true,
		AuthMethods: []AuthMethod{
			{
				Name:        "API Keys",
				Description: "For service-to-service communication",
				Example:     "grpcurl -H 'x-api-key: YOUR_API_KEY' api.example.com:443",
			},
			{
				Name:        "OAuth 2.0",
				Description: "For user-delegated access with bearer tokens",
				Example:     "metadata.add('authorization', 'Bearer ' + accessToken)",
			},
			{
				Name:        "JWT Tokens",
				Description: "For stateless authentication",
				Example:     "metadata.add('authorization', 'Bearer ' + jwtToken)",
			},
		},
		EnableRateLimitingSection: true,
		RateLimits: []RateLimitTier{
			{Name: "Free", RequestsPerSec: 10, RequestsPerDay: 10000, Burst: 20},
			{Name: "Professional", RequestsPerSec: 100, RequestsPerDay: 1000000, Burst: 200},
			{Name: "Enterprise", RequestsPerSec: 1000, RequestsPerDay: 0, Burst: 2000},
		},
		EnableSLASection: true,
		SLATiers: []SLATier{
			{Name: "Standard", UptimePercentage: "99.9%", MonthlyDowntime: "43.8 minutes", LatencyP95: "< 200ms"},
			{Name: "Premium", UptimePercentage: "99.95%", MonthlyDowntime: "21.9 minutes", LatencyP95: "< 100ms"},
			{Name: "Enterprise", UptimePercentage: "99.99%", MonthlyDowntime: "4.38 minutes", LatencyP95: "< 50ms"},
		},
		EnableEnhancedErrors: true,
		EnableVersioningSection: true,
		VersioningPolicy: VersioningPolicy{
			Strategy:          "Semantic Versioning (package.v{major})",
			SupportPeriod:     "12 months for previous version",
			DeprecationNotice: "6 months advance notice",
		},
	}
}

// writeAuthenticationSection writes the authentication and authorization section
func (g *ConsolidatedDocGenerator) writeAuthenticationSection(sb *strings.Builder, doc *ServiceDocumentation, config IndustrialConfig) {
	if !config.EnableAuthenticationSection {
		return
	}

	emoji := ""
	if g.config.UseEmojis {
		emoji = "🔐 "
	}

	sb.WriteString(fmt.Sprintf("## %sAuthentication & Authorization\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"authentication\"></a>\n\n")
	}

	// Supported Authentication Methods
	sb.WriteString("### Supported Authentication Methods\n\n")
	sb.WriteString("This service supports the following authentication methods:\n\n")

	for i, method := range config.AuthMethods {
		sb.WriteString(fmt.Sprintf("%d. **%s** - %s\n", i+1, method.Name, method.Description))
	}
	sb.WriteString("\n")

	// Authentication Examples
	sb.WriteString("### Authentication Examples\n\n")

	// API Key Example
	sb.WriteString("#### Using API Keys\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Command-line (grpcurl)\n")
	sb.WriteString("grpcurl -H 'x-api-key: YOUR_API_KEY' \\\n")
	sb.WriteString("  api.example.com:443 \\\n")
	sb.WriteString(fmt.Sprintf("  %s.%s/%s\n", doc.Service.Package, doc.Service.Name, doc.Methods[0].Name))
	sb.WriteString("```\n\n")

	// OAuth 2.0 Example (Go)
	sb.WriteString("#### Using OAuth 2.0 Bearer Token (Go)\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"context\"\n")
	sb.WriteString("    \"google.golang.org/grpc\"\n")
	sb.WriteString("    \"google.golang.org/grpc/metadata\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("func callWithAuth(client pb.ServiceClient, token string) error {\n")
	sb.WriteString("    // Create metadata with authorization header\n")
	sb.WriteString("    md := metadata.New(map[string]string{\n")
	sb.WriteString("        \"authorization\": \"Bearer \" + token,\n")
	sb.WriteString("    })\n")
	sb.WriteString("    ctx := metadata.NewOutgoingContext(context.Background(), md)\n\n")
	sb.WriteString("    // Make RPC call with authenticated context\n")
	sb.WriteString("    resp, err := client.SomeMethod(ctx, &pb.Request{})\n")
	sb.WriteString("    return err\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// TypeScript Example
	sb.WriteString("#### Using OAuth 2.0 Bearer Token (TypeScript)\n\n")
	sb.WriteString("```typescript\n")
	sb.WriteString("import * as grpc from '@grpc/grpc-js';\n\n")
	sb.WriteString("// Create metadata with authorization header\n")
	sb.WriteString("const metadata = new grpc.Metadata();\n")
	sb.WriteString("metadata.add('authorization', 'Bearer ' + accessToken);\n\n")
	sb.WriteString("// Make RPC call with authenticated metadata\n")
	sb.WriteString("client.someMethod(request, metadata, (error, response) => {\n")
	sb.WriteString("    if (error) {\n")
	sb.WriteString("        console.error('Error:', error);\n")
	sb.WriteString("        return;\n")
	sb.WriteString("    }\n")
	sb.WriteString("    console.log('Response:', response);\n")
	sb.WriteString("});\n")
	sb.WriteString("```\n\n")

	// Authorization Scopes (if methods exist)
	if len(doc.Methods) > 0 {
		sb.WriteString("### Authorization Scopes\n\n")
		sb.WriteString("Different methods require different permission scopes:\n\n")
		sb.WriteString("| Method | Required Scope | Description |\n")
		sb.WriteString("|--------|---------------|-------------|\n")

		for _, method := range doc.Methods {
			scope := g.inferScopeFromMethod(method, doc.Service)
			scopeDesc := g.inferScopeDescription(method)
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", method.Name, scope, scopeDesc))
		}
		sb.WriteString("\n")
	}

	// Security Best Practices
	sb.WriteString("### Security Best Practices\n\n")
	sb.WriteString("- ✅ Always use TLS 1.3+ in production environments\n")
	sb.WriteString("- ✅ Rotate API keys every 90 days\n")
	sb.WriteString("- ✅ Use short-lived tokens (1 hour maximum)\n")
	sb.WriteString("- ✅ Implement request signing for sensitive operations\n")
	sb.WriteString("- ✅ Store credentials securely (use secret management systems)\n")
	sb.WriteString("- ❌ Never log authentication credentials or tokens\n")
	sb.WriteString("- ❌ Never commit API keys to version control\n")
	sb.WriteString("\n---\n\n")
}

// writeRateLimitingSection writes the rate limiting and quotas section
func (g *ConsolidatedDocGenerator) writeRateLimitingSection(sb *strings.Builder, doc *ServiceDocumentation, config IndustrialConfig) {
	if !config.EnableRateLimitingSection {
		return
	}

	emoji := ""
	if g.config.UseEmojis {
		emoji = "⏱️ "
	}

	sb.WriteString(fmt.Sprintf("## %sRate Limits & Quotas\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"rate-limits\"></a>\n\n")
	}

	// Standard Rate Limits Table
	sb.WriteString("### Standard Rate Limits\n\n")
	sb.WriteString("The following rate limits apply to all API requests:\n\n")
	sb.WriteString("| Tier | Requests/Second | Requests/Day | Burst |\n")
	sb.WriteString("|------|----------------|--------------|-------|\n")

	for _, tier := range config.RateLimits {
		requestsPerDay := fmt.Sprintf("%d", tier.RequestsPerDay)
		if tier.RequestsPerDay == 0 {
			requestsPerDay = "Unlimited"
		}
		sb.WriteString(fmt.Sprintf("| **%s** | %d | %s | %d |\n",
			tier.Name, tier.RequestsPerSec, requestsPerDay, tier.Burst))
	}
	sb.WriteString("\n")

	// Rate Limit Headers
	sb.WriteString("### Rate Limit Headers\n\n")
	sb.WriteString("All API responses include rate limit information in the response metadata:\n\n")
	sb.WriteString("```http\n")
	sb.WriteString("x-ratelimit-limit: 100\n")
	sb.WriteString("x-ratelimit-remaining: 87\n")
	sb.WriteString("x-ratelimit-reset: 1634567890\n")
	sb.WriteString("x-ratelimit-retry-after: 42\n")
	sb.WriteString("```\n\n")

	// Handling Rate Limits
	sb.WriteString("### Handling Rate Limits\n\n")
	sb.WriteString("When you exceed rate limits, you'll receive a `RESOURCE_EXHAUSTED` error. ")
	sb.WriteString("Implement exponential backoff to handle rate limiting gracefully:\n\n")

	// TypeScript Example
	sb.WriteString("#### Exponential Backoff Example (TypeScript)\n\n")
	sb.WriteString("```typescript\n")
	sb.WriteString("async function callWithRetry(\n")
	sb.WriteString("  fn: () => Promise<any>,\n")
	sb.WriteString("  maxRetries = 3\n")
	sb.WriteString("): Promise<any> {\n")
	sb.WriteString("  for (let i = 0; i < maxRetries; i++) {\n")
	sb.WriteString("    try {\n")
	sb.WriteString("      return await fn();\n")
	sb.WriteString("    } catch (error: any) {\n")
	sb.WriteString("      if (error.code === grpc.status.RESOURCE_EXHAUSTED) {\n")
	sb.WriteString("        const delay = Math.min(1000 * Math.pow(2, i), 30000);\n")
	sb.WriteString("        console.log(`Rate limited. Retrying in ${delay}ms...`);\n")
	sb.WriteString("        await new Promise(resolve => setTimeout(resolve, delay));\n")
	sb.WriteString("        continue;\n")
	sb.WriteString("      }\n")
	sb.WriteString("      throw error;\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("  throw new Error('Max retries exceeded');\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// Go Example
	sb.WriteString("#### Exponential Backoff Example (Go)\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"context\"\n")
	sb.WriteString("    \"time\"\n")
	sb.WriteString("    \"google.golang.org/grpc/codes\"\n")
	sb.WriteString("    \"google.golang.org/grpc/status\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("func callWithRetry(ctx context.Context, fn func() error, maxRetries int) error {\n")
	sb.WriteString("    for i := 0; i < maxRetries; i++ {\n")
	sb.WriteString("        err := fn()\n")
	sb.WriteString("        if err == nil {\n")
	sb.WriteString("            return nil\n")
	sb.WriteString("        }\n\n")
	sb.WriteString("        if status.Code(err) == codes.ResourceExhausted {\n")
	sb.WriteString("            delay := time.Duration(1000*math.Pow(2, float64(i))) * time.Millisecond\n")
	sb.WriteString("            if delay > 30*time.Second {\n")
	sb.WriteString("                delay = 30 * time.Second\n")
	sb.WriteString("            }\n")
	sb.WriteString("            log.Printf(\"Rate limited. Retrying in %v...\", delay)\n")
	sb.WriteString("            time.Sleep(delay)\n")
	sb.WriteString("            continue\n")
	sb.WriteString("        }\n")
	sb.WriteString("        return err\n")
	sb.WriteString("    }\n")
	sb.WriteString("    return fmt.Errorf(\"max retries exceeded\")\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("---\n\n")
}

// writeSLASection writes the SLA/SLO section
func (g *ConsolidatedDocGenerator) writeSLASection(sb *strings.Builder, doc *ServiceDocumentation, config IndustrialConfig) {
	if !config.EnableSLASection {
		return
	}

	emoji := ""
	if g.config.UseEmojis {
		emoji = "📊 "
	}

	sb.WriteString(fmt.Sprintf("## %sService Level Agreement (SLA)\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"sla\"></a>\n\n")
	}

	// Availability Commitments
	sb.WriteString("### Availability Commitments\n\n")
	sb.WriteString("| Service Tier | Uptime SLA | Monthly Downtime | Latency (p95) |\n")
	sb.WriteString("|-------------|-----------|------------------|---------------|\n")

	for _, tier := range config.SLATiers {
		sb.WriteString(fmt.Sprintf("| **%s** | %s | %s | %s |\n",
			tier.Name, tier.UptimePercentage, tier.MonthlyDowntime, tier.LatencyP95))
	}
	sb.WriteString("\n")

	// Performance Targets
	sb.WriteString("### Performance SLOs\n\n")
	sb.WriteString("#### Latency Targets (p95)\n\n")
	sb.WriteString("| Operation Type | Target | Notes |\n")
	sb.WriteString("|---------------|--------|-------|\n")
	sb.WriteString("| **Read Operations** (Get*) | < 100ms | Measured server-side |\n")
	sb.WriteString("| **Write Operations** (Create*, Update*) | < 500ms | Includes validation |\n")
	sb.WriteString("| **List Operations** (List*) | < 200ms | With pagination |\n")
	sb.WriteString("| **Delete Operations** (Delete*) | < 300ms | Soft delete |\n")
	sb.WriteString("| **Streaming** | < 50ms | Time to first message |\n")
	sb.WriteString("\n")

	// Monitoring
	sb.WriteString("### Monitoring & Observability\n\n")
	sb.WriteString("All services expose Prometheus-compatible metrics:\n\n")
	sb.WriteString("```prometheus\n")
	sb.WriteString("# Request latency histogram (seconds)\n")
	sb.WriteString(fmt.Sprintf("api_request_duration_seconds{service=\"%s\",method=\"%s\",quantile=\"0.95\"}\n\n",
		doc.Service.Name, doc.Methods[0].Name))
	sb.WriteString("# Total request count\n")
	sb.WriteString(fmt.Sprintf("api_requests_total{service=\"%s\",method=\"%s\",status=\"success\"}\n\n",
		doc.Service.Name, doc.Methods[0].Name))
	sb.WriteString("# Error count by code\n")
	sb.WriteString(fmt.Sprintf("api_errors_total{service=\"%s\",method=\"%s\",code=\"INVALID_ARGUMENT\"}\n",
		doc.Service.Name, doc.Methods[0].Name))
	sb.WriteString("```\n\n")

	// Health Check
	sb.WriteString("### Health Check\n\n")
	sb.WriteString("Health status is available via the standard gRPC health check protocol:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("grpc_health_probe -addr=api.example.com:443 \\\n")
	sb.WriteString(fmt.Sprintf("  -service=%s.%s\n", doc.Service.Package, doc.Service.Name))
	sb.WriteString("```\n\n")

	sb.WriteString("---\n\n")
}

// writeVersioningSection writes the versioning and deprecation policy section
func (g *ConsolidatedDocGenerator) writeVersioningSection(sb *strings.Builder, doc *ServiceDocumentation, config IndustrialConfig) {
	if !config.EnableVersioningSection {
		return
	}

	emoji := ""
	if g.config.UseEmojis {
		emoji = "🔄 "
	}

	sb.WriteString(fmt.Sprintf("## %sAPI Versioning & Lifecycle\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"versioning\"></a>\n\n")
	}

	// Versioning Strategy
	sb.WriteString("### Versioning Strategy\n\n")
	sb.WriteString(fmt.Sprintf("**Strategy**: %s\n\n", config.VersioningPolicy.Strategy))
	sb.WriteString("This service follows semantic versioning:\n\n")
	sb.WriteString("- **Major version** (v1, v2): Breaking changes requiring client updates\n")
	sb.WriteString("- **Minor version** (implicit): Backward-compatible additions\n")
	sb.WriteString("- **Patch version**: Bug fixes (not reflected in package name)\n\n")

	sb.WriteString(fmt.Sprintf("**Current Version**: %s\n\n", doc.Service.Version))

	// Version Support Policy
	sb.WriteString("### Version Support Policy\n\n")
	sb.WriteString("| Version State | Support Period | Updates | Deprecation Notice |\n")
	sb.WriteString("|--------------|----------------|---------|-------------------|\n")
	sb.WriteString("| **Current** | Indefinite | Features + Fixes | N/A |\n")
	sb.WriteString(fmt.Sprintf("| **Previous** | %s | Security fixes only | %s prior |\n",
		config.VersioningPolicy.SupportPeriod, config.VersioningPolicy.DeprecationNotice))
	sb.WriteString("| **Deprecated** | 6 months | Critical security only | 12 months prior |\n")
	sb.WriteString("| **Sunset** | 0 months | None | Service disabled |\n")
	sb.WriteString("\n")

	// Breaking vs Non-Breaking Changes
	sb.WriteString("### Breaking vs. Non-Breaking Changes\n\n")
	sb.WriteString("**Breaking changes** (require major version bump):\n")
	sb.WriteString("- ❌ Removing or renaming services, methods, or fields\n")
	sb.WriteString("- ❌ Changing field types or field numbers\n")
	sb.WriteString("- ❌ Changing method behavior significantly\n")
	sb.WriteString("- ❌ Removing enum values\n\n")

	sb.WriteString("**Non-breaking changes** (allowed in current version):\n")
	sb.WriteString("- ✅ Adding new services, methods, or fields\n")
	sb.WriteString("- ✅ Adding optional fields\n")
	sb.WriteString("- ✅ Adding enum values\n")
	sb.WriteString("- ✅ Deprecating (but not removing) fields\n\n")

	// Migration Support
	sb.WriteString("### Migration Support\n\n")
	sb.WriteString("When major version updates are released, we provide:\n\n")
	sb.WriteString("- ✅ Comprehensive migration guides\n")
	sb.WriteString("- ✅ Code examples showing before/after\n")
	sb.WriteString("- ✅ Side-by-side running period (overlap)\n")
	sb.WriteString("- ✅ Automated migration tools (where possible)\n")
	sb.WriteString("- ✅ Dedicated support during migration period\n\n")

	sb.WriteString("---\n\n")
}

// Helper functions

func (g *ConsolidatedDocGenerator) inferScopeFromMethod(method MethodDoc, service *ServiceDoc) string {
	name := strings.ToLower(method.Name)
	serviceName := strings.ToLower(strings.TrimSuffix(service.Name, "Service"))

	if strings.HasPrefix(name, "create") || strings.HasPrefix(name, "update") {
		return fmt.Sprintf("%s.write", serviceName)
	} else if strings.HasPrefix(name, "delete") {
		return fmt.Sprintf("%s.admin", serviceName)
	} else if strings.HasPrefix(name, "get") || strings.HasPrefix(name, "list") {
		return fmt.Sprintf("%s.read", serviceName)
	}
	return fmt.Sprintf("%s.access", serviceName)
}

func (g *ConsolidatedDocGenerator) inferScopeDescription(method MethodDoc) string {
	name := strings.ToLower(method.Name)

	if strings.HasPrefix(name, "create") {
		return "Create new resources"
	} else if strings.HasPrefix(name, "update") {
		return "Modify existing resources"
	} else if strings.HasPrefix(name, "delete") {
		return "Administrative access required"
	} else if strings.HasPrefix(name, "get") {
		return "Read individual resources"
	} else if strings.HasPrefix(name, "list") {
		return "List and query resources"
	}
	return "Access resources"
}
