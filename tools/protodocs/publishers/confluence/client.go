package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/ratelimit"
	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/validation"
)

const (
	maxRetries     = 3
	retryWaitTime  = 2 * time.Second
	retryBackoff   = 2 // exponential backoff multiplier
)

// Client is a Confluence REST API client.
type Client struct {
	baseURL    string
	username   string
	apiToken   string
	httpClient *http.Client
	limiter    *ratelimit.Limiter
}

// NewClient creates a new Confluence API client.
func NewClient(baseURL, username, apiToken string) (*Client, error) {
	// Validate URL
	if err := validation.ValidateURL(baseURL); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid base URL")
	}

	// Validate API token
	if err := validation.ValidateAPIKey(apiToken); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid API token")
	}

	// Validate username (basic email/username format)
	if username == "" {
		return nil, errors.New(errors.ErrorTypeValidation, "username cannot be empty")
	}

	// Create rate limiter (10 requests per second per Confluence API limits)
	limiter := ratelimit.NewLimiter(10, time.Second)

	return &Client{
		baseURL:  baseURL,
		username: username,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: limiter,
	}, nil
}

// Page represents a Confluence page.
type Page struct {
	ID      string  `json:"id,omitempty"`
	Type    string  `json:"type"`
	Status  string  `json:"status,omitempty"`
	Title   string  `json:"title"`
	Space   Space   `json:"space"`
	Body    Body    `json:"body"`
	Version Version `json:"version,omitempty"`

	// For creating child pages
	Ancestors []Ancestor `json:"ancestors,omitempty"`
}

// Space represents a Confluence space.
type Space struct {
	Key string `json:"key"`
}

// Body represents page content.
type Body struct {
	Storage Storage `json:"storage"`
}

// Storage represents the storage format content.
type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version represents page version information.
type Version struct {
	Number  int    `json:"number"`
	Message string `json:"message,omitempty"`
}

// Ancestor represents a parent page.
type Ancestor struct {
	ID string `json:"id"`
}

// PageResponse represents the API response when fetching pages.
type PageResponse struct {
	Results []Page `json:"results"`
	Size    int    `json:"size"`
}

// CreatePage creates a new Confluence page.
func (c *Client) CreatePage(page *Page) (*Page, error) {
	body, err := json.Marshal(page)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal page")
	}

	req, err := http.NewRequest("POST", c.baseURL+"/rest/api/content", bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// Categorize by status code
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return nil, errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("resource not found: %s", string(bodyBytes)))
		case http.StatusConflict:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("conflict: %s", string(bodyBytes)))
		case http.StatusTooManyRequests:
			return nil, errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("rate limited: %s", string(bodyBytes)))
		case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
			return nil, errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var created Page
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	return &created, nil
}

// UpdatePage updates an existing Confluence page.
func (c *Client) UpdatePage(pageID string, page *Page) (*Page, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	body, err := json.Marshal(page)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal page")
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/rest/api/content/"+pageID, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return nil, errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
		case http.StatusConflict:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("conflict (version mismatch): %s", string(bodyBytes)))
		case http.StatusTooManyRequests:
			return nil, errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("rate limited: %s", string(bodyBytes)))
		case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
			return nil, errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var updated Page
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	return &updated, nil
}

// GetPage retrieves a page by ID.
func (c *Client) GetPage(pageID string) (*Page, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	req, err := http.NewRequest("GET", c.baseURL+"/rest/api/content/"+pageID+"?expand=body.storage,version", nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	return &page, nil
}

// FindPageByTitle finds a page by title in a space.
func (c *Client) FindPageByTitle(spaceKey, title string) (*Page, error) {
	// Validate space key
	if err := validation.ValidateSpaceKey(spaceKey); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid space key")
	}

	// Sanitize title
	title = validation.SanitizeString(title)
	if title == "" {
		return nil, errors.New(errors.ErrorTypeValidation, "title cannot be empty")
	}

	url := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&title=%s&expand=body.storage,version",
		c.baseURL, spaceKey, title)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var pageResp PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	if pageResp.Size == 0 {
		return nil, nil
	}

	return &pageResp.Results[0], nil
}

// DeletePage deletes a page by ID.
func (c *Client) DeletePage(pageID string) error {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	req, err := http.NewRequest("DELETE", c.baseURL+"/rest/api/content/"+pageID, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
		default:
			return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	return nil
}

// doRequestWithRetry executes an HTTP request with retry logic for transient failures.
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	waitTime := retryWaitTime

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			time.Sleep(waitTime)
			waitTime *= retryBackoff
		}

		// Clone request for retry (body may be consumed)
		reqCopy := req.Clone(req.Context())

		resp, err = c.httpClient.Do(reqCopy)
		if err == nil {
			// Check for retryable status codes
			if resp.StatusCode < 500 && resp.StatusCode != 429 {
				// Success or client error (not retryable)
				return resp, nil
			}

			// Server error or rate limit - retry
			if attempt < maxRetries {
				resp.Body.Close()
				continue
			}
		}

		// Network error - retry
		if attempt < maxRetries {
			continue
		}
	}

	return resp, err
}

// Attachment represents a Confluence attachment.
type Attachment struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Metadata AttachmentMetadata `json:"metadata,omitempty"`
}

// AttachmentMetadata represents attachment metadata.
type AttachmentMetadata struct {
	MediaType string `json:"mediaType,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// AttachmentResults represents the response from attachment upload.
type AttachmentResults struct {
	Results []Attachment `json:"results"`
}

// UploadAttachment uploads a file attachment to a Confluence page.
func (c *Client) UploadAttachment(pageID string, filename string, content []byte, comment string) (*Attachment, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create form file")
	}

	if _, err := part.Write(content); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write file content")
	}

	// Add comment if provided
	if comment != "" {
		if err := writer.WriteField("comment", comment); err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write comment field")
		}
	}

	if err := writer.Close(); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to close writer")
	}

	// Create request
	url := fmt.Sprintf("%s/rest/api/content/%s/child/attachment", c.baseURL, pageID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "no-check")

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return nil, errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
		case http.StatusRequestEntityTooLarge:
			return nil, errors.New(errors.ErrorTypeValidation, "file size exceeds limit")
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var results AttachmentResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	if len(results.Results) == 0 {
		return nil, errors.New(errors.ErrorTypeInternal, "no attachment returned in response")
	}

	return &results.Results[0], nil
}

// UpdateAttachment updates an existing attachment with new content.
func (c *Client) UpdateAttachment(pageID, attachmentID string, filename string, content []byte, comment string) (*Attachment, error) {
	// Validate IDs
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}
	if err := validation.ValidatePageID(attachmentID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid attachment ID")
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create form file")
	}

	if _, err := part.Write(content); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write file content")
	}

	// Add comment if provided
	if comment != "" {
		if err := writer.WriteField("comment", comment); err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write comment field")
		}
	}

	if err := writer.Close(); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to close writer")
	}

	// Create request
	url := fmt.Sprintf("%s/rest/api/content/%s/child/attachment/%s/data", c.baseURL, pageID, attachmentID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "no-check")

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return nil, errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("attachment not found: %s", string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var results AttachmentResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	if len(results.Results) == 0 {
		return nil, errors.New(errors.ErrorTypeInternal, "no attachment returned in response")
	}

	return &results.Results[0], nil
}

// GetAttachments retrieves all attachments for a page.
func (c *Client) GetAttachments(pageID string) ([]Attachment, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/child/attachment", c.baseURL, pageID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.apiToken)

	// Apply rate limiting
	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
		case http.StatusForbidden:
			return nil, errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
		case http.StatusNotFound:
			return nil, errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
		default:
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
		}
	}

	var results AttachmentResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	return results.Results, nil
}
