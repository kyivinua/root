package confluence

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		username  string
		apiToken  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid configuration",
			baseURL:  "https://company.atlassian.net/wiki",
			username: "test@example.com",
			apiToken: "valid-token-123456789012345",
			wantErr:  false,
		},
		{
			name:      "invalid URL",
			baseURL:   "not-a-url",
			username:  "test@example.com",
			apiToken:  "valid-token-123456789012345",
			wantErr:   true,
			errSubstr: "invalid base URL",
		},
		{
			name:      "empty username",
			baseURL:   "https://company.atlassian.net/wiki",
			username:  "",
			apiToken:  "valid-token-123456789012345",
			wantErr:   true,
			errSubstr: "username cannot be empty",
		},
		{
			name:      "empty API token",
			baseURL:   "https://company.atlassian.net/wiki",
			username:  "test@example.com",
			apiToken:  "",
			wantErr:   true,
			errSubstr: "invalid API token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.username, tt.apiToken)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.baseURL, client.baseURL)
				assert.Equal(t, tt.username, client.username)
				assert.Equal(t, tt.apiToken, client.apiToken)
				assert.NotNil(t, client.httpClient)
				assert.NotNil(t, client.limiter)
			}
		})
	}
}

func TestClient_CreatePage(t *testing.T) {
	tests := []struct {
		name       string
		page       *Page
		statusCode int
		response   string
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "successful creation",
			page: &Page{
				Type:  "page",
				Title: "Test Page",
				Space: Space{Key: "TEST"},
				Body:  Body{Storage: Storage{Value: "<p>Content</p>", Representation: "storage"}},
			},
			statusCode: http.StatusCreated,
			response: `{
				"id": "12345",
				"type": "page",
				"title": "Test Page",
				"space": {"key": "TEST"},
				"version": {"number": 1}
			}`,
			wantErr: false,
		},
		{
			name: "unauthorized",
			page: &Page{
				Type:  "page",
				Title: "Test Page",
				Space: Space{Key: "TEST"},
				Body:  Body{Storage: Storage{Value: "<p>Content</p>", Representation: "storage"}},
			},
			statusCode: http.StatusUnauthorized,
			response:   `{"message": "Invalid credentials"}`,
			wantErr:    true,
			errSubstr:  "authentication failed",
		},
		{
			name: "conflict",
			page: &Page{
				Type:  "page",
				Title: "Test Page",
				Space: Space{Key: "TEST"},
				Body:  Body{Storage: Storage{Value: "<p>Content</p>", Representation: "storage"}},
			},
			statusCode: http.StatusConflict,
			response:   `{"message": "Page already exists"}`,
			wantErr:    true,
			errSubstr:  "conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/rest/api/content", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
			require.NoError(t, err)

			created, err := client.CreatePage(tt.page)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				assert.Nil(t, created)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, created)
				assert.Equal(t, "12345", created.ID)
				assert.Equal(t, "Test Page", created.Title)
			}
		})
	}
}

func TestClient_UpdatePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/content/12345")

		w.WriteHeader(http.StatusOK)
		resp := Page{
			ID:      "12345",
			Type:    "page",
			Title:   "Updated Page",
			Version: Version{Number: 2},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
	require.NoError(t, err)

	page := &Page{
		ID:      "12345",
		Type:    "page",
		Title:   "Updated Page",
		Space:   Space{Key: "TEST"},
		Body:    Body{Storage: Storage{Value: "<p>New content</p>", Representation: "storage"}},
		Version: Version{Number: 2},
	}

	updated, err := client.UpdatePage("12345", page)
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "12345", updated.ID)
	assert.Equal(t, "Updated Page", updated.Title)
	assert.Equal(t, 2, updated.Version.Number)
}

func TestClient_GetPage(t *testing.T) {
	tests := []struct {
		name       string
		pageID     string
		statusCode int
		response   string
		wantPage   bool
		wantErr    bool
	}{
		{
			name:       "page found",
			pageID:     "12345",
			statusCode: http.StatusOK,
			response: `{
				"id": "12345",
				"title": "Test Page",
				"version": {"number": 1},
				"body": {
					"storage": {
						"value": "<p>Content</p>",
						"representation": "storage"
					}
				}
			}`,
			wantPage: true,
			wantErr:  false,
		},
		{
			name:       "page not found",
			pageID:     "99999",
			statusCode: http.StatusNotFound,
			response:   `{"message": "Page not found"}`,
			wantPage:   false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/content/"+tt.pageID)

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
			require.NoError(t, err)

			page, err := client.GetPage(tt.pageID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.wantPage {
					assert.NotNil(t, page)
					assert.Equal(t, tt.pageID, page.ID)
				} else {
					assert.Nil(t, page)
				}
			}
		})
	}
}

func TestClient_FindPageByTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "spaceKey=TEST")
		assert.Contains(t, r.URL.RawQuery, "title=Test+Page")

		w.WriteHeader(http.StatusOK)
		resp := PageResponse{
			Results: []Page{
				{
					ID:    "12345",
					Title: "Test Page",
				},
			},
			Size: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
	require.NoError(t, err)

	page, err := client.FindPageByTitle("TEST", "Test Page")
	require.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, "12345", page.ID)
	assert.Equal(t, "Test Page", page.Title)
}

func TestClient_DeletePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/content/12345")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
	require.NoError(t, err)

	err = client.DeletePage("12345")
	require.NoError(t, err)
}

func TestClient_UploadAttachment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/content/12345/child/attachment")
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		w.WriteHeader(http.StatusOK)
		resp := AttachmentResults{
			Results: []Attachment{
				{
					ID:    "att-001",
					Title: "diagram.png",
					Type:  "attachment",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
	require.NoError(t, err)

	content := []byte("fake image content")
	attachment, err := client.UploadAttachment("12345", "diagram.png", content, "Test diagram")
	require.NoError(t, err)
	assert.NotNil(t, attachment)
	assert.Equal(t, "att-001", attachment.ID)
	assert.Equal(t, "diagram.png", attachment.Title)
}

func TestClient_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		resp := Page{ID: fmt.Sprintf("page-%d", requestCount)}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "valid-token-123456789012345")
	require.NoError(t, err)

	// Make 12 rapid requests (rate limit is 10/second)
	start := time.Now()
	for i := 0; i < 12; i++ {
		_, _ = client.GetPage(fmt.Sprintf("page-%d", i))
	}
	elapsed := time.Since(start)

	// Should take at least 1 second due to rate limiting
	assert.GreaterOrEqual(t, elapsed, time.Second)
	assert.Equal(t, 12, requestCount)
}
