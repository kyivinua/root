package confluence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/validation"
)

// Label represents a Confluence label
type Label struct {
	Prefix string `json:"prefix,omitempty"`
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
}

// LabelResults represents the response from label operations
type LabelResults struct {
	Results []Label `json:"results"`
	Size    int     `json:"size"`
}

// AddLabels adds labels to a Confluence page
func (c *Client) AddLabels(ctx context.Context, pageID string, labels []string) error {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	if len(labels) == 0 {
		return nil
	}

	// Convert labels to API format
	labelObjects := make([]Label, len(labels))
	for i, label := range labels {
		labelObjects[i] = Label{
			Name:   label,
			Prefix: "global",
		}
	}

	body, err := json.Marshal(labelObjects)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal labels")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/label", c.baseURL, pageID)

	var result error
	err = WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		return nil
	})

	if err != nil {
		result = err
	}

	return result
}

// GetLabels retrieves all labels for a page
func (c *Client) GetLabels(ctx context.Context, pageID string) ([]Label, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/label", c.baseURL, pageID)

	var labels []Label
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		var results LabelResults
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
		}

		labels = results.Results
		return nil
	})

	return labels, err
}

// RemoveLabel removes a label from a page
func (c *Client) RemoveLabel(ctx context.Context, pageID, labelName string) error {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/label/%s", c.baseURL, pageID, labelName)

	return WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("label not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		return nil
	})
}

// BatchOperation represents a batch operation result
type BatchOperation struct {
	PageID  string
	Success bool
	Error   error
}

// BatchCreatePages creates multiple pages in parallel
func (c *Client) BatchCreatePages(ctx context.Context, pages []*Page, maxConcurrency int) []BatchOperation {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	results := make([]BatchOperation, len(pages))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for i, page := range pages {
		wg.Add(1)
		go func(idx int, p *Page) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create page with retry
			created, err := c.CreatePage(p)
			results[idx] = BatchOperation{
				Success: err == nil,
				Error:   err,
			}
			if created != nil {
				results[idx].PageID = created.ID
			}
		}(i, page)
	}

	wg.Wait()
	return results
}

// BatchUpdatePages updates multiple pages in parallel
func (c *Client) BatchUpdatePages(ctx context.Context, updates map[string]*Page, maxConcurrency int) map[string]BatchOperation {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	results := make(map[string]BatchOperation)
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for pageID, page := range updates {
		wg.Add(1)
		go func(id string, p *Page) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Update page with retry
			_, err := c.UpdatePage(id, p)

			mu.Lock()
			results[id] = BatchOperation{
				PageID:  id,
				Success: err == nil,
				Error:   err,
			}
			mu.Unlock()
		}(pageID, page)
	}

	wg.Wait()
	return results
}

// ComputeContentHash computes a hash of page content for diff checking
func ComputeContentHash(content string) string {
	// Use a simple but effective hash
	// In production, use crypto/sha256
	h := 0
	for i := 0; i < len(content); i++ {
		h = 31*h + int(content[i])
	}
	return fmt.Sprintf("%x", h)
}

// PageNeedsUpdate checks if a page needs updating by comparing content hashes
func (c *Client) PageNeedsUpdate(ctx context.Context, pageID string, newContent string) (bool, error) {
	// Get existing page
	existingPage, err := c.GetPage(pageID)
	if err != nil {
		return false, err
	}

	if existingPage == nil {
		return true, nil // Page doesn't exist, needs creation
	}

	// Compare content hashes
	existingHash := ComputeContentHash(existingPage.Body.Storage.Value)
	newHash := ComputeContentHash(newContent)

	return existingHash != newHash, nil
}

// BulkLabelPages adds the same labels to multiple pages
func (c *Client) BulkLabelPages(ctx context.Context, pageIDs []string, labels []string, maxConcurrency int) []BatchOperation {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	results := make([]BatchOperation, len(pageIDs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for i, pageID := range pageIDs {
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Add labels with retry
			err := c.AddLabels(ctx, id, labels)
			results[idx] = BatchOperation{
				PageID:  id,
				Success: err == nil,
				Error:   err,
			}
		}(i, pageID)
	}

	wg.Wait()
	return results
}

// CreatePageWithContext creates a page with context support
func (c *Client) CreatePageWithContext(ctx context.Context, page *Page) (*Page, error) {
	var result *Page
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		var err error
		result, err = c.CreatePage(page)
		return err
	})
	return result, err
}

// UpdatePageWithContext updates a page with context support
func (c *Client) UpdatePageWithContext(ctx context.Context, pageID string, page *Page) (*Page, error) {
	var result *Page
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		var err error
		result, err = c.UpdatePage(pageID, page)
		return err
	})
	return result, err
}

// GetPageWithContext retrieves a page with context support
func (c *Client) GetPageWithContext(ctx context.Context, pageID string) (*Page, error) {
	var result *Page
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		var err error
		result, err = c.GetPage(pageID)
		return err
	})
	return result, err
}

// FindPageByTitleWithContext finds a page with context support
func (c *Client) FindPageByTitleWithContext(ctx context.Context, spaceKey, title string) (*Page, error) {
	var result *Page
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		var err error
		result, err = c.FindPageByTitle(spaceKey, title)
		return err
	})
	return result, err
}

// RestrictionOperation represents a type of restriction operation
type RestrictionOperation string

const (
	// RestrictionOperationRead restricts who can read/view the page
	RestrictionOperationRead RestrictionOperation = "read"
	// RestrictionOperationUpdate restricts who can edit the page
	RestrictionOperationUpdate RestrictionOperation = "update"
)

// RestrictionType represents the type of restriction subject
type RestrictionType string

const (
	// RestrictionTypeUser restricts by specific user
	RestrictionTypeUser RestrictionType = "user"
	// RestrictionTypeGroup restricts by user group
	RestrictionTypeGroup RestrictionType = "group"
)

// PageRestriction represents a single restriction on a page
type PageRestriction struct {
	Operation RestrictionOperation `json:"operation"`
	Subject   RestrictionSubject   `json:"restrictions"`
}

// RestrictionSubject represents the subject of a restriction
type RestrictionSubject struct {
	User  *RestrictionResults `json:"user,omitempty"`
	Group *RestrictionResults `json:"group,omitempty"`
}

// RestrictionResults contains restriction subjects
type RestrictionResults struct {
	Results []RestrictionEntity `json:"results"`
}

// RestrictionEntity represents a user or group in a restriction
type RestrictionEntity struct {
	Type        string `json:"type"`
	AccountID   string `json:"accountId,omitempty"`   // For users
	Name        string `json:"name,omitempty"`        // For groups
	DisplayName string `json:"displayName,omitempty"` // Display name
}

// PageRestrictionsResponse represents the API response for page restrictions
type PageRestrictionsResponse struct {
	Results []PageRestriction `json:"results"`
}

// AddPageRestrictions adds restrictions to a page
// operation: "read" or "update"
// users: list of user account IDs
// groups: list of group names
func (c *Client) AddPageRestrictions(ctx context.Context, pageID string, operation RestrictionOperation, users []string, groups []string) error {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	// Build restriction payload
	restriction := PageRestriction{
		Operation: operation,
		Subject: RestrictionSubject{
			User:  &RestrictionResults{Results: []RestrictionEntity{}},
			Group: &RestrictionResults{Results: []RestrictionEntity{}},
		},
	}

	// Add users
	for _, userID := range users {
		restriction.Subject.User.Results = append(restriction.Subject.User.Results, RestrictionEntity{
			Type:      "known",
			AccountID: userID,
		})
	}

	// Add groups
	for _, groupName := range groups {
		restriction.Subject.Group.Results = append(restriction.Subject.Group.Results, RestrictionEntity{
			Type: "group",
			Name: groupName,
		})
	}

	body, err := json.Marshal(restriction)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal restrictions")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/restriction", c.baseURL, pageID)

	return WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		return nil
	})
}

// GetPageRestrictions retrieves all restrictions for a page
func (c *Client) GetPageRestrictions(ctx context.Context, pageID string) (*PageRestrictionsResponse, error) {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/restriction", c.baseURL, pageID)

	var restrictions *PageRestrictionsResponse
	err := WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("page not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		restrictions = &PageRestrictionsResponse{}
		if err := json.NewDecoder(resp.Body).Decode(restrictions); err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
		}

		return nil
	})

	return restrictions, err
}

// RemovePageRestrictions removes all restrictions for a specific operation from a page
func (c *Client) RemovePageRestrictions(ctx context.Context, pageID string, operation RestrictionOperation) error {
	// Validate page ID
	if err := validation.ValidatePageID(pageID); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid page ID")
	}

	url := fmt.Sprintf("%s/rest/api/content/%s/restriction?operation=%s", c.baseURL, pageID, operation)

	return WithExponentialBackoff(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req.SetBasicAuth(c.username, c.apiToken)

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)

			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("authentication failed: %s", string(bodyBytes)))
			case http.StatusForbidden:
				return errors.New(errors.ErrorTypePermission, fmt.Sprintf("permission denied: %s", string(bodyBytes)))
			case http.StatusNotFound:
				return errors.New(errors.ErrorTypeNotFound, fmt.Sprintf("restriction not found: %s", string(bodyBytes)))
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				return errors.New(errors.ErrorTypeRetryable, fmt.Sprintf("service unavailable (status %d): %s", resp.StatusCode, string(bodyBytes)))
			default:
				return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(bodyBytes)))
			}
		}

		return nil
	})
}

// UpdatePageRestrictions updates page restrictions by removing old ones and adding new ones
// This is a helper method that combines RemovePageRestrictions and AddPageRestrictions
func (c *Client) UpdatePageRestrictions(ctx context.Context, pageID string, operation RestrictionOperation, users []string, groups []string) error {
	// First remove existing restrictions for this operation
	if err := c.RemovePageRestrictions(ctx, pageID, operation); err != nil {
		// If restriction doesn't exist, that's fine - continue
		if !errors.IsType(err, errors.ErrorTypeNotFound) {
			return fmt.Errorf("failed to remove existing restrictions: %w", err)
		}
	}

	// Then add new restrictions
	return c.AddPageRestrictions(ctx, pageID, operation, users, groups)
}

// MakePagePublic removes all restrictions from a page, making it accessible to all space members
func (c *Client) MakePagePublic(ctx context.Context, pageID string) error {
	// Remove both read and update restrictions
	if err := c.RemovePageRestrictions(ctx, pageID, RestrictionOperationRead); err != nil {
		if !errors.IsType(err, errors.ErrorTypeNotFound) {
			return err
		}
	}

	if err := c.RemovePageRestrictions(ctx, pageID, RestrictionOperationUpdate); err != nil {
		if !errors.IsType(err, errors.ErrorTypeNotFound) {
			return err
		}
	}

	return nil
}
