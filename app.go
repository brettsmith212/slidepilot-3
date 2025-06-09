package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// App struct
type App struct {
	ctx     context.Context
	aiAgent *AIAgent
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		aiAgent: NewAIAgent(),
	}
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
	return a.aiAgent.SendMessage(message)
}

// GetSlides returns a list of slide image files in the slides directory
func (a *App) GetSlides() ([]string, error) {
	slidesDir := "slides"
	
	// Check if slides directory exists
	if _, err := os.Stat(slidesDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	var slides []string
	err := filepath.WalkDir(slidesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && (filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".jpeg") {
			slides = append(slides, path)
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

// LoadPresentation loads a PowerPoint file and exports slides to JPEG
func (a *App) LoadPresentation(pptxPath string) ([]string, error) {
	slides, err := ConvertPPTXToJPEG(pptxPath, "slides")
	if err != nil {
		return nil, fmt.Errorf("failed to load presentation: %v", err)
	}
	
	return slides, nil
}
