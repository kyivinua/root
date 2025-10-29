package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
version: "1.0"
project_name: "test-project"
output_dir: "./output"
services:
  - name: "TestService"
    proto_files:
      - "./proto/test.proto"
    enabled: true
quality:
  coverage_target: 90
  description_quality_target: 85
enricher:
  provider: "claude"
  api_key: "test-key"
  enabled: false
diagrams:
  enabled: true
logging:
  level: "debug"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	config, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify loaded values
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, "test-project", config.ProjectName)
	assert.Equal(t, "./output", config.OutputDir)
	assert.Len(t, config.Services, 1)
	assert.Equal(t, "TestService", config.Services[0].Name)
	assert.Equal(t, 90.0, config.Quality.CoverageTarget)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				ProjectName: "test",
				OutputDir:   "./output",
				Services: []ServiceConfig{
					{
						Name:       "TestService",
						ProtoFiles: []string{"test.proto"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			config: Config{
				OutputDir: "./output",
				Services: []ServiceConfig{
					{Name: "TestService", ProtoFiles: []string{"test.proto"}},
				},
			},
			wantErr: true,
		},
		{
			name: "missing output dir",
			config: Config{
				ProjectName: "test",
				Services: []ServiceConfig{
					{Name: "TestService", ProtoFiles: []string{"test.proto"}},
				},
			},
			wantErr: true,
		},
		{
			name: "no services",
			config: Config{
				ProjectName: "test",
				OutputDir:   "./output",
				Services:    []ServiceConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	config := &Config{
		ProjectName: "test",
		OutputDir:   "./output",
		Services: []ServiceConfig{
			{Name: "TestService", ProtoFiles: []string{"test.proto"}},
		},
	}

	applyDefaults(config)

	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, 80.0, config.Quality.CoverageTarget)
	assert.Equal(t, 85.0, config.Quality.DescriptionQualityTarget)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
}

func TestGetServiceByName(t *testing.T) {
	config := &Config{
		Services: []ServiceConfig{
			{Name: "Service1", ProtoFiles: []string{"s1.proto"}},
			{Name: "Service2", ProtoFiles: []string{"s2.proto"}},
		},
	}

	svc := config.GetServiceByName("Service1")
	assert.NotNil(t, svc)
	assert.Equal(t, "Service1", svc.Name)

	svc = config.GetServiceByName("NonExistent")
	assert.Nil(t, svc)
}

func TestGetEnabledServices(t *testing.T) {
	config := &Config{
		Services: []ServiceConfig{
			{Name: "Service1", Enabled: true, ProtoFiles: []string{"s1.proto"}},
			{Name: "Service2", Enabled: false, ProtoFiles: []string{"s2.proto"}},
			{Name: "Service3", Enabled: true, ProtoFiles: []string{"s3.proto"}},
		},
	}

	enabled := config.GetEnabledServices()
	assert.Len(t, enabled, 2)
	assert.Equal(t, "Service1", enabled[0].Name)
	assert.Equal(t, "Service3", enabled[1].Name)
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := InitConfig(configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and validate the created config
	config, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "my-project", config.ProjectName)
}
