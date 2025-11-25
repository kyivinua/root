package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtoConsolidator_ConsolidateService(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	outDir := filepath.Join(tmpDir, "out")

	// Create source proto files
	serviceDir := filepath.Join(srcDir, "services", "user-service", "api", "v1")
	require.NoError(t, os.MkdirAll(serviceDir, 0755))

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
	userProtoPath := filepath.Join(serviceDir, "user.proto")
	require.NoError(t, os.WriteFile(userProtoPath, []byte(userProto), 0644))

	// Create service group
	group := &ServiceGroup{
		ServiceName: "user-service",
		ProtoFiles: []*ServiceProtoFile{
			{
				FilePath:     userProtoPath,
				RelativePath: "services/user-service/api/v1/user.proto",
				PackageName:  "user.v1",
				Services:     []string{"UserService"},
				ServiceOwner: "user-service",
			},
		},
		PackageName:   "user.v1",
		RootPath:      serviceDir,
		TotalFiles:    1,
		TotalServices: 1,
	}

	// Create consolidator
	config := &ConsolidationConfig{
		OutputRoot:           outDir,
		CreateBufConfig:      true,
		PreserveDirStructure: true,
	}
	consolidator := NewProtoConsolidator(config)

	// Consolidate service
	ctx := context.Background()
	result, err := consolidator.ConsolidateService(ctx, group)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify result
	assert.Equal(t, "user-service", result.ServiceName)
	assert.Equal(t, filepath.Join(outDir, "user-service"), result.OutputPath)
	assert.Equal(t, 1, result.FilesCopied)
	assert.True(t, result.BufConfigCreated)
	assert.Empty(t, result.Errors)

	// Verify files were created
	consolidatedProto := filepath.Join(outDir, "user-service", "services", "user-service", "api", "v1", "user.proto")
	assert.FileExists(t, consolidatedProto)

	bufConfig := filepath.Join(outDir, "user-service", "buf.yaml")
	assert.FileExists(t, bufConfig)

	readme := filepath.Join(outDir, "user-service", "README.md")
	assert.FileExists(t, readme)

	// Verify consolidated proto content
	consolidatedContent, err := os.ReadFile(consolidatedProto)
	require.NoError(t, err)
	assert.Equal(t, userProto, string(consolidatedContent))
}

func TestProtoConsolidator_ConsolidateService_FlattenStructure(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	outDir := filepath.Join(tmpDir, "out")

	// Create multiple proto files
	dir1 := filepath.Join(srcDir, "api", "v1")
	dir2 := filepath.Join(srcDir, "api", "v2")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))

	proto1Path := filepath.Join(dir1, "user.proto")
	proto2Path := filepath.Join(dir2, "user.proto")
	require.NoError(t, os.WriteFile(proto1Path, []byte("syntax = \"proto3\";"), 0644))
	require.NoError(t, os.WriteFile(proto2Path, []byte("syntax = \"proto3\";"), 0644))

	group := &ServiceGroup{
		ServiceName: "test-service",
		ProtoFiles: []*ServiceProtoFile{
			{
				FilePath:     proto1Path,
				RelativePath: "api/v1/user.proto",
				PackageName:  "test.v1",
			},
			{
				FilePath:     proto2Path,
				RelativePath: "api/v2/user.proto",
				PackageName:  "test.v2",
			},
		},
		TotalFiles: 2,
	}

	config := &ConsolidationConfig{
		OutputRoot:       outDir,
		FlattenStructure: true,
		CreateBufConfig:  false,
	}
	consolidator := NewProtoConsolidator(config)

	ctx := context.Background()
	result, err := consolidator.ConsolidateService(ctx, group)
	require.NoError(t, err)

	assert.Equal(t, 2, result.FilesCopied)

	// With flatten, both files should be in root (but will conflict with same name)
	// Check that at least one was copied
	files, err := os.ReadDir(filepath.Join(outDir, "test-service"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1)
}

func TestProtoConsolidator_ConsolidateAll(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	outDir := filepath.Join(tmpDir, "out")

	// Create proto files for multiple services
	userServiceDir := filepath.Join(srcDir, "services", "user-service")
	billingServiceDir := filepath.Join(srcDir, "services", "billing-service")
	require.NoError(t, os.MkdirAll(userServiceDir, 0755))
	require.NoError(t, os.MkdirAll(billingServiceDir, 0755))

	userProtoPath := filepath.Join(userServiceDir, "user.proto")
	billingProtoPath := filepath.Join(billingServiceDir, "billing.proto")
	require.NoError(t, os.WriteFile(userProtoPath, []byte("syntax = \"proto3\";\npackage user.v1;"), 0644))
	require.NoError(t, os.WriteFile(billingProtoPath, []byte("syntax = \"proto3\";\npackage billing.v1;"), 0644))

	serviceGroups := map[string]*ServiceGroup{
		"user-service": {
			ServiceName: "user-service",
			ProtoFiles: []*ServiceProtoFile{
				{
					FilePath:     userProtoPath,
					RelativePath: "services/user-service/user.proto",
					PackageName:  "user.v1",
				},
			},
			TotalFiles: 1,
		},
		"billing-service": {
			ServiceName: "billing-service",
			ProtoFiles: []*ServiceProtoFile{
				{
					FilePath:     billingProtoPath,
					RelativePath: "services/billing-service/billing.proto",
					PackageName:  "billing.v1",
				},
			},
			TotalFiles: 1,
		},
	}

	config := DefaultConsolidationConfig(outDir)
	consolidator := NewProtoConsolidator(config)

	ctx := context.Background()
	results, err := consolidator.ConsolidateAll(ctx, serviceGroups)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify user-service
	userResult, exists := results["user-service"]
	assert.True(t, exists)
	assert.Equal(t, "user-service", userResult.ServiceName)
	assert.Equal(t, 1, userResult.FilesCopied)

	// Verify billing-service
	billingResult, exists := results["billing-service"]
	assert.True(t, exists)
	assert.Equal(t, "billing-service", billingResult.ServiceName)
	assert.Equal(t, 1, billingResult.FilesCopied)

	// Verify directory structure
	assert.DirExists(t, filepath.Join(outDir, "user-service"))
	assert.DirExists(t, filepath.Join(outDir, "billing-service"))
}

func TestProtoConsolidator_CreateBufConfig(t *testing.T) {
	tmpDir := t.TempDir()

	group := &ServiceGroup{
		ServiceName: "test-service",
		ProtoFiles: []*ServiceProtoFile{
			{PackageName: "test.v1"},
			{PackageName: "test.v2"},
		},
	}

	consolidator := NewProtoConsolidator(nil)
	err := consolidator.createBufConfig(group, tmpDir)
	require.NoError(t, err)

	bufConfigPath := filepath.Join(tmpDir, "buf.yaml")
	assert.FileExists(t, bufConfigPath)

	content, err := os.ReadFile(bufConfigPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "version: v1")
	assert.Contains(t, contentStr, "name: buf.build/test-service")
	assert.Contains(t, contentStr, "# - test.v1")
	assert.Contains(t, contentStr, "# - test.v2")
}

func TestProtoConsolidator_CreateServiceReadme(t *testing.T) {
	tmpDir := t.TempDir()

	group := &ServiceGroup{
		ServiceName: "user-service",
		ProtoFiles: []*ServiceProtoFile{
			{
				RelativePath: "api/v1/user.proto",
				PackageName:  "user.v1",
				Services:     []string{"UserService"},
			},
			{
				RelativePath: "api/v1/account.proto",
				PackageName:  "user.v1",
				Services:     []string{"AccountService"},
			},
		},
		TotalFiles:    2,
		TotalServices: 2,
	}

	consolidator := NewProtoConsolidator(nil)
	err := consolidator.createServiceReadme(group, tmpDir)
	require.NoError(t, err)

	readmePath := filepath.Join(tmpDir, "README.md")
	assert.FileExists(t, readmePath)

	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# user-service")
	assert.Contains(t, contentStr, "**Total Proto Files**: 2")
	assert.Contains(t, contentStr, "**Total Services**: 2")
	assert.Contains(t, contentStr, "`UserService`")
	assert.Contains(t, contentStr, "`AccountService`")
	assert.Contains(t, contentStr, "`user.v1`")
	assert.Contains(t, contentStr, "api/v1/user.proto")
	assert.Contains(t, contentStr, "api/v1/account.proto")
}

func TestProtoConsolidator_ConsolidateService_NilGroup(t *testing.T) {
	consolidator := NewProtoConsolidator(nil)
	ctx := context.Background()

	_, err := consolidator.ConsolidateService(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service group cannot be nil")
}

func TestProtoConsolidator_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	outDir := filepath.Join(tmpDir, "out")

	// Create many proto files to ensure cancellation can happen
	serviceGroups := make(map[string]*ServiceGroup)
	for i := 0; i < 10; i++ {
		serviceDir := filepath.Join(srcDir, "service", string(rune('a'+i)))
		require.NoError(t, os.MkdirAll(serviceDir, 0755))

		protoPath := filepath.Join(serviceDir, "test.proto")
		require.NoError(t, os.WriteFile(protoPath, []byte("syntax = \"proto3\";"), 0644))

		serviceName := "service-" + string(rune('a'+i))
		serviceGroups[serviceName] = &ServiceGroup{
			ServiceName: serviceName,
			ProtoFiles: []*ServiceProtoFile{
				{
					FilePath:     protoPath,
					RelativePath: "test.proto",
				},
			},
			TotalFiles: 1,
		}
	}

	config := DefaultConsolidationConfig(outDir)
	consolidator := NewProtoConsolidator(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := consolidator.ConsolidateAll(ctx, serviceGroups)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestProtoConsolidator_CleanOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "out")

	// Create some files
	require.NoError(t, os.MkdirAll(filepath.Join(outDir, "service1"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "service1", "test.proto"), []byte("test"), 0644))

	config := &ConsolidationConfig{
		OutputRoot: outDir,
	}
	consolidator := NewProtoConsolidator(config)

	// Clean
	err := consolidator.CleanOutputDirectory()
	require.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(outDir)
	assert.True(t, os.IsNotExist(err))
}

func TestProtoConsolidator_GetConsolidatedPath(t *testing.T) {
	config := &ConsolidationConfig{
		OutputRoot: "/tmp/consolidated",
	}
	consolidator := NewProtoConsolidator(config)

	path := consolidator.GetConsolidatedPath("user-service")
	assert.Equal(t, "/tmp/consolidated/user-service", path)
}

func TestNewProtoConsolidator_DefaultConfig(t *testing.T) {
	consolidator := NewProtoConsolidator(nil)
	assert.NotNil(t, consolidator)
	assert.NotNil(t, consolidator.config)
	assert.Equal(t, "./consolidated-protos", consolidator.config.OutputRoot)
	assert.True(t, consolidator.config.CreateBufConfig)
	assert.False(t, consolidator.config.CopyDependencies)
	assert.True(t, consolidator.config.PreserveDirStructure)
	assert.False(t, consolidator.config.FlattenStructure)
}

func TestDefaultConsolidationConfig(t *testing.T) {
	config := DefaultConsolidationConfig("/custom/path")
	assert.NotNil(t, config)
	assert.Equal(t, "/custom/path", config.OutputRoot)
	assert.True(t, config.CreateBufConfig)
	assert.False(t, config.CopyDependencies)
	assert.True(t, config.PreserveDirStructure)
	assert.False(t, config.FlattenStructure)
}

func TestProtoConsolidator_MultiplePackages(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	outDir := filepath.Join(tmpDir, "out")

	// Create proto files with different packages
	dir1 := filepath.Join(srcDir, "v1")
	dir2 := filepath.Join(srcDir, "v2")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))

	proto1Path := filepath.Join(dir1, "user.proto")
	proto2Path := filepath.Join(dir2, "user.proto")
	require.NoError(t, os.WriteFile(proto1Path, []byte("syntax = \"proto3\";\npackage user.v1;"), 0644))
	require.NoError(t, os.WriteFile(proto2Path, []byte("syntax = \"proto3\";\npackage user.v2;"), 0644))

	group := &ServiceGroup{
		ServiceName: "user-service",
		ProtoFiles: []*ServiceProtoFile{
			{
				FilePath:     proto1Path,
				RelativePath: "v1/user.proto",
				PackageName:  "user.v1",
			},
			{
				FilePath:     proto2Path,
				RelativePath: "v2/user.proto",
				PackageName:  "user.v2",
			},
		},
		TotalFiles: 2,
	}

	config := &ConsolidationConfig{
		OutputRoot:           outDir,
		CreateBufConfig:      true,
		PreserveDirStructure: true,
	}
	consolidator := NewProtoConsolidator(config)

	ctx := context.Background()
	result, err := consolidator.ConsolidateService(ctx, group)
	require.NoError(t, err)
	assert.Equal(t, 2, result.FilesCopied)

	// Verify buf.yaml contains both packages
	bufConfigPath := filepath.Join(outDir, "user-service", "buf.yaml")
	content, err := os.ReadFile(bufConfigPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "user.v1")
	assert.Contains(t, string(content), "user.v2")
}
