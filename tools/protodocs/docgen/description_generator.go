package docgen

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/descriptorpb"
)

// DescriptionGenerator generates smart descriptions from context and naming patterns
type DescriptionGenerator struct{}

// NewDescriptionGenerator creates a new description generator
func NewDescriptionGenerator() *DescriptionGenerator {
	return &DescriptionGenerator{}
}

// EnhanceServiceDescription improves service description with context
func (dg *DescriptionGenerator) EnhanceServiceDescription(name, description string) string {
	if description != "" {
		return dg.formatDescription(description)
	}

	// Generate smart default based on service name
	return dg.generateServiceDescription(name)
}

// EnhanceMethodDescription improves method description with context
func (dg *DescriptionGenerator) EnhanceMethodDescription(methodName, description, inputType, outputType string, clientStreaming, serverStreaming bool) string {
	if description != "" {
		desc := dg.formatDescription(description)
		// Add streaming info if not mentioned
		streamingInfo := dg.getStreamingDescription(clientStreaming, serverStreaming)
		if streamingInfo != "" && !strings.Contains(strings.ToLower(desc), "stream") {
			desc = fmt.Sprintf("%s %s", desc, streamingInfo)
		}
		return desc
	}

	// Generate smart default
	return dg.generateMethodDescription(methodName, inputType, outputType, clientStreaming, serverStreaming)
}

// EnhanceFieldDescription improves field description with context
func (dg *DescriptionGenerator) EnhanceFieldDescription(fieldName, description, fieldType, typeName string, isRepeated, isOneof bool) string {
	if description != "" {
		return dg.formatDescription(description)
	}

	// Generate smart default based on field name and type
	return dg.generateFieldDescription(fieldName, fieldType, typeName, isRepeated, isOneof)
}

// EnhanceMessageDescription improves message description
func (dg *DescriptionGenerator) EnhanceMessageDescription(name, description string, isRequest, isResponse bool) string {
	if description != "" {
		return dg.formatDescription(description)
	}

	// Generate smart default
	return dg.generateMessageDescription(name, isRequest, isResponse)
}

// EnhanceEnumDescription improves enum description
func (dg *DescriptionGenerator) EnhanceEnumDescription(name, description string) string {
	if description != "" {
		return dg.formatDescription(description)
	}

	return dg.generateEnumDescription(name)
}

// formatDescription formats and cleans up description text
func (dg *DescriptionGenerator) formatDescription(desc string) string {
	// Remove excessive whitespace
	desc = strings.TrimSpace(desc)

	// Ensure first letter is capitalized
	if len(desc) > 0 {
		desc = strings.ToUpper(string(desc[0])) + desc[1:]
	}

	// Ensure it ends with period if it's a sentence
	if len(desc) > 0 && !strings.HasSuffix(desc, ".") && !strings.HasSuffix(desc, "!") && !strings.HasSuffix(desc, "?") {
		desc = desc + "."
	}

	return desc
}

// generateServiceDescription generates description from service name
func (dg *DescriptionGenerator) generateServiceDescription(serviceName string) string {
	// Remove "Service" suffix for cleaner description
	name := strings.TrimSuffix(serviceName, "Service")

	// Split camelCase
	words := dg.splitCamelCase(name)

	// Common service patterns
	patterns := map[string]string{
		"User":         "Manages user accounts, profiles, and authentication",
		"Payment":      "Handles payment processing, transactions, and billing",
		"Notification": "Manages notifications and messaging across channels",
		"Analytics":    "Provides analytics, reporting, and data insights",
		"Auth":         "Handles authentication and authorization",
		"Order":        "Manages orders and order processing",
		"Product":      "Manages product catalog and inventory",
		"Search":       "Provides search and discovery capabilities",
		"Storage":      "Manages data storage and retrieval",
		"Workflow":     "Orchestrates workflows and business processes",
	}

	for pattern, desc := range patterns {
		if strings.Contains(name, pattern) {
			return desc + "."
		}
	}

	// Generic description
	return fmt.Sprintf("Provides %s management and operations.", strings.ToLower(strings.Join(words, " ")))
}

// generateMethodDescription generates description from method name and signature
func (dg *DescriptionGenerator) generateMethodDescription(methodName, inputType, outputType string, clientStreaming, serverStreaming bool) string {
	// Extract action verb from method name
	verb := dg.extractVerb(methodName)
	entity := dg.extractEntity(methodName)

	var desc string

	// Pattern matching for common operations
	switch verb {
	case "Create":
		desc = fmt.Sprintf("Creates a new %s", entity)
	case "Get":
		desc = fmt.Sprintf("Retrieves %s by identifier", entity)
	case "Update":
		desc = fmt.Sprintf("Updates an existing %s", entity)
	case "Delete":
		desc = fmt.Sprintf("Deletes %s", entity)
	case "List":
		desc = fmt.Sprintf("Lists %ss with optional filtering and pagination", entity)
	case "Search":
		desc = fmt.Sprintf("Searches for %ss based on criteria", entity)
	case "Batch":
		desc = fmt.Sprintf("Performs batch operations on %ss", entity)
	case "Stream":
		desc = fmt.Sprintf("Streams %s updates in real-time", entity)
	case "Subscribe":
		desc = fmt.Sprintf("Subscribes to %s events", entity)
	case "Cancel":
		desc = fmt.Sprintf("Cancels %s operation", entity)
	case "Validate":
		desc = fmt.Sprintf("Validates %s data", entity)
	case "Process":
		desc = fmt.Sprintf("Processes %s", entity)
	case "Calculate":
		desc = fmt.Sprintf("Calculates %s", entity)
	default:
		desc = fmt.Sprintf("Performs %s operation", methodName)
	}

	// Add streaming information
	streamingInfo := dg.getStreamingDescription(clientStreaming, serverStreaming)
	if streamingInfo != "" {
		desc = fmt.Sprintf("%s. %s", desc, streamingInfo)
	}

	return desc + "."
}

// generateFieldDescription generates description from field name and type
func (dg *DescriptionGenerator) generateFieldDescription(fieldName, fieldType, typeName string, isRepeated, isOneof bool) string {
	// Split snake_case or camelCase
	words := dg.splitSnakeCase(fieldName)
	if len(words) == 1 {
		words = dg.splitCamelCase(fieldName)
	}

	// Join words
	fieldDesc := strings.Join(words, " ")
	fieldDesc = cases.Title(language.English).String(fieldDesc)

	// Add type context
	var typeContext string
	if isRepeated {
		typeContext = " (list)"
	} else if isOneof {
		typeContext = " (one of multiple options)"
	}

	// Common field patterns
	patterns := map[string]string{
		"id":           "Unique identifier",
		"uuid":         "Universally unique identifier",
		"name":         "Name",
		"email":        "Email address",
		"phone":        "Phone number",
		"address":      "Physical address",
		"created_at":   "Timestamp when created",
		"updated_at":   "Timestamp when last updated",
		"deleted_at":   "Timestamp when deleted (soft delete)",
		"is_active":    "Indicates if active",
		"is_enabled":   "Indicates if enabled",
		"is_deleted":   "Indicates if deleted",
		"status":       "Current status",
		"type":         "Type classification",
		"metadata":     "Additional metadata",
		"description":  "Detailed description",
		"total":        "Total count or amount",
		"count":        "Number of items",
		"amount":       "Monetary amount",
		"price":        "Price value",
		"quantity":     "Quantity",
		"url":          "URL address",
		"token":        "Authentication or access token",
		"code":         "Code or identifier",
		"version":      "Version number",
		"timestamp":    "Point in time",
		"duration":     "Time duration",
		"timeout":      "Timeout period",
		"limit":        "Maximum limit",
		"offset":       "Starting offset",
		"page":         "Page number",
		"page_size":    "Items per page",
		"sort_by":      "Field to sort by",
		"sort_order":   "Sort direction (asc/desc)",
		"filter":       "Filter criteria",
		"query":        "Search query",
		"data":         "Data payload",
		"payload":      "Request or response payload",
		"error":        "Error information",
		"message":      "Message content",
		"title":        "Title",
		"body":         "Main content body",
		"tags":         "Associated tags",
		"labels":       "Associated labels",
		"settings":     "Configuration settings",
		"config":       "Configuration",
		"options":      "Available options",
		"params":       "Parameters",
	}

	// Check for exact matches
	lowerField := strings.ToLower(fieldName)
	for pattern, desc := range patterns {
		if lowerField == pattern || strings.Contains(lowerField, pattern) {
			if typeContext != "" {
				return desc + typeContext + "."
			}
			return desc + "."
		}
	}

	// Default description
	if typeContext != "" {
		return fieldDesc + typeContext + "."
	}
	return fieldDesc + "."
}

// generateMessageDescription generates description for messages
func (dg *DescriptionGenerator) generateMessageDescription(name string, isRequest, isResponse bool) string {
	if isRequest {
		// Extract method name from request
		methodName := strings.TrimSuffix(name, "Request")
		verb := dg.extractVerb(methodName)
		entity := dg.extractEntity(methodName)
		return fmt.Sprintf("Request message for %s %s operation.", strings.ToLower(verb), entity)
	}

	if isResponse {
		// Extract method name from response
		methodName := strings.TrimSuffix(name, "Response")
		verb := dg.extractVerb(methodName)
		entity := dg.extractEntity(methodName)
		return fmt.Sprintf("Response message for %s %s operation.", strings.ToLower(verb), entity)
	}

	// Split camelCase for entity descriptions
	words := dg.splitCamelCase(name)
	return fmt.Sprintf("Represents %s information.", strings.ToLower(strings.Join(words, " ")))
}

// generateEnumDescription generates description for enums
func (dg *DescriptionGenerator) generateEnumDescription(name string) string {
	words := dg.splitCamelCase(name)

	// Common enum patterns
	if strings.HasSuffix(name, "Status") {
		entity := strings.TrimSuffix(name, "Status")
		return fmt.Sprintf("Represents possible status values for %s.", entity)
	}
	if strings.HasSuffix(name, "Type") {
		entity := strings.TrimSuffix(name, "Type")
		return fmt.Sprintf("Defines types of %s.", entity)
	}
	if strings.HasSuffix(name, "State") {
		entity := strings.TrimSuffix(name, "State")
		return fmt.Sprintf("Represents state of %s.", entity)
	}
	if strings.HasSuffix(name, "Role") {
		return "Defines user roles and permissions."
	}
	if strings.HasSuffix(name, "Level") {
		return "Defines priority or severity levels."
	}

	return fmt.Sprintf("Enumeration of %s values.", strings.ToLower(strings.Join(words, " ")))
}

// getStreamingDescription returns streaming type description
func (dg *DescriptionGenerator) getStreamingDescription(clientStreaming, serverStreaming bool) string {
	if clientStreaming && serverStreaming {
		return "Uses bidirectional streaming for real-time communication"
	} else if serverStreaming {
		return "Uses server-side streaming to deliver multiple responses"
	} else if clientStreaming {
		return "Uses client-side streaming to send multiple requests"
	}
	return ""
}

// extractVerb extracts the action verb from method name
func (dg *DescriptionGenerator) extractVerb(methodName string) string {
	verbs := []string{"Create", "Get", "Update", "Delete", "List", "Search", "Batch",
		"Stream", "Subscribe", "Cancel", "Validate", "Process", "Calculate", "Generate",
		"Send", "Receive", "Sync", "Import", "Export", "Upload", "Download"}

	for _, verb := range verbs {
		if strings.HasPrefix(methodName, verb) {
			return verb
		}
	}

	// Return first word as verb
	words := dg.splitCamelCase(methodName)
	if len(words) > 0 {
		return words[0]
	}
	return methodName
}

// extractEntity extracts the entity name from method name
func (dg *DescriptionGenerator) extractEntity(methodName string) string {
	verb := dg.extractVerb(methodName)
	entity := strings.TrimPrefix(methodName, verb)

	if entity == "" {
		return "resource"
	}

	// Convert to lowercase for better readability
	words := dg.splitCamelCase(entity)
	return strings.ToLower(strings.Join(words, " "))
}

// splitCamelCase splits camelCase string into words
func (dg *DescriptionGenerator) splitCamelCase(s string) []string {
	// Go's regexp doesn't support negative lookahead, so use a simpler pattern
	re := regexp.MustCompile(`([A-Z]+[a-z]*|[a-z]+|\d+)`)
	matches := re.FindAllString(s, -1)
	return matches
}

// splitSnakeCase splits snake_case string into words
func (dg *DescriptionGenerator) splitSnakeCase(s string) []string {
	return strings.Split(s, "_")
}

// DetectDeprecation checks if field/message is deprecated
func (dg *DescriptionGenerator) DetectDeprecation(options interface{}) (bool, string) {
	// Check for deprecated option in various descriptor types
	switch opts := options.(type) {
	case *descriptorpb.FieldOptions:
		if opts != nil && opts.GetDeprecated() {
			return true, "⚠️ DEPRECATED - This field is deprecated and should not be used in new code."
		}
	case *descriptorpb.MessageOptions:
		if opts != nil && opts.GetDeprecated() {
			return true, "⚠️ DEPRECATED - This message is deprecated and should not be used in new code."
		}
	case *descriptorpb.MethodOptions:
		if opts != nil && opts.GetDeprecated() {
			return true, "⚠️ DEPRECATED - This method is deprecated and should not be used in new code."
		}
	case *descriptorpb.EnumOptions:
		if opts != nil && opts.GetDeprecated() {
			return true, "⚠️ DEPRECATED - This enum is deprecated and should not be used in new code."
		}
	case *descriptorpb.EnumValueOptions:
		if opts != nil && opts.GetDeprecated() {
			return true, "⚠️ DEPRECATED - This enum value is deprecated and should not be used."
		}
	}
	return false, ""
}

// AddValidationHints adds validation and constraint information to description
func (dg *DescriptionGenerator) AddValidationHints(fieldName, fieldType string, description string) string {
	hints := []string{}

	lowerName := strings.ToLower(fieldName)

	// Email validation
	if strings.Contains(lowerName, "email") {
		hints = append(hints, "Must be a valid email address format")
	}

	// URL validation
	if strings.Contains(lowerName, "url") || strings.Contains(lowerName, "uri") {
		hints = append(hints, "Must be a valid URL")
	}

	// ID fields
	if lowerName == "id" || strings.HasSuffix(lowerName, "_id") {
		hints = append(hints, "Must be a non-empty identifier")
	}

	// UUID fields
	if strings.Contains(lowerName, "uuid") {
		hints = append(hints, "Must be a valid UUID format")
	}

	// Phone fields
	if strings.Contains(lowerName, "phone") {
		hints = append(hints, "Should follow E.164 format")
	}

	// Token fields
	if strings.Contains(lowerName, "token") {
		hints = append(hints, "Sensitive - should be transmitted securely")
	}

	// Password fields
	if strings.Contains(lowerName, "password") {
		hints = append(hints, "Sensitive - should be hashed/encrypted")
	}

	// Timestamp fields
	if strings.Contains(lowerName, "timestamp") || strings.Contains(lowerName, "_at") {
		hints = append(hints, "RFC 3339 timestamp format")
	}

	if len(hints) > 0 {
		return fmt.Sprintf("%s (%s)", description, strings.Join(hints, "; "))
	}

	return description
}

// GenerateFieldConstraints generates constraint documentation
func (dg *DescriptionGenerator) GenerateFieldConstraints(fieldName, fieldType string) string {
	constraints := []string{}

	lowerName := strings.ToLower(fieldName)

	// Common constraints based on naming patterns
	if strings.Contains(lowerName, "page_size") {
		constraints = append(constraints, "Typical range: 1-100")
	}
	if strings.Contains(lowerName, "limit") {
		constraints = append(constraints, "Maximum value may be service-specific")
	}
	if strings.Contains(lowerName, "offset") {
		constraints = append(constraints, "Must be >= 0")
	}
	if strings.Contains(lowerName, "quantity") || strings.Contains(lowerName, "count") {
		constraints = append(constraints, "Must be >= 0")
	}
	if strings.Contains(lowerName, "percentage") || strings.Contains(lowerName, "percent") {
		constraints = append(constraints, "Range: 0-100")
	}
	if strings.Contains(lowerName, "priority") {
		constraints = append(constraints, "Higher values indicate higher priority")
	}

	if len(constraints) > 0 {
		return " " + strings.Join(constraints, "; ") + "."
	}

	return ""
}
