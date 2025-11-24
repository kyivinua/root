package confluence

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
)

// SpaceExportConfig holds configuration for space export
type SpaceExportConfig struct {
	SpaceKey         string        // Space key to export
	OutputPath       string        // Output file path (ZIP archive)
	IncludeAttachments bool        // Whether to include attachments
	IncludeComments  bool          // Whether to include comments
	MaxConcurrency   int           // Maximum concurrent page fetches
	Timeout          time.Duration // Overall timeout for export
}

// SpaceExportResult contains information about the export operation
type SpaceExportResult struct {
	SpaceKey       string
	PagesExported  int
	AttachmentsExported int
	TotalSize      int64
	ExportPath     string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
}

// SpaceMetadata contains metadata about the exported space
type SpaceMetadata struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ExportDate  time.Time `json:"export_date"`
	Pages       int       `json:"pages"`
	Attachments int       `json:"attachments"`
}

// ExportSpace exports an entire Confluence space to a ZIP archive
func (c *Client) ExportSpace(ctx context.Context, config *SpaceExportConfig) (*SpaceExportResult, error) {
	if config == nil {
		return nil, errors.New(errors.ErrorTypeValidation, "config cannot be nil")
	}

	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 5
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	result := &SpaceExportResult{
		SpaceKey:  config.SpaceKey,
		StartTime: time.Now(),
	}

	// Create ZIP archive
	zipFile, err := os.Create(config.OutputPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create output file")
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// Get space info
	spaceInfo, err := c.GetSpaceInfo(ctx, config.SpaceKey)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeAPI, "failed to get space info")
	}

	// Fetch all pages in the space
	pages, err := c.GetAllPagesInSpace(ctx, config.SpaceKey)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeAPI, "failed to fetch pages")
	}

	result.PagesExported = len(pages)

	// Export pages
	for _, page := range pages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Write page content
		if err := c.writePageToZip(zipWriter, &page); err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write page")
		}

		// Export attachments if requested
		if config.IncludeAttachments {
			attachments, err := c.GetAttachments(page.ID)
			if err != nil {
				// Log error but continue
				continue
			}

			for _, attachment := range attachments {
				if err := c.writeAttachmentToZip(ctx, zipWriter, &page, &attachment); err != nil {
					// Log error but continue
					continue
				}
				result.AttachmentsExported++
			}
		}
	}

	// Write metadata
	metadata := SpaceMetadata{
		Key:         spaceInfo.Key,
		Name:        spaceInfo.Name,
		Description: spaceInfo.Description,
		ExportDate:  time.Now(),
		Pages:       result.PagesExported,
		Attachments: result.AttachmentsExported,
	}

	if err := c.writeMetadataToZip(zipWriter, &metadata); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write metadata")
	}

	// Close ZIP writer to finalize
	if err := zipWriter.Close(); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to finalize ZIP")
	}

	// Get file size
	fileInfo, _ := zipFile.Stat()
	result.TotalSize = fileInfo.Size()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.ExportPath = config.OutputPath

	return result, nil
}

// GetSpaceInfo retrieves information about a space
func (c *Client) GetSpaceInfo(ctx context.Context, spaceKey string) (*SpaceInfo, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/rest/api/space/"+spaceKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(c.username, c.apiToken)

	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("failed to get space info: status %d", resp.StatusCode))
	}

	var spaceInfo SpaceInfo
	if err := json.NewDecoder(resp.Body).Decode(&spaceInfo); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
	}

	return &spaceInfo, nil
}

// GetAllPagesInSpace retrieves all pages in a space with pagination
func (c *Client) GetAllPagesInSpace(ctx context.Context, spaceKey string) ([]Page, error) {
	var allPages []Page
	start := 0
	limit := 100

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		url := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&type=page&limit=%d&start=%d&expand=body.storage,version",
			c.baseURL, spaceKey, limit, start)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
		}

		req = req.WithContext(ctx)
		req.SetBasicAuth(c.username, c.apiToken)

		c.limiter.Wait()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to execute request")
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, errors.New(errors.ErrorTypeAPI, fmt.Sprintf("failed to get pages: status %d", resp.StatusCode))
		}

		var pageResp PageResponse
		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			_ = resp.Body.Close()
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to decode response")
		}
		_ = resp.Body.Close()

		allPages = append(allPages, pageResp.Results...)

		// Check if there are more pages
		if len(pageResp.Results) < limit {
			break
		}

		start += limit
	}

	return allPages, nil
}

// writePageToZip writes a page to the ZIP archive
func (c *Client) writePageToZip(zipWriter *zip.Writer, page *Page) error {
	// Create directory structure: pages/<pageID>/
	pagePath := filepath.Join("pages", page.ID)

	// Write page metadata
	metadataPath := filepath.Join(pagePath, "metadata.json")
	metadataWriter, err := zipWriter.Create(metadataPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create metadata entry")
	}

	metadata := map[string]interface{}{
		"id":      page.ID,
		"title":   page.Title,
		"type":    page.Type,
		"status":  page.Status,
		"version": page.Version,
		"space":   page.Space,
	}

	if err := json.NewEncoder(metadataWriter).Encode(metadata); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write metadata")
	}

	// Write page content (storage format)
	contentPath := filepath.Join(pagePath, "content.html")
	contentWriter, err := zipWriter.Create(contentPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create content entry")
	}

	if _, err := contentWriter.Write([]byte(page.Body.Storage.Value)); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write content")
	}

	return nil
}

// writeAttachmentToZip writes an attachment to the ZIP archive
func (c *Client) writeAttachmentToZip(ctx context.Context, zipWriter *zip.Writer, page *Page, attachment *Attachment) error {
	// Download attachment content
	req, err := http.NewRequest("GET", c.baseURL+attachment.Links.Download, nil)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create request")
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(c.username, c.apiToken)

	c.limiter.Wait()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to download attachment")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.New(errors.ErrorTypeAPI, fmt.Sprintf("failed to download attachment: status %d", resp.StatusCode))
	}

	// Create attachment entry in ZIP
	attachmentPath := filepath.Join("pages", page.ID, "attachments", attachment.Title)
	writer, err := zipWriter.Create(attachmentPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create attachment entry")
	}

	// Copy attachment content
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write attachment content")
	}

	return nil
}

// writeMetadataToZip writes space metadata to the ZIP archive
func (c *Client) writeMetadataToZip(zipWriter *zip.Writer, metadata *SpaceMetadata) error {
	writer, err := zipWriter.Create("space-metadata.json")
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create metadata entry")
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write metadata")
	}

	return nil
}

// SpaceInfo represents information about a Confluence space
type SpaceInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}
