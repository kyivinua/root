package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublisher(t *testing.T) {
	tests := []struct {
		name    string
		config  *PublisherConfig
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &PublisherConfig{
				BaseURL:  "https://test.atlassian.net/wiki",
				Username: "test@example.com",
				APIToken: "valid-token-123456789012345",
				SpaceKey: "TEST",
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			config: &PublisherConfig{
				BaseURL:  "invalid-url",
				Username: "test@example.com",
				APIToken: "valid-token-123456789012345",
				SpaceKey: "TEST",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher, err := NewPublisher(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, publisher)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, publisher)
				assert.NotNil(t, publisher.client)
				assert.NotNil(t, publisher.formatter)
				assert.NotNil(t, publisher.logger)
			}
		})
	}
}

func TestPublisher_ExtractTitle(t *testing.T) {
	config := &PublisherConfig{
		BaseURL:  "https://test.atlassian.net/wiki",
		Username: "test@example.com",
		APIToken: "valid-token-123456789012345",
		SpaceKey: "TEST",
	}
	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		markdown string
		filename string
		expected string
	}{
		{
			name:     "extract from H1",
			markdown: "# User Service\n\nContent here",
			filename: "/path/to/file.md",
			expected: "User Service",
		},
		{
			name:     "extract from H1 with emoji",
			markdown: "# 📚 Documentation\n\nContent",
			filename: "/path/to/file.md",
			expected: "Documentation",
		},
		{
			name:     "fallback to filename",
			markdown: "Some content without H1",
			filename: "/path/to/UserService.md",
			expected: "UserService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := publisher.extractTitle(tt.markdown, tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPublisher_ExtractServiceName(t *testing.T) {
	config := &PublisherConfig{
		BaseURL:  "https://test.atlassian.net/wiki",
		Username: "test@example.com",
		APIToken: "valid-token-123456789012345",
		SpaceKey: "TEST",
	}
	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "simple service name",
			filePath: "/docs/UserService.md",
			expected: "User",
		},
		{
			name:     "service with API suffix",
			filePath: "/docs/PaymentAPI.md",
			expected: "Payment",
		},
		{
			name:     "nested path",
			filePath: "/project/docs/AuthService.md",
			expected: "Auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := publisher.extractServiceName(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPublisher_PublishFromMarkdownFiles(t *testing.T) {
	// Create test server
	pageID := "12345"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Search for existing page
		case r.Method == "GET" && r.URL.Path == "/rest/api/content":
			w.WriteHeader(http.StatusOK)
			resp := PageResponse{
				Results: []Page{},
				Size:    0,
			}
			_ = json.NewEncoder(w).Encode(resp)

		// Create new page
		case r.Method == "POST" && r.URL.Path == "/rest/api/content":
			w.WriteHeader(http.StatusCreated)
			resp := Page{
				ID:    pageID,
				Title: "Test Page",
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create temporary directory with markdown files
	tmpDir, err := os.MkdirTemp("", "confluence-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test markdown files
	testFiles := map[string]string{
		"UserService.md": "# User Service\n\nThis is the user service documentation.",
		"AuthService.md": "# Auth Service\n\nThis is the auth service documentation.",
		"README.md":      "# README\n\nThis should be skipped.",
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create publisher
	config := &PublisherConfig{
		BaseURL:              server.URL,
		Username:             "test@example.com",
		APIToken:             "valid-token-123456789012345",
		SpaceKey:             "TEST",
		CreatePagePerService: true,
		UpdateExisting:       false,
	}

	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	// Test publishing
	result, err := publisher.PublishFromMarkdownFiles(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should create 2 pages (UserService and AuthService), skip README
	assert.Equal(t, 2, result.PagesCreated)
	assert.Equal(t, 0, result.PagesUpdated)
	assert.Len(t, result.PageURLs, 2)
	assert.Len(t, result.PageIDs, 2)
}

func TestPublisher_PublishConsolidatedPage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Search for existing page
		case r.Method == "GET" && r.URL.Path == "/rest/api/content":
			w.WriteHeader(http.StatusOK)
			resp := PageResponse{
				Results: []Page{},
				Size:    0,
			}
			_ = json.NewEncoder(w).Encode(resp)

		// Create new page
		case r.Method == "POST" && r.URL.Path == "/rest/api/content":
			w.WriteHeader(http.StatusCreated)
			resp := Page{
				ID:    "consolidated-123",
				Title: "All Services",
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create temporary directory with markdown files
	tmpDir, err := os.MkdirTemp("", "confluence-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test markdown files
	testFiles := map[string]string{
		"Service1.md": "# Service 1\n\nService 1 content.",
		"Service2.md": "# Service 2\n\nService 2 content.",
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create publisher
	config := &PublisherConfig{
		BaseURL:        server.URL,
		Username:       "test@example.com",
		APIToken:       "valid-token-123456789012345",
		SpaceKey:       "TEST",
		IncludeTOC:     true,
		UpdateExisting: false,
	}

	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	// Test consolidated publishing
	result, err := publisher.PublishConsolidatedPage(tmpDir, "All Services")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should create 1 consolidated page
	assert.Equal(t, 1, result.PagesCreated)
	assert.Equal(t, 0, result.PagesUpdated)
	assert.Len(t, result.PageURLs, 1)
	assert.Contains(t, result.PageMap, "consolidated")
}

func TestPublisher_UpdateExistingPage(t *testing.T) {
	existingPageID := "123456789"

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Search returns existing page
		case r.Method == "GET" && r.URL.Path == "/rest/api/content":
			w.WriteHeader(http.StatusOK)
			resp := PageResponse{
				Results: []Page{
					{
						ID:      existingPageID,
						Title:   "Existing Page",
						Version: Version{Number: 1},
					},
				},
				Size: 1,
			}
			_ = json.NewEncoder(w).Encode(resp)

		// Update existing page
		case r.Method == "PUT" && r.URL.Path == "/rest/api/content/"+existingPageID:
			w.WriteHeader(http.StatusOK)
			resp := Page{
				ID:      existingPageID,
				Title:   "Existing Page",
				Version: Version{Number: 2},
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create temporary directory with markdown file
	tmpDir, err := os.MkdirTemp("", "confluence-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(
		filepath.Join(tmpDir, "ExistingPage.md"),
		[]byte("# Existing Page\n\nUpdated content."),
		0644,
	)
	require.NoError(t, err)

	// Create publisher with update enabled
	config := &PublisherConfig{
		BaseURL:        server.URL,
		Username:       "test@example.com",
		APIToken:       "valid-token-123456789012345",
		SpaceKey:       "TEST",
		UpdateExisting: true,
		VersionLabel:   "Auto-updated",
	}

	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	// Test publishing with update
	result, err := publisher.PublishFromMarkdownFiles(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should update 1 page, create 0
	assert.Equal(t, 0, result.PagesCreated)
	assert.Equal(t, 1, result.PagesUpdated)
	assert.Len(t, result.PageURLs, 1)
	assert.Contains(t, result.PageURLs[0], existingPageID)
}

func TestPublisher_WithTitlePrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			resp := PageResponse{Results: []Page{}, Size: 0}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.Method == "POST" {
			// Verify title has prefix
			var page Page
			_ = json.NewDecoder(r.Body).Decode(&page)
			assert.Contains(t, page.Title, "API Doc -")

			w.WriteHeader(http.StatusCreated)
			resp := Page{ID: "123", Title: page.Title}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "confluence-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(
		filepath.Join(tmpDir, "Service.md"),
		[]byte("# Service\n\nContent."),
		0644,
	)
	require.NoError(t, err)

	config := &PublisherConfig{
		BaseURL:         server.URL,
		Username:        "test@example.com",
		APIToken:        "valid-token-123456789012345",
		SpaceKey:        "TEST",
		PageTitlePrefix: "API Doc -",
	}

	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	result, err := publisher.PublishFromMarkdownFiles(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 1, result.PagesCreated)
}

func TestPublisher_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("CONFLUENCE_USERNAME", "env-user@example.com")
	os.Setenv("CONFLUENCE_API_TOKEN", "env-token-456-valid-long-token")
	defer func() {
		os.Unsetenv("CONFLUENCE_USERNAME")
		os.Unsetenv("CONFLUENCE_API_TOKEN")
	}()

	config := &PublisherConfig{
		BaseURL:  "https://test.atlassian.net/wiki",
		Username: "original-user@example.com",           // Should be overridden
		APIToken: "original-token-valid-length-token", // Should be overridden
		SpaceKey: "TEST",
	}

	publisher, err := NewPublisher(config)
	require.NoError(t, err)

	// Environment variables should have been applied
	assert.Equal(t, "env-user@example.com", publisher.config.Username)
	assert.Equal(t, "env-token-456-valid-long-token", publisher.config.APIToken)
}
