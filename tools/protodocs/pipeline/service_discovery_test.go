package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonorepoDiscovery_ParseProtoFile(t *testing.T) {
	// Create temporary proto file
	tmpDir := t.TempDir()
	protoContent := `syntax = "proto3";

package user.v1;

import "google/protobuf/timestamp.proto";
import "common/types.proto";

// UserService manages user operations
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

// AccountService manages accounts
service AccountService {
  rpc GetAccount(GetAccountRequest) returns (GetAccountResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  User user = 1;
}
`

	protoPath := filepath.Join(tmpDir, "user_service.proto")
	err := os.WriteFile(protoPath, []byte(protoContent), 0644)
	require.NoError(t, err)

	// Create discovery instance
	config := DefaultMonorepoDiscoveryConfig(tmpDir)
	md := NewMonorepoDiscovery(config)

	// Parse file
	parsed, err := md.parseProtoFile(protoPath)
	require.NoError(t, err)
	assert.NotNil(t, parsed)

	// Verify parsed data
	assert.Equal(t, protoPath, parsed.FilePath)
	assert.Equal(t, "user.v1", parsed.PackageName)
	assert.Len(t, parsed.Services, 2)
	assert.Contains(t, parsed.Services, "UserService")
	assert.Contains(t, parsed.Services, "AccountService")
	assert.Len(t, parsed.Dependencies, 2)
	assert.Contains(t, parsed.Dependencies, "google/protobuf/timestamp.proto")
	assert.Contains(t, parsed.Dependencies, "common/types.proto")
}

func TestMonorepoDiscovery_DetectByServiceDefinition(t *testing.T) {
	md := NewMonorepoDiscovery(nil)

	tests := []struct {
		name     string
		file     *ServiceProtoFile
		expected string
	}{
		{
			name: "file with service",
			file: &ServiceProtoFile{
				Services: []string{"UserService", "AccountService"},
			},
			expected: "UserService",
		},
		{
			name: "file without service",
			file: &ServiceProtoFile{
				Services: []string{},
			},
			expected: "common",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.detectByServiceDefinition(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonorepoDiscovery_DetectByDirectory(t *testing.T) {
	md := NewMonorepoDiscovery(DefaultMonorepoDiscoveryConfig("/monorepo"))

	tests := []struct {
		name         string
		relativePath string
		expected     string
	}{
		{
			name:         "services directory pattern",
			relativePath: "services/user-service/api/v1/user.proto",
			expected:     "user-service",
		},
		{
			name:         "pkg directory pattern",
			relativePath: "pkg/billing-service/proto/billing.proto",
			expected:     "billing-service",
		},
		{
			name:         "apps directory pattern",
			relativePath: "apps/notifications/api/notify.proto",
			expected:     "notifications",
		},
		{
			name:         "service suffix pattern",
			relativePath: "internal/auth-service/v1/auth.proto",
			expected:     "auth-service",
		},
		{
			name:         "api suffix pattern",
			relativePath: "internal/payment-api/v1/payment.proto",
			expected:     "payment-api",
		},
		{
			name:         "unknown pattern",
			relativePath: "random/path/file.proto",
			expected:     "random",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &ServiceProtoFile{
				RelativePath: tt.relativePath,
			}
			result := md.detectByDirectory(file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonorepoDiscovery_DetectByPackage(t *testing.T) {
	md := NewMonorepoDiscovery(nil)

	tests := []struct {
		name        string
		packageName string
		expected    string
	}{
		{
			name:        "simple package",
			packageName: "user.v1",
			expected:    "user",
		},
		{
			name:        "company package",
			packageName: "com.company.billing.v1",
			expected:    "billing",
		},
		{
			name:        "org package",
			packageName: "org.example.auth.v2",
			expected:    "auth",
		},
		{
			name:        "nested package",
			packageName: "io.grpc.health.v1",
			expected:    "health",
		},
		{
			name:        "no version",
			packageName: "notifications.internal",
			expected:    "notifications",
		},
		{
			name:        "empty package",
			packageName: "",
			expected:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &ServiceProtoFile{
				PackageName: tt.packageName,
			}
			result := md.detectByPackage(file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonorepoDiscovery_DiscoverAll(t *testing.T) {
	// Create temporary monorepo structure
	tmpDir := t.TempDir()

	// Create service directories
	userServiceDir := filepath.Join(tmpDir, "services", "user-service", "api", "v1")
	billingServiceDir := filepath.Join(tmpDir, "services", "billing-service", "api", "v1")
	commonDir := filepath.Join(tmpDir, "pkg", "common", "proto")

	require.NoError(t, os.MkdirAll(userServiceDir, 0755))
	require.NoError(t, os.MkdirAll(billingServiceDir, 0755))
	require.NoError(t, os.MkdirAll(commonDir, 0755))

	// Create proto files
	userProto := `syntax = "proto3";
package user.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string name = 1;
}
`

	billingProto := `syntax = "proto3";
package billing.v1;

service BillingService {
  rpc CreateInvoice(CreateInvoiceRequest) returns (CreateInvoiceResponse);
}

message CreateInvoiceRequest {
  string customer_id = 1;
}

message CreateInvoiceResponse {
  string invoice_id = 1;
}
`

	commonProto := `syntax = "proto3";
package common.v1;

message Address {
  string street = 1;
  string city = 2;
}
`

	require.NoError(t, os.WriteFile(filepath.Join(userServiceDir, "user.proto"), []byte(userProto), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(billingServiceDir, "billing.proto"), []byte(billingProto), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(commonDir, "types.proto"), []byte(commonProto), 0644))

	// Create discovery instance
	config := DefaultMonorepoDiscoveryConfig(tmpDir)
	md := NewMonorepoDiscovery(config)

	// Discover all services
	ctx := context.Background()
	groups, err := md.DiscoverAll(ctx)
	require.NoError(t, err)

	// Verify service groups
	assert.Len(t, groups, 3, "Should discover 3 services (user-service, billing-service, common)")

	// Verify user-service group
	userGroup, exists := groups["user-service"]
	assert.True(t, exists, "user-service group should exist")
	if exists {
		assert.Equal(t, "user-service", userGroup.ServiceName)
		assert.Equal(t, 1, userGroup.TotalFiles)
		assert.Equal(t, 1, userGroup.TotalServices)
		assert.Contains(t, userGroup.GetServiceNames(), "UserService")
	}

	// Verify billing-service group
	billingGroup, exists := groups["billing-service"]
	assert.True(t, exists, "billing-service group should exist")
	if exists {
		assert.Equal(t, "billing-service", billingGroup.ServiceName)
		assert.Equal(t, 1, billingGroup.TotalFiles)
		assert.Equal(t, 1, billingGroup.TotalServices)
		assert.Contains(t, billingGroup.GetServiceNames(), "BillingService")
	}

	// Verify common group
	commonGroup, exists := groups["common"]
	assert.True(t, exists, "common group should exist")
	if exists {
		assert.Equal(t, "common", commonGroup.ServiceName)
		assert.Equal(t, 1, commonGroup.TotalFiles)
		assert.Equal(t, 0, commonGroup.TotalServices)
	}
}

func TestMonorepoDiscovery_ShouldExclude(t *testing.T) {
	config := DefaultMonorepoDiscoveryConfig("/monorepo")
	md := NewMonorepoDiscovery(config)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "vendor directory",
			filePath: "/monorepo/vendor/github.com/foo/bar.proto",
			expected: true,
		},
		{
			name:     "third_party directory",
			filePath: "/monorepo/third_party/google/protobuf/any.proto",
			expected: true,
		},
		{
			name:     "node_modules directory",
			filePath: "/monorepo/node_modules/package/file.proto",
			expected: true,
		},
		{
			name:     "test proto file",
			filePath: "/monorepo/services/user/api/user_test.proto",
			expected: true,
		},
		{
			name:     "regular proto file",
			filePath: "/monorepo/services/user/api/user.proto",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.shouldExclude(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonorepoDiscovery_ListServices(t *testing.T) {
	tmpDir := t.TempDir()

	// Create simple structure
	serviceDir := filepath.Join(tmpDir, "services", "test-service", "api")
	require.NoError(t, os.MkdirAll(serviceDir, 0755))

	protoContent := `syntax = "proto3";
package test.v1;

service TestService {
  rpc Test(TestRequest) returns (TestResponse);
}
`
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "test.proto"), []byte(protoContent), 0644))

	config := DefaultMonorepoDiscoveryConfig(tmpDir)
	md := NewMonorepoDiscovery(config)

	ctx := context.Background()
	services, err := md.ListServices(ctx)
	require.NoError(t, err)

	assert.Contains(t, services, "test-service")
}

func TestMonorepoDiscovery_GetServiceGroup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create service
	serviceDir := filepath.Join(tmpDir, "services", "user-service", "api")
	require.NoError(t, os.MkdirAll(serviceDir, 0755))

	protoContent := `syntax = "proto3";
package user.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
`
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "user.proto"), []byte(protoContent), 0644))

	config := DefaultMonorepoDiscoveryConfig(tmpDir)
	md := NewMonorepoDiscovery(config)

	ctx := context.Background()

	// Get existing service
	group, err := md.GetServiceGroup(ctx, "user-service")
	require.NoError(t, err)
	assert.Equal(t, "user-service", group.ServiceName)

	// Get non-existent service
	_, err = md.GetServiceGroup(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestServiceGroup_GetPackageNames(t *testing.T) {
	group := &ServiceGroup{
		ProtoFiles: []*ServiceProtoFile{
			{PackageName: "user.v1"},
			{PackageName: "user.v1"}, // Duplicate
			{PackageName: "user.v2"},
			{PackageName: ""},
		},
	}

	packages := group.GetPackageNames()
	assert.Len(t, packages, 2, "Should deduplicate packages")
	assert.Contains(t, packages, "user.v1")
	assert.Contains(t, packages, "user.v2")
}

func TestServiceGroup_GetServiceNames(t *testing.T) {
	group := &ServiceGroup{
		ProtoFiles: []*ServiceProtoFile{
			{Services: []string{"UserService", "AccountService"}},
			{Services: []string{"UserService"}}, // Duplicate
			{Services: []string{"ProfileService"}},
		},
	}

	services := group.GetServiceNames()
	assert.Len(t, services, 3, "Should deduplicate service names")
	assert.Contains(t, services, "UserService")
	assert.Contains(t, services, "AccountService")
	assert.Contains(t, services, "ProfileService")
}

func TestMonorepoDiscovery_FindCommonAncestor(t *testing.T) {
	md := NewMonorepoDiscovery(DefaultMonorepoDiscoveryConfig("/monorepo"))

	tests := []struct {
		name     string
		path1    string
		path2    string
		expected string
	}{
		{
			name:     "same directory",
			path1:    "/monorepo/services/user/api",
			path2:    "/monorepo/services/user/api",
			expected: "/monorepo/services/user/api",
		},
		{
			name:     "different files same directory",
			path1:    "/monorepo/services/user/api/v1",
			path2:    "/monorepo/services/user/api/v2",
			expected: "/monorepo/services/user/api",
		},
		{
			name:     "different services",
			path1:    "/monorepo/services/user/api",
			path2:    "/monorepo/services/billing/api",
			expected: "/monorepo/services",
		},
		{
			name:     "completely different paths",
			path1:    "/monorepo/services/user",
			path2:    "/monorepo/pkg/common",
			expected: "/monorepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.findCommonAncestor(tt.path1, tt.path2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
