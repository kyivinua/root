package docgen

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/internal/validation"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ProtoParser parses protocol buffer files and extracts documentation
type ProtoParser struct {
	protoFiles  []string
	importPaths []string
	descGen     *DescriptionGenerator
}

// NewProtoParser creates a new proto parser
func NewProtoParser(protoFiles, importPaths []string) (*ProtoParser, error) {
	// Validate all proto files
	validatedProtoFiles := make([]string, 0, len(protoFiles))
	for _, protoFile := range protoFiles {
		validPath, err := validation.ValidateProtoFile(protoFile)
		if err != nil {
			return nil, fmt.Errorf("invalid proto file %q: %w", protoFile, err)
		}
		validatedProtoFiles = append(validatedProtoFiles, validPath)
	}

	// Validate all import paths
	validatedImportPaths := make([]string, 0, len(importPaths))
	for _, importPath := range importPaths {
		validPath, err := validation.ValidateDirectory(importPath)
		if err != nil {
			return nil, fmt.Errorf("invalid import path %q: %w", importPath, err)
		}
		validatedImportPaths = append(validatedImportPaths, validPath)
	}

	return &ProtoParser{
		protoFiles:  validatedProtoFiles,
		importPaths: validatedImportPaths,
		descGen:     NewDescriptionGenerator(),
	}, nil
}

// Parse parses proto files and returns service documentation
func (p *ProtoParser) Parse() ([]*ServiceDocumentation, error) {
	// Generate FileDescriptorSet using protoc
	descriptorSet, err := p.generateDescriptorSet()
	if err != nil {
		return nil, fmt.Errorf("failed to generate descriptor set: %w", err)
	}

	// Parse descriptor set
	docs, err := p.parseDescriptorSet(descriptorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse descriptor set: %w", err)
	}

	return docs, nil
}

// generateDescriptorSet runs protoc to generate FileDescriptorSet
func (p *ProtoParser) generateDescriptorSet() (*descriptorpb.FileDescriptorSet, error) {
	// Create temporary file for descriptor output
	tmpFile, err := os.CreateTemp("", "proto_descriptor_*.pb")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Build protoc command
	args := []string{
		"--descriptor_set_out=" + tmpFile.Name(),
		"--include_imports",
		"--include_source_info",
	}

	// Add import paths
	for _, importPath := range p.importPaths {
		args = append(args, "--proto_path="+importPath)
	}

	// Add proto files
	args = append(args, p.protoFiles...)

	// Run protoc
	cmd := exec.Command("protoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("protoc failed: %w\nOutput: %s", err, string(output))
	}

	// Read descriptor set
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("read descriptor set: %w", err)
	}

	// Unmarshal descriptor set
	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return nil, fmt.Errorf("unmarshal descriptor set: %w", err)
	}

	return fds, nil
}

// parseDescriptorSet parses FileDescriptorSet into ServiceDocumentation
func (p *ProtoParser) parseDescriptorSet(fds *descriptorpb.FileDescriptorSet) ([]*ServiceDocumentation, error) {
	var allDocs []*ServiceDocumentation

	for _, file := range fds.GetFile() {
		// Skip google proto files
		if strings.HasPrefix(file.GetName(), "google/protobuf/") {
			continue
		}

		// Parse services from this file
		for _, service := range file.GetService() {
			doc := p.parseService(file, service)
			allDocs = append(allDocs, doc)
		}
	}

	return allDocs, nil
}

// parseService parses a service descriptor
func (p *ProtoParser) parseService(file *descriptorpb.FileDescriptorProto, service *descriptorpb.ServiceDescriptorProto) *ServiceDocumentation {
	// Find service index
	serviceIndex := -1
	for i, svc := range file.GetService() {
		if svc.GetName() == service.GetName() {
			serviceIndex = i
			break
		}
	}

	// Extract service description from source code info
	serviceDesc := ""
	if serviceIndex >= 0 {
		path := []int32{6, int32(serviceIndex)} // 6 = service field number
		serviceDesc = extractCommentFromPath(file, path)
	}

	// Enhance service description with smart generation
	enhancedServiceDesc := p.descGen.EnhanceServiceDescription(service.GetName(), serviceDesc)

	// Extract version from package name
	apiVersion := extractVersion(file.GetPackage())

	doc := &ServiceDocumentation{
		Service: &ServiceDoc{
			Name:        service.GetName(),
			FullName:    fmt.Sprintf("%s.%s", file.GetPackage(), service.GetName()),
			Package:     file.GetPackage(),
			Description: enhancedServiceDesc,
			Version:     apiVersion,
			ProtoFile:   file.GetName(),
		},
		Methods:  make([]MethodDoc, 0),
		Messages: make([]MessageDoc, 0),
		Enums:    make([]EnumDoc, 0),
		Diagrams: make(map[string]string),
		Metadata: DocumentMetadata{
			Generated: time.Now(),
			Version:   GeneratorVersion,
		},
	}

	// Parse methods
	for i, method := range service.GetMethod() {
		methodDoc := p.parseMethod(file, serviceIndex, i, method)
		doc.Methods = append(doc.Methods, methodDoc)
	}

	// Parse messages used by this service
	messagesMap := p.collectServiceMessages(file, service)
	for _, msg := range messagesMap {
		doc.Messages = append(doc.Messages, msg)
	}

	// Parse enums used by this service
	enumsMap := p.collectServiceEnums(file, service)
	for _, enum := range enumsMap {
		doc.Enums = append(doc.Enums, enum)
	}

	return doc
}

// parseMethod parses a method descriptor
func (p *ProtoParser) parseMethod(file *descriptorpb.FileDescriptorProto, serviceIndex, methodIndex int, method *descriptorpb.MethodDescriptorProto) MethodDoc {
	// Extract method description
	methodDesc := ""
	if serviceIndex >= 0 && methodIndex >= 0 {
		path := []int32{6, int32(serviceIndex), 2, int32(methodIndex)} // 6=service, 2=method
		methodDesc = extractCommentFromPath(file, path)
	}

	inputType := strings.TrimPrefix(method.GetInputType(), ".")
	outputType := strings.TrimPrefix(method.GetOutputType(), ".")

	// Enhance method description with smart generation
	enhancedMethodDesc := p.descGen.EnhanceMethodDescription(
		method.GetName(),
		methodDesc,
		inputType,
		outputType,
		method.GetClientStreaming(),
		method.GetServerStreaming(),
	)

	// Check for deprecation
	deprecated, deprecationMsg := p.descGen.DetectDeprecation(method.GetOptions())
	if deprecated {
		enhancedMethodDesc = deprecationMsg + "\n\n" + enhancedMethodDesc
	}

	return MethodDoc{
		Name:            method.GetName(),
		FullName:        fmt.Sprintf("%s.%s", file.GetPackage(), method.GetName()),
		Description:     enhancedMethodDesc,
		InputType:       inputType,
		OutputType:      outputType,
		ClientStreaming: method.GetClientStreaming(),
		ServerStreaming: method.GetServerStreaming(),
		HTTPBindings:    p.extractHTTPBindings(method),
		Examples:        make([]Example, 0),
	}
}

// collectServiceMessages collects all messages used by the service
func (p *ProtoParser) collectServiceMessages(file *descriptorpb.FileDescriptorProto, service *descriptorpb.ServiceDescriptorProto) map[string]MessageDoc {
	messages := make(map[string]MessageDoc)
	toProcess := make([]string, 0)

	// Collect from methods
	for _, method := range service.GetMethod() {
		inputType := strings.TrimPrefix(method.GetInputType(), ".")
		outputType := strings.TrimPrefix(method.GetOutputType(), ".")
		toProcess = append(toProcess, inputType, outputType)
	}

	// Process all messages recursively
	processed := make(map[string]bool)
	for len(toProcess) > 0 {
		// Pop first item
		typeName := toProcess[0]
		toProcess = toProcess[1:]

		if processed[typeName] {
			continue
		}
		processed[typeName] = true

		// Find and parse the message
		for _, msg := range file.GetMessageType() {
			fullName := fmt.Sprintf("%s.%s", file.GetPackage(), msg.GetName())
			if fullName == typeName {
				msgDoc := p.parseMessage(file, msg)
				messages[fullName] = msgDoc

				// Collect referenced types from fields
				for _, field := range msgDoc.Fields {
					if field.TypeName != "" {
						// Skip well-known types
						if !strings.HasPrefix(field.TypeName, "google.protobuf.") {
							toProcess = append(toProcess, field.TypeName)
						}
					}
				}
				break
			}
		}
	}

	return messages
}

// parseMessage parses a message descriptor
func (p *ProtoParser) parseMessage(file *descriptorpb.FileDescriptorProto, msg *descriptorpb.DescriptorProto) MessageDoc {
	fullName := fmt.Sprintf("%s.%s", file.GetPackage(), msg.GetName())

	// Find message index
	messageIndex := -1
	for i, m := range file.GetMessageType() {
		if m.GetName() == msg.GetName() {
			messageIndex = i
			break
		}
	}

	// Extract message description
	messageDesc := ""
	if messageIndex >= 0 {
		path := []int32{4, int32(messageIndex)} // 4 = message_type field number
		messageDesc = extractCommentFromPath(file, path)
	}

	// Detect if this is a request or response message
	msgName := msg.GetName()
	isRequest := strings.HasSuffix(msgName, "Request")
	isResponse := strings.HasSuffix(msgName, "Response")

	// Enhance message description with smart generation
	enhancedMessageDesc := p.descGen.EnhanceMessageDescription(msgName, messageDesc, isRequest, isResponse)

	// Check for deprecation
	deprecated, deprecationMsg := p.descGen.DetectDeprecation(msg.GetOptions())
	if deprecated {
		enhancedMessageDesc = deprecationMsg + "\n\n" + enhancedMessageDesc
	}

	doc := MessageDoc{
		Name:        msgName,
		FullName:    fullName,
		Description: enhancedMessageDesc,
		Fields:      make([]FieldDoc, 0),
		NestedTypes: make([]string, 0),
	}

	// Parse fields
	for i, field := range msg.GetField() {
		fieldDoc := p.parseFieldWithOneofs(file, messageIndex, i, field, msg)
		doc.Fields = append(doc.Fields, fieldDoc)
	}

	// Parse nested types
	for _, nested := range msg.GetNestedType() {
		doc.NestedTypes = append(doc.NestedTypes, nested.GetName())
	}

	return doc
}

// parseFieldWithOneofs parses a field descriptor with oneof information
func (p *ProtoParser) parseFieldWithOneofs(file *descriptorpb.FileDescriptorProto, messageIndex, fieldIndex int, field *descriptorpb.FieldDescriptorProto, msg *descriptorpb.DescriptorProto) FieldDoc {
	label := "optional"
	if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		label = "repeated"
	} else if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
		label = "required"
	}

	// Convert protobuf type enum to readable name
	fieldType := getFieldTypeName(field.GetType())
	typeName := strings.TrimPrefix(field.GetTypeName(), ".")

	// Check if this is a map field
	if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE && field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		// Check if the message is a map entry
		for _, nested := range msg.GetNestedType() {
			nestedFullName := nested.GetName()
			fieldTypeName := getShortTypeName(field.GetTypeName())
			if nestedFullName == fieldTypeName && nested.GetOptions().GetMapEntry() {
				// This is a map field - format as map<K, V>
				var keyType, valueType string
				for _, mapField := range nested.GetField() {
					if mapField.GetName() == "key" {
						keyType = getFieldTypeName(mapField.GetType())
						if keyType == "" {
							keyType = getShortTypeName(mapField.GetTypeName())
						}
					} else if mapField.GetName() == "value" {
						valueType = getFieldTypeName(mapField.GetType())
						if valueType == "" {
							valueType = getShortTypeName(mapField.GetTypeName())
						}
					}
				}
				if keyType != "" && valueType != "" {
					fieldType = fmt.Sprintf("map<%s, %s>", keyType, valueType)
					typeName = "" // Clear typeName since we have the full map syntax
					label = ""     // Maps don't need labels
				}
			}
		}
	}

	// Get oneof group name if applicable
	oneofGroup := ""
	if field.OneofIndex != nil && field.GetOneofIndex() >= 0 {
		oneofDecls := msg.GetOneofDecl()
		if int(field.GetOneofIndex()) < len(oneofDecls) {
			oneofGroup = oneofDecls[field.GetOneofIndex()].GetName()
		}
	}

	// If oneof field, show oneof group in label instead of optional
	if oneofGroup != "" {
		label = fmt.Sprintf("oneof `%s`", oneofGroup)
	}

	// Extract field description
	fieldDesc := ""
	if messageIndex >= 0 && fieldIndex >= 0 {
		path := []int32{4, int32(messageIndex), 2, int32(fieldIndex)} // 4=message, 2=field
		fieldDesc = extractCommentFromPath(file, path)
	}

	// Enhance field description with smart generation
	isRepeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	enhancedFieldDesc := p.descGen.EnhanceFieldDescription(
		field.GetName(),
		fieldDesc,
		fieldType,
		typeName,
		isRepeated,
		oneofGroup != "",
	)

	// Add validation hints
	enhancedFieldDesc = p.descGen.AddValidationHints(field.GetName(), fieldType, enhancedFieldDesc)

	// Add constraint hints
	constraints := p.descGen.GenerateFieldConstraints(field.GetName(), fieldType)
	if constraints != "" {
		enhancedFieldDesc = strings.TrimSuffix(enhancedFieldDesc, ".") + constraints
	}

	// Check for deprecation
	deprecated, deprecationMsg := p.descGen.DetectDeprecation(field.GetOptions())
	if deprecated {
		enhancedFieldDesc = deprecationMsg + " " + enhancedFieldDesc
	}

	return FieldDoc{
		Name:         field.GetName(),
		Number:       field.GetNumber(),
		Type:         fieldType,
		TypeName:     typeName,
		Label:        label,
		Description:  enhancedFieldDesc,
		OneofGroup:   oneofGroup,
		DefaultValue: "",
	}
}

// collectServiceEnums collects all enums used by the service
func (p *ProtoParser) collectServiceEnums(file *descriptorpb.FileDescriptorProto, service *descriptorpb.ServiceDescriptorProto) map[string]EnumDoc {
	enums := make(map[string]EnumDoc)

	for _, enum := range file.GetEnumType() {
		fullName := fmt.Sprintf("%s.%s", file.GetPackage(), enum.GetName())
		enums[fullName] = p.parseEnum(file, enum)
	}

	return enums
}

// parseEnum parses an enum descriptor
func (p *ProtoParser) parseEnum(file *descriptorpb.FileDescriptorProto, enum *descriptorpb.EnumDescriptorProto) EnumDoc {
	fullName := fmt.Sprintf("%s.%s", file.GetPackage(), enum.GetName())

	// Find enum index
	enumIndex := -1
	for i, e := range file.GetEnumType() {
		if e.GetName() == enum.GetName() {
			enumIndex = i
			break
		}
	}

	// Extract enum description
	enumDesc := ""
	if enumIndex >= 0 {
		path := []int32{5, int32(enumIndex)} // 5 = enum_type field number
		enumDesc = extractCommentFromPath(file, path)
	}

	// Enhance enum description with smart generation
	enhancedEnumDesc := p.descGen.EnhanceEnumDescription(enum.GetName(), enumDesc)

	// Check for deprecation
	deprecated, deprecationMsg := p.descGen.DetectDeprecation(enum.GetOptions())
	if deprecated {
		enhancedEnumDesc = deprecationMsg + "\n\n" + enhancedEnumDesc
	}

	doc := EnumDoc{
		Name:        enum.GetName(),
		FullName:    fullName,
		Description: enhancedEnumDesc,
		Values:      make([]EnumValueDoc, 0),
	}

	for i, value := range enum.GetValue() {
		// Extract enum value description
		valueDesc := ""
		if enumIndex >= 0 {
			path := []int32{5, int32(enumIndex), 2, int32(i)} // 5=enum, 2=value
			valueDesc = extractCommentFromPath(file, path)
		}

		// If no description, use formatting at minimum
		if valueDesc == "" {
			valueDesc = fmt.Sprintf("%s value.", value.GetName())
		}

		// Check for deprecation on enum values
		deprecated, deprecationMsg := p.descGen.DetectDeprecation(value.GetOptions())
		if deprecated {
			valueDesc = deprecationMsg + " " + valueDesc
		}

		doc.Values = append(doc.Values, EnumValueDoc{
			Name:        value.GetName(),
			Number:      value.GetNumber(),
			Description: valueDesc,
		})
	}

	return doc
}

// extractCommentFromPath extracts comments from source code info for a given path
func extractCommentFromPath(file *descriptorpb.FileDescriptorProto, path []int32) string {
	if file.GetSourceCodeInfo() == nil {
		return ""
	}

	// Match the path in source code info
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if pathsEqual(loc.GetPath(), path) {
			// Prefer leading comments, fall back to trailing comments
			if loc.GetLeadingComments() != "" {
				return strings.TrimSpace(loc.GetLeadingComments())
			}
			if loc.GetTrailingComments() != "" {
				return strings.TrimSpace(loc.GetTrailingComments())
			}
		}
	}

	return ""
}

// pathsEqual checks if two paths are equal
func pathsEqual(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// extractHTTPBindings extracts HTTP bindings from method options
func (p *ProtoParser) extractHTTPBindings(method *descriptorpb.MethodDescriptorProto) []HTTPBinding {
	// This would require parsing google.api.http annotations
	// For now, return empty slice
	return []HTTPBinding{}
}

// Helper method for FieldDoc
func (fd *FieldDoc) GetOneofGroup() string {
	if fd.OneofGroup != "" {
		return fd.OneofGroup
	}
	return ""
}

// getFieldTypeName converts protobuf field type enum to readable type name
func getFieldTypeName(fieldType descriptorpb.FieldDescriptorProto_Type) string {
	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "" // Will use TypeName instead
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "" // Will use TypeName instead
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	default:
		return fieldType.String() // Fallback to enum name
	}
}

// getShortTypeName extracts just the type name from a full type path
// e.g., "users.v1.UserPreferences" -> "UserPreferences"
// e.g., ".users.v1.CustomMetadataEntry" -> "CustomMetadataEntry"
func getShortTypeName(fullTypeName string) string {
	fullTypeName = strings.TrimPrefix(fullTypeName, ".")
	parts := strings.Split(fullTypeName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullTypeName
}

// extractVersion extracts API version from package name
// e.g., "users.v1" -> "v1"
// e.g., "analytics.v2alpha1" -> "v2alpha1"
// e.g., "common" -> "" (no version)
func extractVersion(packageName string) string {
	parts := strings.Split(packageName, ".")
	for _, part := range parts {
		// Check if part starts with 'v' followed by a digit
		if len(part) > 1 && strings.HasPrefix(part, "v") {
			// Check if the character after 'v' is a digit
			if part[1] >= '0' && part[1] <= '9' {
				return part
			}
		}
	}
	return "" // No version found
}
