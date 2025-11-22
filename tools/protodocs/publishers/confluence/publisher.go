package confluence

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PublisherConfig holds the configuration for the Confluence publisher.
type PublisherConfig struct {
	BaseURL      string
	Username     string
	APIToken     string
	SpaceKey     string
	ParentPageID string

	CreatePagePerService bool
	PageTitlePrefix      string
	IncludeTOC           bool
	IncludeDiagrams      bool
	IncludeCodeExamples  bool

	UpdateExisting   bool
	VersionLabel     string
	VisibilityFilter []string
}

// Publisher publishes documentation to Confluence.
type Publisher struct {
	config    *PublisherConfig
	client    *Client
	formatter *Formatter
	logger    *log.Logger
}

// NewPublisher creates a new Confluence publisher.
func NewPublisher(config *PublisherConfig) *Publisher {
	// Override with environment variables if set
	if envUsername := os.Getenv("CONFLUENCE_USERNAME"); envUsername != "" {
		config.Username = envUsername
	}
	if envToken := os.Getenv("CONFLUENCE_API_TOKEN"); envToken != "" {
		config.APIToken = envToken
	}

	return &Publisher{
		config:    config,
		client:    NewClient(config.BaseURL, config.Username, config.APIToken),
		formatter: NewFormatter(config.IncludeDiagrams),
		logger:    log.New(os.Stdout, "[confluence] ", log.LstdFlags),
	}
}

// PublishResult holds the result of publishing.
type PublishResult struct {
	PagesCreated int
	PagesUpdated int
	Errors       []error
	PageURLs     []string
}

// PublishFromMarkdownFiles publishes documentation from Markdown files.
func (p *Publisher) PublishFromMarkdownFiles(docsDir string) (*PublishResult, error) {
	p.logger.Printf("Publishing documentation from %s", docsDir)

	result := &PublishResult{
		Errors:   make([]error, 0),
		PageURLs: make([]string, 0),
	}

	// Find all markdown files
	var mdFiles []string
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk docs directory: %w", err)
	}

	if len(mdFiles) == 0 {
		p.logger.Println("No markdown files found")
		return result, nil
	}

	p.logger.Printf("Found %d markdown files", len(mdFiles))

	// Publish each file
	for _, mdFile := range mdFiles {
		// Skip README.md as it's typically an index
		if filepath.Base(mdFile) == "README.md" {
			p.logger.Printf("Skipping index file: %s", mdFile)
			continue
		}

		if err := p.publishFile(mdFile, result); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", mdFile, err))
			p.logger.Printf("Error publishing %s: %v", mdFile, err)
			continue
		}
	}

	p.logger.Printf("Publishing complete: %d created, %d updated, %d errors",
		result.PagesCreated, result.PagesUpdated, len(result.Errors))

	return result, nil
}

// publishFile publishes a single markdown file.
func (p *Publisher) publishFile(mdFile string, result *PublishResult) error {
	// Read markdown content
	content, err := os.ReadFile(mdFile)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	markdown := string(content)

	// Extract title from first H1 or filename
	title := p.extractTitle(markdown, mdFile)

	// Add prefix to title
	if p.config.PageTitlePrefix != "" {
		title = p.config.PageTitlePrefix + " " + title
	}

	p.logger.Printf("Publishing page: %s", title)

	// Convert to Confluence Storage Format
	storage := p.formatter.ConvertMarkdownToStorage(markdown)

	// Add TOC if configured
	if p.config.IncludeTOC {
		storage = p.formatter.FormatTOC() + "\n\n" + storage
	}

	// Check if page already exists
	existingPage, err := p.client.FindPageByTitle(p.config.SpaceKey, title)
	if err != nil {
		return fmt.Errorf("check existing page: %w", err)
	}

	if existingPage != nil && p.config.UpdateExisting {
		// Update existing page
		p.logger.Printf("Updating existing page: %s (ID: %s)", title, existingPage.ID)

		updatedPage := &Page{
			ID:      existingPage.ID,
			Type:    "page",
			Title:   title,
			Space:   Space{Key: p.config.SpaceKey},
			Body:    Body{Storage: Storage{Value: storage, Representation: "storage"}},
			Version: Version{Number: existingPage.Version.Number + 1, Message: p.config.VersionLabel},
		}

		_, err := p.client.UpdatePage(existingPage.ID, updatedPage)
		if err != nil {
			return fmt.Errorf("update page: %w", err)
		}

		result.PagesUpdated++
		pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", p.config.BaseURL, existingPage.ID)
		result.PageURLs = append(result.PageURLs, pageURL)
		p.logger.Printf("✓ Updated: %s", pageURL)

	} else if existingPage == nil {
		// Create new page
		p.logger.Printf("Creating new page: %s", title)

		newPage := &Page{
			Type:  "page",
			Title: title,
			Space: Space{Key: p.config.SpaceKey},
			Body:  Body{Storage: Storage{Value: storage, Representation: "storage"}},
		}

		// Add parent if configured
		if p.config.ParentPageID != "" {
			newPage.Ancestors = []Ancestor{{ID: p.config.ParentPageID}}
		}

		created, err := p.client.CreatePage(newPage)
		if err != nil {
			return fmt.Errorf("create page: %w", err)
		}

		result.PagesCreated++
		pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", p.config.BaseURL, created.ID)
		result.PageURLs = append(result.PageURLs, pageURL)
		p.logger.Printf("✓ Created: %s", pageURL)

	} else {
		p.logger.Printf("Page exists but update disabled: %s", title)
	}

	return nil
}

// extractTitle extracts the title from markdown or filename.
func (p *Publisher) extractTitle(markdown, filename string) string {
	// Try to extract from first H1
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			// Remove emoji if present
			title = strings.TrimSpace(strings.TrimPrefix(title, "📚"))
			title = strings.TrimSpace(strings.TrimPrefix(title, "📖"))
			title = strings.TrimSpace(strings.TrimPrefix(title, "🏗️"))
			return title
		}
	}

	// Fall back to filename
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// PublishConsolidatedPage publishes all services as a single consolidated page.
func (p *Publisher) PublishConsolidatedPage(docsDir, title string) (*PublishResult, error) {
	p.logger.Printf("Publishing consolidated page: %s", title)

	result := &PublishResult{
		Errors:   make([]error, 0),
		PageURLs: make([]string, 0),
	}

	// Find all markdown files
	var mdFiles []string
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" && filepath.Base(path) != "README.md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk docs directory: %w", err)
	}

	// Consolidate all content
	var consolidatedContent strings.Builder
	consolidatedContent.WriteString(fmt.Sprintf("# %s\n\n", title))
	consolidatedContent.WriteString("This page contains consolidated documentation for all services.\n\n")

	for _, mdFile := range mdFiles {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", mdFile, err))
			continue
		}

		// Add separator and content
		consolidatedContent.WriteString("\n\n---\n\n")
		consolidatedContent.WriteString(string(content))
	}

	// Convert and publish
	storage := p.formatter.ConvertMarkdownToStorage(consolidatedContent.String())

	if p.config.IncludeTOC {
		storage = p.formatter.FormatTOC() + "\n\n" + storage
	}

	// Check for existing page
	existingPage, err := p.client.FindPageByTitle(p.config.SpaceKey, title)
	if err != nil {
		return nil, fmt.Errorf("check existing page: %w", err)
	}

	if existingPage != nil && p.config.UpdateExisting {
		updatedPage := &Page{
			ID:      existingPage.ID,
			Type:    "page",
			Title:   title,
			Space:   Space{Key: p.config.SpaceKey},
			Body:    Body{Storage: Storage{Value: storage, Representation: "storage"}},
			Version: Version{Number: existingPage.Version.Number + 1, Message: p.config.VersionLabel},
		}

		_, err := p.client.UpdatePage(existingPage.ID, updatedPage)
		if err != nil {
			return nil, fmt.Errorf("update page: %w", err)
		}

		result.PagesUpdated++
		pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", p.config.BaseURL, existingPage.ID)
		result.PageURLs = append(result.PageURLs, pageURL)

	} else if existingPage == nil {
		newPage := &Page{
			Type:  "page",
			Title: title,
			Space: Space{Key: p.config.SpaceKey},
			Body:  Body{Storage: Storage{Value: storage, Representation: "storage"}},
		}

		if p.config.ParentPageID != "" {
			newPage.Ancestors = []Ancestor{{ID: p.config.ParentPageID}}
		}

		created, err := p.client.CreatePage(newPage)
		if err != nil {
			return nil, fmt.Errorf("create page: %w", err)
		}

		result.PagesCreated++
		pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", p.config.BaseURL, created.ID)
		result.PageURLs = append(result.PageURLs, pageURL)
	}

	p.logger.Printf("Publishing complete: %d pages", result.PagesCreated+result.PagesUpdated)
	return result, nil
}
