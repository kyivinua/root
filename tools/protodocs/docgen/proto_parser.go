package docgen

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ProtoParser parses protocol buffer files and extracts documentation
type ProtoParser struct {
	protoFiles []string
	importPaths []string
}

// NewProtoParser creates a new proto parser
func NewProtoParser(protoFiles, importPaths []string) *ProtoParser {
	return &ProtoParser{
		protoFiles: protoFiles,
		importPaths: importPaths,
	}
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
	doc := &ServiceDocumentation{
		Service: &ServiceDoc{
			Name:        service.GetName(),
			FullName:    fmt.Sprintf("%s.%s", file.GetPackage(), service.GetName()),
			Package:     file.GetPackage(),
			Description: "", // Will be extracted from source code info in enhanced version
			ProtoFile:   file.GetName(),
		},
		Methods:  make([]MethodDoc, 0),
		Messages: make([]MessageDoc, 0),
		Enums:    make([]EnumDoc, 0),
		Diagrams: make(map[string]string),
	}

	// Parse methods
	for _, method := range service.GetMethod() {
		methodDoc := p.parseMethod(file, method)
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
func (p *ProtoParser) parseMethod(file *descriptorpb.FileDescriptorProto, method *descriptorpb.MethodDescriptorProto) MethodDoc {
	return MethodDoc{
		Name:            method.GetName(),
		FullName:        fmt.Sprintf("%s.%s", file.GetPackage(), method.GetName()),
		Description:     "", // Will be extracted from source code info in enhanced version
		InputType:       strings.TrimPrefix(method.GetInputType(), "."),
		OutputType:      strings.TrimPrefix(method.GetOutputType(), "."),
		ClientStreaming: method.GetClientStreaming(),
		ServerStreaming: method.GetServerStreaming(),
		HTTPBindings:    p.extractHTTPBindings(method),
		Examples:        make([]Example, 0),
	}
}

// collectServiceMessages collects all messages used by the service
func (p *ProtoParser) collectServiceMessages(file *descriptorpb.FileDescriptorProto, service *descriptorpb.ServiceDescriptorProto) map[string]MessageDoc {
	messages := make(map[string]MessageDoc)

	// Collect from methods
	for _, method := range service.GetMethod() {
		inputType := strings.TrimPrefix(method.GetInputType(), ".")
		outputType := strings.TrimPrefix(method.GetOutputType(), ".")

		// Find message descriptors
		for _, msg := range file.GetMessageType() {
			fullName := fmt.Sprintf("%s.%s", file.GetPackage(), msg.GetName())
			if fullName == inputType || fullName == outputType {
				messages[fullName] = p.parseMessage(file, msg)
			}
		}
	}

	return messages
}

// parseMessage parses a message descriptor
func (p *ProtoParser) parseMessage(file *descriptorpb.FileDescriptorProto, msg *descriptorpb.DescriptorProto) MessageDoc {
	fullName := fmt.Sprintf("%s.%s", file.GetPackage(), msg.GetName())

	doc := MessageDoc{
		Name:        msg.GetName(),
		FullName:    fullName,
		Description: "", // Will be extracted from source code info in enhanced version
		Fields:      make([]FieldDoc, 0),
		NestedTypes: make([]string, 0),
	}

	// Parse fields
	for _, field := range msg.GetField() {
		fieldDoc := p.parseFieldWithOneofs(field, msg)
		doc.Fields = append(doc.Fields, fieldDoc)
	}

	// Parse nested types
	for _, nested := range msg.GetNestedType() {
		doc.NestedTypes = append(doc.NestedTypes, nested.GetName())
	}

	return doc
}

// parseFieldWithOneofs parses a field descriptor with oneof information
func (p *ProtoParser) parseFieldWithOneofs(field *descriptorpb.FieldDescriptorProto, msg *descriptorpb.DescriptorProto) FieldDoc {
	label := "optional"
	if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		label = "repeated"
	} else if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
		label = "required"
	}

	fieldType := field.GetType().String()
	typeName := strings.TrimPrefix(field.GetTypeName(), ".")

	// Get oneof group name if applicable
	oneofGroup := ""
	if field.OneofIndex != nil && field.GetOneofIndex() >= 0 {
		oneofDecls := msg.GetOneofDecl()
		if int(field.GetOneofIndex()) < len(oneofDecls) {
			oneofGroup = oneofDecls[field.GetOneofIndex()].GetName()
		}
	}

	return FieldDoc{
		Name:         field.GetName(),
		Number:       field.GetNumber(),
		Type:         fieldType,
		TypeName:     typeName,
		Label:        label,
		Description:  "", // Will be extracted from source code info in enhanced version
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

	doc := EnumDoc{
		Name:        enum.GetName(),
		FullName:    fullName,
		Description: "", // Will be extracted from source code info in enhanced version
		Values:      make([]EnumValueDoc, 0),
	}

	for _, value := range enum.GetValue() {
		doc.Values = append(doc.Values, EnumValueDoc{
			Name:        value.GetName(),
			Number:      value.GetNumber(),
			Description: "", // Will be extracted from source code info in enhanced version
		})
	}

	return doc
}

// extractLeadingComments extracts leading comments from source code info
// This is a placeholder for future enhancement
// Full implementation would use source code info paths to match specific elements
func extractLeadingComments(file *descriptorpb.FileDescriptorProto, path []int32) string {
	if file.GetSourceCodeInfo() == nil {
		return ""
	}

	// Match the path in source code info
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if pathsEqual(loc.GetPath(), path) {
			if loc.GetLeadingComments() != "" {
				return strings.TrimSpace(loc.GetLeadingComments())
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
