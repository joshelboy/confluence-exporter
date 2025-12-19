package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"confluence-exporter/internal/api"
	"confluence-exporter/internal/config"
	"confluence-exporter/internal/models"
	"confluence-exporter/internal/output"
	"confluence-exporter/pkg/utils"
)

// ProgressTracker keeps track of export progress
type ProgressTracker struct {
	startTime          time.Time
	lastUpdate         time.Time
	totalPages         int
	processedPages     int
	lastPagesPerMinute float64
}

func NewProgressTracker(totalPages int) *ProgressTracker {
	return &ProgressTracker{
		startTime:      time.Now(),
		lastUpdate:     time.Now(),
		totalPages:     totalPages,
		processedPages: 0,
	}
}

func (pt *ProgressTracker) Update() {
	pt.processedPages++
	now := time.Now()
	elapsed := now.Sub(pt.startTime).Minutes()
	if elapsed > 0 {
		pt.lastPagesPerMinute = float64(pt.processedPages) / elapsed
	}
	pt.lastUpdate = now
}

func (pt *ProgressTracker) GetProgressBar() string {
	width := 40
	progress := float64(pt.processedPages) / float64(pt.totalPages)
	filled := int(progress * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, progress*100)
}

func (pt *ProgressTracker) GetStats() string {
	elapsed := time.Since(pt.startTime).Round(time.Second)
	return fmt.Sprintf("⏱️  %s | 📊 %.1f pages/min | 📄 %d/%d pages",
		elapsed, pt.lastPagesPerMinute, pt.processedPages, pt.totalPages)
}

func exportSpace(client *api.ConfluenceClient, spaceKey string, cfg *config.Config, progress *ProgressTracker, handler output.Handler) error {
	// Get all pages from specified space
	log.Printf("🔍 Fetching pages from space: %s", spaceKey)
	pages, err := client.GetPages(spaceKey)
	if err != nil {
		return fmt.Errorf("failed to fetch pages: %v", err)
	}

	log.Printf("📚 Found %d pages to export in space %s", len(pages), spaceKey)

	// Create a progress tracker for this space's pages
	spaceProgress := NewProgressTracker(len(pages))
	spaceProgress.startTime = time.Now()

	// Process each page
	for _, page := range pages {
		// Update and display progress for this space
		spaceProgress.Update()
		fmt.Printf("\r%s | Space: %s | Pages: %s", progress.GetProgressBar(), spaceKey, spaceProgress.GetStats())

		// Save page using the output handler
		if err := handler.SavePage(client, page, spaceKey); err != nil {
			fmt.Println() // New line for error message
			log.Printf("❌ Failed to save page %s: %v", page.Title, err)
			continue
		}
	}

	return nil
}

// fetchPageTree retrieves a page and all of its descendant pages
func fetchPageTree(client *api.ConfluenceClient, rootPageID string) ([]models.Page, error) {
	rootPage, err := client.GetPage(rootPageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch root page %s: %w", rootPageID, err)
	}

	allPages := []models.Page{*rootPage}

	children, err := client.GetChildPages(rootPage.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch child pages for %s: %w", rootPageID, err)
	}

	for _, child := range children {
		subtree, err := collectChildPages(client, child)
		if err != nil {
			return nil, err
		}
		allPages = append(allPages, subtree...)
	}

	return allPages, nil
}

// collectChildPages recursively collects descendant pages for the provided page
func collectChildPages(client *api.ConfluenceClient, page models.Page) ([]models.Page, error) {
	pages := []models.Page{page}

	children, err := client.GetChildPages(page.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch child pages for %s: %w", page.ID, err)
	}

	for _, child := range children {
		subtree, err := collectChildPages(client, child)
		if err != nil {
			return nil, err
		}
		pages = append(pages, subtree...)
	}

	return pages, nil
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logging
	if err := utils.InitLogger(cfg.Logging.File); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	log.Printf("🚀 Starting Confluence export process...")

	// Initialize output handler
	handler, err := output.NewHandler(cfg.Export.OutputType, cfg.Export.OutputDir, cfg.Export.IncludeAttachments)
	if err != nil {
		log.Fatalf("Failed to initialize output handler: %v", err)
	}
	defer handler.Close()

	if err := handler.Initialize(); err != nil {
		log.Fatalf("Failed to initialize output: %v", err)
	}

	// Initialize Confluence client
	client := api.NewConfluenceClient(
		cfg.Confluence.BaseURL,
		cfg.Confluence.Username,
		cfg.Confluence.APIToken,
	)

	var progress *ProgressTracker
	summaryLabel := "Total spaces processed"

	if cfg.Export.PageID != "" {
		log.Printf("📄 Root page ID provided (%s), exporting page tree...", cfg.Export.PageID)
		pages, err := fetchPageTree(client, cfg.Export.PageID)
		if err != nil {
			log.Fatalf("Failed to fetch page tree: %v", err)
		}

		if len(pages) == 0 {
			log.Fatalf("No pages found for root page ID %s", cfg.Export.PageID)
		}

		rootPage := pages[0]
		log.Printf("📚 Found %d pages to export under root page %s (%s)", len(pages), rootPage.Title, cfg.Export.PageID)

		progress = NewProgressTracker(len(pages))
		summaryLabel = "Total pages processed"

		for _, page := range pages {
			progress.Update()
			fmt.Printf("\r%s | Page tree: %s | %s", progress.GetProgressBar(), rootPage.Title, progress.GetStats())

			if err := handler.SavePage(client, page, page.SpaceKey); err != nil {
				fmt.Println()
				log.Printf("❌ Failed to save page %s: %v", page.Title, err)
				continue
			}
		}

		log.Printf("✅ Successfully exported page tree rooted at %s (%s)", rootPage.Title, cfg.Export.PageID)
	} else {
		// Get all spaces if no specific space key is provided
		var spaces []models.Space
		if cfg.Export.SpaceKey == "" {
			log.Printf("🌍 No space key provided, fetching all accessible spaces...")
			spaces, err = client.GetSpaces()
			if err != nil {
				log.Fatalf("Failed to fetch spaces: %v", err)
			}
			log.Printf("📚 Found %d spaces to export", len(spaces))
		} else {
			// Create a single space entry for the specified space
			spaces = []models.Space{{Key: cfg.Export.SpaceKey}}
		}

		// Initialize progress tracker with total spaces
		progress = NewProgressTracker(len(spaces))
		progress.totalPages = len(spaces) // Use spaces count for progress bar

		// Export each space
		for _, space := range spaces {
			log.Printf("🚀 Starting export of space: %s", space.Key)
			if err := exportSpace(client, space.Key, cfg, progress, handler); err != nil {
				log.Printf("❌ Failed to export space %s: %v", space.Key, err)
				continue
			}
			log.Printf("✅ Successfully exported space: %s", space.Key)
			progress.Update() // Update progress after each space
			fmt.Printf("\r%s | %s", progress.GetProgressBar(), progress.GetStats())
		}
	}

	// Print final progress bar
	fmt.Println("\n")
	log.Printf("🎉 Export completed successfully!")

	// Print output location based on type
	switch cfg.Export.OutputType {
	case "file":
		fmt.Printf("✨ Export completed successfully! Files saved to %s\n", cfg.Export.OutputDir)
	case "meilisearch":
		fmt.Printf("✨ Export completed successfully! MeiliSearch JSON saved to %s/confluence_pages_meilisearch.json\n", cfg.Export.OutputDir)
	case "singletxt":
		fmt.Printf("✨ Export completed successfully! Single text file saved to %s/confluence_export.txt\n", cfg.Export.OutputDir)
	default:
		fmt.Printf("✨ Export completed successfully! Data saved to confluence_pages.db\n")
	}

	fmt.Printf("📊 Final statistics:\n")
	fmt.Printf("   • Total time: %s\n", time.Since(progress.startTime).Round(time.Second))
	fmt.Printf("   • %s: %d\n", summaryLabel, progress.processedPages)
}

// getSafeFilename converts a string to a safe filename
func getSafeFilename(name string) string {
	// Replace characters that are not allowed in filenames
	// This is a simplified version, you might need to handle more cases
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		" ", "_",
	)
	return replacer.Replace(name)
}

// downloadAttachment downloads and saves an attachment to disk
func downloadAttachment(client *api.ConfluenceClient, attachment models.Attachment, outputPath string) error {
	// Construct the full download URL
	downloadURL := client.GetBaseURL() + attachment.DownloadURL

	// Get the file
	resp, err := client.GetAttachmentContent(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the output file
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the content to the file
	_, err = io.Copy(out, resp.Body)
	return err
}
