package resolver

import (
	"testing"

	"github.com/kyivinua/docgen-tool/internal/docgen"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSharedResourceResolver(t *testing.T) {
	logger := zerolog.Nop()
	config := Config{
		AutoDiscover: true,
	}

	resolver := NewSharedResourceResolver(config, logger)

	assert.NotNil(t, resolver)
	assert.Len(t, resolver.config.CommonDirPatterns, 6) // Default patterns
}

func TestParseProtoFile(t *testing.T) {
	logger := zerolog.Nop()
	resolver := NewSharedResourceResolver(Config{}, logger)

	protoContent := `
syntax = "proto3";

package user.v1;

import "common/types.proto";
import "google/protobuf/timestamp.proto";

message User {
	string id = 1;
	string name = 2;
	string email = 3;
}

message CreateUserRequest {
	string name = 1;
	string email = 2;
}

service UserService {
	rpc CreateUser(CreateUserRequest) returns (User);
	rpc GetUser(GetUserRequest) returns (User);
}
`

	proto := resolver.parseProtoFile("test.proto", protoContent)

	assert.Equal(t, "test.proto", proto.Path)
	assert.Equal(t, "user.v1", proto.Package)
	assert.Len(t, proto.Imports, 2)
	assert.Contains(t, proto.Imports, "common/types.proto")
	assert.Contains(t, proto.Imports, "google/protobuf/timestamp.proto")
	assert.Len(t, proto.Messages, 2)
	assert.Len(t, proto.Services, 1)
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"package user.v1;", "user.v1"},
		{"package   foo.bar.v2  ;", "foo.bar.v2"},
		{"package test;", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := extractPackageName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractImportPath(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{`import "common/types.proto";`, "common/types.proto"},
		{`import   "google/protobuf/timestamp.proto"  ;`, "google/protobuf/timestamp.proto"},
		{`import "foo.proto";`, "foo.proto"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := extractImportPath(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractMessageName(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"message User {", "User"},
		{"message   CreateUserRequest  {", "CreateUserRequest"},
		{"message Foo{", "Foo"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := extractMessageName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"service UserService {", "UserService"},
		{"service   FooService  {", "FooService"},
		{"service Bar{", "Bar"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := extractServiceName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMinimalService(t *testing.T) {
	logger := zerolog.Nop()
	resolver := NewSharedResourceResolver(Config{}, logger)

	service := resolver.generateMinimalService("UserService")

	assert.Equal(t, "UserService", service.Name)
	assert.NotEmpty(t, service.Package)
	assert.Len(t, service.Methods, 1)
	assert.Len(t, service.Messages, 2)
	assert.Equal(t, "Get", service.Methods[0].Name)
}

func TestInferPackageName(t *testing.T) {
	tests := []struct {
		serviceName string
		expected    string
	}{
		{"UserService", "user.v1"},
		{"ProductService", "product.v1"},
		{"OrderManagementService", "ordermanagement.v1"},
	}

	for _, tt := range tests {
		t.Run(tt.serviceName, func(t *testing.T) {
			result := inferPackageName(tt.serviceName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user.v1", "user_v1"},
		{"foo/bar", "foo_bar"},
		{"test.user.v1", "test_user_v1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeduplicateMessages(t *testing.T) {
	logger := zerolog.Nop()
	resolver := NewSharedResourceResolver(Config{}, logger)

	service := &docgen.Service{
		Name:    "TestService",
		Package: "test.v1",
		Messages: []docgen.Message{
			{Name: "User"},
			{Name: "User"}, // Duplicate
			{Name: "Product"},
			{Name: "User"}, // Another duplicate
		},
	}

	resolver.deduplicateMessages(service)

	assert.Len(t, service.Messages, 2)
	assert.Equal(t, "User", service.Messages[0].Name)
	assert.Equal(t, "Product", service.Messages[1].Name)
}

func TestImportGraph(t *testing.T) {
	graph := NewImportGraph()

	graph.AddNode("a.proto", []string{"b.proto", "c.proto"})
	graph.AddNode("b.proto", []string{"c.proto"})
	graph.AddNode("c.proto", []string{})

	// Test GetDependencies
	deps := graph.GetDependencies("a.proto")
	assert.Len(t, deps, 2)
	assert.Contains(t, deps, "b.proto")
	assert.Contains(t, deps, "c.proto")

	deps = graph.GetDependencies("b.proto")
	assert.Len(t, deps, 1)
	assert.Contains(t, deps, "c.proto")

	deps = graph.GetDependencies("c.proto")
	assert.Len(t, deps, 0)
}

func TestIsMessageRelevant(t *testing.T) {
	logger := zerolog.Nop()
	resolver := NewSharedResourceResolver(Config{}, logger)

	service := &docgen.Service{
		Name: "UserService",
		Methods: []docgen.Method{
			{
				Name:       "CreateUser",
				InputType:  "CreateUserRequest",
				OutputType: "User",
			},
		},
	}

	tests := []struct {
		name     string
		message  *docgen.Message
		relevant bool
	}{
		{
			name:     "message used in method input",
			message:  &docgen.Message{Name: "CreateUserRequest"},
			relevant: true,
		},
		{
			name:     "message used in method output",
			message:  &docgen.Message{Name: "User"},
			relevant: true,
		},
		{
			name:     "message with Request suffix",
			message:  &docgen.Message{Name: "SomeRequest"},
			relevant: true,
		},
		{
			name:     "message with Response suffix",
			message:  &docgen.Message{Name: "SomeResponse"},
			relevant: true,
		},
		{
			name:     "common message",
			message:  &docgen.Message{Name: "CommonData"},
			relevant: true,
		},
		{
			name:     "unrelated message",
			message:  &docgen.Message{Name: "UnrelatedData"},
			relevant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.isMessageRelevant(service, tt.message)
			assert.Equal(t, tt.relevant, result)
		})
	}
}

func TestHandleServiceWithoutProtos(t *testing.T) {
	logger := zerolog.Nop()
	resolver := NewSharedResourceResolver(Config{}, logger)

	service, err := resolver.HandleServiceWithoutProtos("TestService")

	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "TestService", service.Name)
	assert.NotEmpty(t, service.Package)
	assert.Greater(t, len(service.Methods), 0)
	assert.Greater(t, len(service.Messages), 0)
}

func TestMapKeys(t *testing.T) {
	m := map[string]bool{
		"a": true,
		"b": false,
		"c": true,
	}

	keys := mapKeys(m)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "a")
	assert.Contains(t, keys, "b")
	assert.Contains(t, keys, "c")
}
