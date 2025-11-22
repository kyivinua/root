package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Confluence REST API client.
type Client struct {
	baseURL    string
	username   string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Confluence API client.
func NewClient(baseURL, username, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
		return nil, fmt.Errorf("marshal page: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/rest/api/content", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var created Page
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &created, nil
}

// UpdatePage updates an existing Confluence page.
func (c *Client) UpdatePage(pageID string, page *Page) (*Page, error) {
	body, err := json.Marshal(page)
	if err != nil {
		return nil, fmt.Errorf("marshal page: %w", err)
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/rest/api/content/"+pageID, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var updated Page
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &updated, nil
}

// GetPage retrieves a page by ID.
func (c *Client) GetPage(pageID string) (*Page, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/rest/api/content/"+pageID+"?expand=body.storage,version", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &page, nil
}

// FindPageByTitle finds a page by title in a space.
func (c *Client) FindPageByTitle(spaceKey, title string) (*Page, error) {
	url := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&title=%s&expand=body.storage,version",
		c.baseURL, spaceKey, title)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var pageResp PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if pageResp.Size == 0 {
		return nil, nil
	}

	return &pageResp.Results[0], nil
}

// DeletePage deletes a page by ID.
func (c *Client) DeletePage(pageID string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/rest/api/content/"+pageID, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
