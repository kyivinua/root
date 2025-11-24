package confluence

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetSpaceInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/space/TEST", r.URL.Path)

		response := SpaceInfo{
			Key:         "TEST",
			Name:        "Test Space",
			Description: "A test space",
			Type:        "global",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	ctx := context.Background()
	spaceInfo, err := client.GetSpaceInfo(ctx, "TEST")
	require.NoError(t, err)

	assert.Equal(t, "TEST", spaceInfo.Key)
	assert.Equal(t, "Test Space", spaceInfo.Name)
	assert.Equal(t, "A test space", spaceInfo.Description)
	assert.Equal(t, "global", spaceInfo.Type)
}

func TestClient_GetAllPagesInSpace(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/content", r.URL.Path)
		assert.Equal(t, "TEST", r.URL.Query().Get("spaceKey"))

		start := r.URL.Query().Get("start")

		var pages []Page
		if start == "0" {
			// First batch
			pages = []Page{
				{ID: "1", Title: "Page 1", Type: "page"},
				{ID: "2", Title: "Page 2", Type: "page"},
			}
			pageCount = 2
		} else {
			// Second batch (empty)
			pages = []Page{}
		}

		response := PageResponse{
			Results: pages,
			Size:    len(pages),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	ctx := context.Background()
	pages, err := client.GetAllPagesInSpace(ctx, "TEST")
	require.NoError(t, err)

	assert.Equal(t, pageCount, len(pages))
	assert.Equal(t, "1", pages[0].ID)
	assert.Equal(t, "Page 1", pages[0].Title)
	assert.Equal(t, "2", pages[1].ID)
	assert.Equal(t, "Page 2", pages[1].Title)
}

func TestClient_GetAllPagesInSpace_Pagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		limit := r.URL.Query().Get("limit")

		assert.Equal(t, "100", limit)

		var pages []Page
		if start == "0" {
			// First batch (full)
			for i := 0; i < 100; i++ {
				pages = append(pages, Page{
					ID:    string(rune('A' + i)),
					Title: "Page " + string(rune('A'+i)),
					Type:  "page",
				})
			}
		} else if start == "100" {
			// Second batch (partial)
			for i := 0; i < 50; i++ {
				pages = append(pages, Page{
					ID:    string(rune('a' + i)),
					Title: "Page " + string(rune('a'+i)),
					Type:  "page",
				})
			}
		}

		response := PageResponse{
			Results: pages,
			Size:    len(pages),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	ctx := context.Background()
	pages, err := client.GetAllPagesInSpace(ctx, "TEST")
	require.NoError(t, err)

	assert.Equal(t, 150, len(pages))
}

func TestClient_ExportSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/api/space/TEST":
			// Space info request
			response := SpaceInfo{
				Key:         "TEST",
				Name:        "Test Space",
				Description: "Test description",
				Type:        "global",
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/rest/api/content":
			// Pages list request
			response := PageResponse{
				Results: []Page{
					{
						ID:    "123",
						Title: "Test Page",
						Type:  "page",
						Body: Body{
							Storage: Storage{
								Value:          "<p>Test content</p>",
								Representation: "storage",
							},
						},
						Version: Version{Number: 1},
						Space:   Space{Key: "TEST"},
					},
				},
				Size: 1,
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/rest/api/content/123/child/attachment":
			// Attachments request
			response := AttachmentResults{
				Results: []Attachment{},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	// Create temporary output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.zip")

	ctx := context.Background()
	config := &SpaceExportConfig{
		SpaceKey:         "TEST",
		OutputPath:       outputPath,
		IncludeAttachments: false,
		MaxConcurrency:   5,
		Timeout:          5 * time.Minute,
	}

	result, err := client.ExportSpace(ctx, config)
	require.NoError(t, err)

	assert.Equal(t, "TEST", result.SpaceKey)
	assert.Equal(t, 1, result.PagesExported)
	assert.Equal(t, outputPath, result.ExportPath)
	assert.Greater(t, result.TotalSize, int64(0))
	assert.True(t, result.Duration > 0)

	// Verify ZIP file exists and contains expected files
	assert.FileExists(t, outputPath)

	// Open and verify ZIP contents
	zipReader, err := zip.OpenReader(outputPath)
	require.NoError(t, err)
	defer func() { _ = zipReader.Close() }()

	foundMetadata := false
	foundPageMetadata := false
	foundPageContent := false

	for _, file := range zipReader.File {
		switch file.Name {
		case "space-metadata.json":
			foundMetadata = true
			// Verify space metadata
			rc, err := file.Open()
			require.NoError(t, err)
			var metadata SpaceMetadata
			err = json.NewDecoder(rc).Decode(&metadata)
			_ = rc.Close()
			require.NoError(t, err)
			assert.Equal(t, "TEST", metadata.Key)
			assert.Equal(t, "Test Space", metadata.Name)
			assert.Equal(t, 1, metadata.Pages)

		case "pages/123/metadata.json":
			foundPageMetadata = true

		case "pages/123/content.html":
			foundPageContent = true
			// Verify page content
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			_ = rc.Close()
			require.NoError(t, err)
			assert.Equal(t, "<p>Test content</p>", string(content))
		}
	}

	assert.True(t, foundMetadata, "space-metadata.json not found in ZIP")
	assert.True(t, foundPageMetadata, "page metadata not found in ZIP")
	assert.True(t, foundPageContent, "page content not found in ZIP")
}

func TestClient_ExportSpace_WithAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/api/space/TEST":
			response := SpaceInfo{
				Key:  "TEST",
				Name: "Test Space",
				Type: "global",
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/rest/api/content":
			response := PageResponse{
				Results: []Page{
					{
						ID:    "123",
						Title: "Test Page",
						Type:  "page",
						Body: Body{
							Storage: Storage{
								Value:          "<p>Test content</p>",
								Representation: "storage",
							},
						},
						Space: Space{Key: "TEST"},
					},
				},
				Size: 1,
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/rest/api/content/123/child/attachment":
			response := AttachmentResults{
				Results: []Attachment{
					{
						ID:    "att1",
						Title: "test.txt",
						Links: AttachmentLinks{
							Download: "/download/attachments/123/test.txt",
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)

		case r.URL.Path == "/download/attachments/123/test.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Test attachment content"))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.zip")

	ctx := context.Background()
	config := &SpaceExportConfig{
		SpaceKey:         "TEST",
		OutputPath:       outputPath,
		IncludeAttachments: true,
		MaxConcurrency:   5,
	}

	result, err := client.ExportSpace(ctx, config)
	require.NoError(t, err)

	assert.Equal(t, 1, result.AttachmentsExported)

	// Verify attachment in ZIP
	zipReader, err := zip.OpenReader(outputPath)
	require.NoError(t, err)
	defer func() { _ = zipReader.Close() }()

	foundAttachment := false
	for _, file := range zipReader.File {
		if file.Name == "pages/123/attachments/test.txt" {
			foundAttachment = true
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			_ = rc.Close()
			require.NoError(t, err)
			assert.Equal(t, "Test attachment content", string(content))
		}
	}

	assert.True(t, foundAttachment, "attachment not found in ZIP")
}

func TestClient_ExportSpace_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SpaceInfo{Key: "TEST", Name: "Test"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testtokentesttokentesttoken")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.zip")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := &SpaceExportConfig{
		SpaceKey:   "TEST",
		OutputPath: outputPath,
	}

	_, err = client.ExportSpace(ctx, config)
	assert.Error(t, err)
	// Context cancellation error should be wrapped
	assert.Contains(t, err.Error(), "context canceled")
}

func TestClient_ExportSpace_InvalidConfig(t *testing.T) {
	client, err := NewClient("http://localhost", "user", "testtokentesttokentesttoken")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.ExportSpace(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestClient_writePageToZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)

	client, _ := NewClient("http://localhost", "user", "testtokentesttokentesttoken")

	page := &Page{
		ID:     "123",
		Title:  "Test Page",
		Type:   "page",
		Status: "current",
		Body: Body{
			Storage: Storage{
				Value:          "<p>Test content</p>",
				Representation: "storage",
			},
		},
		Version: Version{Number: 1},
		Space:   Space{Key: "TEST"},
	}

	err = client.writePageToZip(zipWriter, page)
	require.NoError(t, err)

	err = zipWriter.Close()
	require.NoError(t, err)
	_ = zipFile.Close()

	// Verify ZIP contents
	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)
	defer func() { _ = zipReader.Close() }()

	foundMetadata := false
	foundContent := false

	for _, file := range zipReader.File {
		if file.Name == "pages/123/metadata.json" {
			foundMetadata = true
		}
		if file.Name == "pages/123/content.html" {
			foundContent = true
		}
	}

	assert.True(t, foundMetadata)
	assert.True(t, foundContent)
}
