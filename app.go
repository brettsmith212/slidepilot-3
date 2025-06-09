package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                     context.Context
	aiAgent                 *AIAgent
	imageCache              map[string]string // Cache for base64 images
	currentPresentationPath string            // Track currently loaded presentation
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		imageCache: make(map[string]string),
	}
	app.aiAgent = NewAIAgent(app)
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start LibreOffice headless service
	if err := StartLibreOfficeHeadless(); err != nil {
		fmt.Printf("Failed to start LibreOffice service: %v\n", err)
	}

	// Create slides directory if it doesn't exist
	os.MkdirAll("slides", 0755)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// SendMessageToAI sends a message to the AI agent and returns the response
func (a *App) SendMessageToAI(message string) (string, error) {
	response, err := a.aiAgent.SendMessage(message)
	// Clear image cache after AI interaction since slides might have been modified
	a.ClearImageCache()
	return response, err
}

// GetSlides returns a list of slide image files in the slides directory
func (a *App) GetSlides() ([]string, error) {
	slidesDir := "slides"

	// Check if slides directory exists
	if _, err := os.Stat(slidesDir); os.IsNotExist(err) {
		return make([]string, 0), nil
	}

	slides := make([]string, 0)
	err := filepath.WalkDir(slidesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && (filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".jpeg") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			slides = append(slides, absPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort slides to ensure consistent ordering
	sort.Strings(slides)

	return slides, nil
}

// OpenPresentationDialog opens a file dialog to select a PowerPoint presentation
func (a *App) OpenPresentationDialog() ([]string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select PowerPoint Presentation",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PowerPoint Files (*.pptx)",
				Pattern:     "*.pptx",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %v", err)
	}

	if selection == "" {
		// User cancelled
		return []string{}, nil
	}

	return a.LoadPresentation(selection)
}

// LoadPresentation loads a PowerPoint file and exports slides to JPEG
func (a *App) LoadPresentation(pptxPath string) ([]string, error) {
	// Clear image cache since we're loading new slides
	a.ClearImageCache()

	// Ensure we have absolute path for AI tools
	absPath, err := filepath.Abs(pptxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	slides, err := ConvertPPTXToJPEG(absPath, "slides")
	if err != nil {
		return nil, fmt.Errorf("failed to load presentation: %v", err)
	}

	// Store the absolute current presentation path for AI tools
	a.currentPresentationPath = absPath
	fmt.Printf("Loaded presentation: %s\n", absPath)

	return slides, nil
}

// GetSlideImagePath returns the absolute path for a slide image
func (a *App) GetSlideImagePath(slidePath string) (string, error) {
	absPath, err := filepath.Abs(slidePath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// GetSlideImageAsBase64 reads a slide image and returns it as base64 data URI
func (a *App) GetSlideImageAsBase64(slidePath string) (string, error) {
	// Check cache first
	if cachedData, exists := a.imageCache[slidePath]; exists {
		return cachedData, nil
	}

	imageBytes, err := os.ReadFile(slidePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}

	// Determine the MIME type based on file extension
	ext := filepath.Ext(slidePath)
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	default:
		mimeType = "image/jpeg" // default to jpeg
	}

	// Convert to base64 data URI
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// Cache the result
	a.imageCache[slidePath] = dataURI

	return dataURI, nil
}

// ClearImageCache clears the image cache (useful when slides are updated)
func (a *App) ClearImageCache() {
	a.imageCache = make(map[string]string)
}

// CheckSlideExists returns whether a slide file exists without logging large data
func (a *App) CheckSlideExists(slidePath string) bool {
	_, err := os.Stat(slidePath)
	return err == nil
}

// GetSlideImageQuiet loads and caches base64 data without logging it, returns simple status
func (a *App) GetSlideImageQuiet(slidePath string) (string, error) {
	// Check cache first
	if _, exists := a.imageCache[slidePath]; exists {
		return "CACHED_BASE64_DATA_AVAILABLE", nil
	}

	// Load image file directly (don't call GetSlideImageAsBase64 to avoid logging)
	imageBytes, err := os.ReadFile(slidePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}

	// Determine MIME type
	ext := filepath.Ext(slidePath)
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	default:
		mimeType = "image/jpeg"
	}

	// Convert to base64 data URI and cache it
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	a.imageCache[slidePath] = dataURI

	// Return simple status instead of the massive base64 string
	return "BASE64_DATA_LOADED", nil
}

// GetCurrentPresentationName returns the name of currently loaded presentation
func (a *App) GetCurrentPresentationName() string {
	if a.currentPresentationPath == "" {
		return ""
	}
	return filepath.Base(a.currentPresentationPath)
}

// HasPresentationLoaded returns whether a presentation is currently loaded
func (a *App) HasPresentationLoaded() bool {
	return a.currentPresentationPath != ""
}
