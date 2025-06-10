package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// ListSlidesDefinition defines the list_slides tool
var ListSlidesDefinition = ToolDefinition{
	Name: "list_slides",
	Description: `List all slides in a PowerPoint presentation with basic information.

Use this tool to get an overview of the presentation structure, including slide numbers, titles, and layout information. This is typically the first tool to use when working with a presentation.`,
	InputSchema: ListSlidesInputSchema,
	Function:    ListSlides,
}

type ListSlidesInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
}

var ListSlidesInputSchema = GenerateSchema[ListSlidesInput]()

func ListSlides(app *App, input json.RawMessage) (string, error) {
	listSlidesInput := ListSlidesInput{}
	err := json.Unmarshal(input, &listSlidesInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if listSlidesInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			listSlidesInput.PresentationPath = app.currentPresentationPath
			fmt.Printf("Using current presentation path: %s\n", app.currentPresentationPath)
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	fmt.Printf("Listing slides in: %s\n", listSlidesInput.PresentationPath)

	// Check if file exists
	if _, err := os.Stat(listSlidesInput.PresentationPath); os.IsNotExist(err) {
		return "", fmt.Errorf("presentation file not found: %s", listSlidesInput.PresentationPath)
	}

	// Call Python UNO script
	cmd := exec.Command("python3", "scripts/uno_list_slides.py", listSlidesInput.PresentationPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list slides: %v\nOutput: %s", err, string(output))
	}

	// Validate that the output is valid JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("invalid JSON output from UNO script: %v", err)
	}

	return string(output), nil
}

// ReadSlideDefinition defines the read_slide tool
var ReadSlideDefinition = ToolDefinition{
	Name: "read_slide",
	Description: `Read detailed content from a specific slide including all text shapes and their content.

Use this tool to get detailed information about a specific slide's content, including shape indices, types, and text content. This is essential for understanding slide structure before making edits.`,
	InputSchema: ReadSlideInputSchema,
	Function:    ReadSlide,
}

type ReadSlideInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
	SlideNumber      int    `json:"slide_number" jsonschema_description:"Slide number to read (1-based indexing)"`
}

var ReadSlideInputSchema = GenerateSchema[ReadSlideInput]()

func ReadSlide(app *App, input json.RawMessage) (string, error) {
	readSlideInput := ReadSlideInput{}
	err := json.Unmarshal(input, &readSlideInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if readSlideInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			readSlideInput.PresentationPath = app.currentPresentationPath
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	if readSlideInput.SlideNumber < 1 {
		return "", fmt.Errorf("slide_number must be 1 or greater")
	}

	fmt.Printf("Reading slide %d from: %s\n", readSlideInput.SlideNumber, readSlideInput.PresentationPath)

	// Call Python UNO script
	cmd := exec.Command("python3", "scripts/uno_read_slide.py", readSlideInput.PresentationPath, fmt.Sprintf("%d", readSlideInput.SlideNumber))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to read slide: %v\nOutput: %s", err, string(output))
	}

	// Validate that the output is valid JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("invalid JSON output from UNO script: %v", err)
	}

	return string(output), nil
}

// EditSlideTextDefinition defines the edit_slide_text tool
var EditSlideTextDefinition = ToolDefinition{
	Name: "edit_slide_text",
	Description: `Edit text content on a slide by targeting specific shapes or elements.

Can target by shape index, shape type, or replace specific text. This tool allows precise editing of slide content including titles, text boxes, and bullet points.

Target types:
- "shape_index": Edit specific shape by index (0, 1, 2, ...)
- "shape_type": Edit by type ("title", "content", "text_box")
- "text_replace": Replace specific text (requires old_text)
- "bullet_point": Edit specific bullet point by index
- "bullet_list": Format entire shape as bullet list with proper LibreOffice formatting
  
IMPORTANT for bullet_list: Provide text with each line representing a bullet point, 
but WITHOUT bullet characters (•, *, -). LibreOffice will add proper bullets automatically.
Example: "First point\nSecond point\nThird point" (not "• First point\n• Second point")`,
	InputSchema: EditSlideTextInputSchema,
	Function:    EditSlideText,
}

type EditSlideTextInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
	SlideNumber      int    `json:"slide_number" jsonschema_description:"Slide number to edit (1-based indexing)"`
	TargetType       string `json:"target_type" jsonschema_description:"How to target: 'shape_index', 'shape_type', 'bullet_point', 'bullet_list', or 'text_replace'"`
	TargetValue      string `json:"target_value" jsonschema_description:"Shape index (0,1,2...), shape type ('title','content','text_box'), bullet index, or text to find"`
	NewText          string `json:"new_text" jsonschema_description:"New text content to set"`
	OldText          string `json:"old_text,omitempty" jsonschema_description:"(Optional) For text_replace mode, the exact text to replace"`
}

var EditSlideTextInputSchema = GenerateSchema[EditSlideTextInput]()

func EditSlideText(app *App, input json.RawMessage) (string, error) {
	editInput := EditSlideTextInput{}
	err := json.Unmarshal(input, &editInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if editInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			editInput.PresentationPath = app.currentPresentationPath
			fmt.Printf("EditSlideText using current presentation path: %s\n", app.currentPresentationPath)
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	// Check if file exists and is accessible
	if _, err := os.Stat(editInput.PresentationPath); os.IsNotExist(err) {
		return "", fmt.Errorf("presentation file not found: %s", editInput.PresentationPath)
	}

	fmt.Printf("EditSlideText operating on: %s (slide %d, target: %s)\n",
		editInput.PresentationPath, editInput.SlideNumber, editInput.TargetType)

	if editInput.SlideNumber < 1 {
		return "", fmt.Errorf("slide_number must be 1 or greater")
	}

	if editInput.TargetType == "" {
		return "", fmt.Errorf("target_type is required")
	}

	if editInput.TargetValue == "" {
		return "", fmt.Errorf("target_value is required")
	}

	if editInput.NewText == "" {
		return "", fmt.Errorf("new_text is required")
	}

	if editInput.TargetType == "text_replace" && editInput.OldText == "" {
		return "", fmt.Errorf("old_text is required for text_replace mode")
	}

	fmt.Printf("Editing slide %d: %s=%s -> '%s'\n",
		editInput.SlideNumber, editInput.TargetType, editInput.TargetValue, editInput.NewText)

	// Build command arguments
	args := []string{
		"scripts/uno_edit_slide.py",
		editInput.PresentationPath,
		fmt.Sprintf("%d", editInput.SlideNumber),
		editInput.TargetType,
		editInput.TargetValue,
		editInput.NewText,
	}

	// Add old_text if provided
	if editInput.OldText != "" {
		args = append(args, editInput.OldText)
	}

	// Call Python UNO script
	cmd := exec.Command("python3", args...)
	
	// Log working directory for debugging
	wd, _ := os.Getwd()
	fmt.Printf("EditSlideText working directory: %s\n", wd)
	fmt.Printf("EditSlideText command: python3 %v\n", args)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to edit slide: %v\nOutput: %s", err, string(output))
	}

	// Validate that the output is valid JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("invalid JSON output from UNO script: %v", err)
	}

	// Parse result to check if edit was successful
	var editResult map[string]interface{}
	if err := json.Unmarshal(output, &editResult); err == nil {
		if success, ok := editResult["success"].(bool); ok && success {
			// Auto-export the edited slide to update UI
			fmt.Printf("EditSlideText: Auto-exporting slide %d to update UI\n", editInput.SlideNumber)
			exportInput := ExportSlidesInput{
				PresentationPath: editInput.PresentationPath,
				SlideNumbers:     []int{editInput.SlideNumber},
				OutputDir:        "slides",
			}
			exportInputJSON, _ := json.Marshal(exportInput)
			_, exportErr := ExportSlides(app, exportInputJSON)
			if exportErr != nil {
				fmt.Printf("Warning: Failed to auto-export slide after edit: %v\n", exportErr)
			}
		}
	}

	return string(output), nil
}

// ExportSlidesDefinition defines the export_slides tool
var ExportSlidesDefinition = ToolDefinition{
	Name: "export_slides",
	Description: `Export slides as JPEG images for preview or verification.

Use this tool to generate visual representations of slides, especially useful after making edits to verify changes. Can export all slides or specific slides.`,
	InputSchema: ExportSlidesInputSchema,
	Function:    ExportSlides,
}

type ExportSlidesInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
	SlideNumbers     []int  `json:"slide_numbers,omitempty" jsonschema_description:"Specific slides to export (optional, defaults to all slides)"`
	OutputDir        string `json:"output_dir,omitempty" jsonschema_description:"Directory to save images (optional, defaults to 'slides/')"`
}

var ExportSlidesInputSchema = GenerateSchema[ExportSlidesInput]()

func ExportSlides(app *App, input json.RawMessage) (string, error) {
	exportInput := ExportSlidesInput{}
	err := json.Unmarshal(input, &exportInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if exportInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			exportInput.PresentationPath = app.currentPresentationPath
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	// Set default output directory
	outputDir := exportInput.OutputDir
	if outputDir == "" {
		outputDir = "slides"
	}

	fmt.Printf("Exporting slides from: %s to %s/\n", exportInput.PresentationPath, outputDir)

	// Use our existing conversion function
	slides, err := ConvertPPTXToJPEG(exportInput.PresentationPath, outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to export slides: %v", err)
	}

	// Filter slides if specific slide numbers were requested
	var filteredSlides []string
	if len(exportInput.SlideNumbers) > 0 {
		slideMap := make(map[int]bool)
		for _, num := range exportInput.SlideNumbers {
			slideMap[num-1] = true // Convert to 0-based indexing
		}

		for i, slide := range slides {
			if slideMap[i] {
				filteredSlides = append(filteredSlides, slide)
			}
		}
		slides = filteredSlides
	}

	result := map[string]interface{}{
		"success":     true,
		"slide_count": len(slides),
		"slides":      slides,
		"output_dir":  outputDir,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// AddSlideDefinition defines the add_slide tool
var AddSlideDefinition = ToolDefinition{
	Name: "add_slide",
	Description: `Add a new slide to the presentation with optional initial content.

Use this tool to create new slides in the presentation. You can specify position, layout type, and initial title content.`,
	InputSchema: AddSlideInputSchema,
	Function:    AddSlide,
}

type AddSlideInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
	Position         int    `json:"position,omitempty" jsonschema_description:"Position to insert slide (optional, defaults to end, 1-based indexing)"`
	Layout           string `json:"layout,omitempty" jsonschema_description:"Slide layout type (optional, defaults to 'blank')"`
	Title            string `json:"title,omitempty" jsonschema_description:"Initial title text for the slide (optional)"`
}

var AddSlideInputSchema = GenerateSchema[AddSlideInput]()

func AddSlide(app *App, input json.RawMessage) (string, error) {
	addSlideInput := AddSlideInput{}
	err := json.Unmarshal(input, &addSlideInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if addSlideInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			addSlideInput.PresentationPath = app.currentPresentationPath
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	// Set defaults
	layout := addSlideInput.Layout
	if layout == "" {
		layout = "blank"
	}

	fmt.Printf("Adding slide to: %s\n", addSlideInput.PresentationPath)
	if addSlideInput.Position > 0 {
		fmt.Printf("Position: %d\n", addSlideInput.Position)
	}
	if addSlideInput.Title != "" {
		fmt.Printf("Title: %s\n", addSlideInput.Title)
	}

	// Build command arguments
	args := []string{
		"scripts/uno_add_slide.py",
		addSlideInput.PresentationPath,
	}

	// Add optional arguments
	if addSlideInput.Position > 0 {
		args = append(args, fmt.Sprintf("%d", addSlideInput.Position))
	} else {
		args = append(args, "") // Empty position means append to end
	}

	args = append(args, layout)

	if addSlideInput.Title != "" {
		args = append(args, addSlideInput.Title)
	}

	// Call Python UNO script
	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to add slide: %v\nOutput: %s", err, string(output))
	}

	// Validate that the output is valid JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("invalid JSON output from UNO script: %v", err)
	}

	// Parse the result to get slide information
	var addResult map[string]interface{}
	if err := json.Unmarshal(output, &addResult); err != nil {
		return "", fmt.Errorf("failed to parse add slide result: %v", err)
	}

	// Automatically export slides for visual verification (like edit_slide_text does)
	fmt.Printf("Exporting slides for visual verification...\n")
	slides, exportErr := ConvertPPTXToJPEG(addSlideInput.PresentationPath, "slides")
	if exportErr != nil {
		// Don't fail the add operation if export fails, just warn
		fmt.Printf("Warning: Failed to export slides for preview: %v\n", exportErr)
	} else {
		// Add export information to the result
		addResult["exported_slides"] = slides
		addResult["slides_directory"] = "slides"

		// Re-marshal the enhanced result
		enhancedResult, _ := json.Marshal(addResult)
		return string(enhancedResult), nil
	}

	return string(output), nil
}

// DeleteSlideDefinition defines the delete_slide tool
var DeleteSlideDefinition = ToolDefinition{
	Name: "delete_slide",
	Description: `Delete a slide from the presentation.

Use this tool to remove unwanted slides from the presentation. The slide numbers will be automatically adjusted after deletion.`,
	InputSchema: DeleteSlideInputSchema,
	Function:    DeleteSlide,
}

type DeleteSlideInput struct {
	PresentationPath string `json:"presentation_path" jsonschema_description:"Path to the PowerPoint (.pptx) file"`
	SlideNumber      int    `json:"slide_number" jsonschema_description:"Slide number to delete (1-based indexing)"`
}

var DeleteSlideInputSchema = GenerateSchema[DeleteSlideInput]()

func DeleteSlide(app *App, input json.RawMessage) (string, error) {
	deleteSlideInput := DeleteSlideInput{}
	err := json.Unmarshal(input, &deleteSlideInput)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Use current presentation path if not provided
	if deleteSlideInput.PresentationPath == "" {
		if app != nil && app.currentPresentationPath != "" {
			deleteSlideInput.PresentationPath = app.currentPresentationPath
		} else {
			return "", fmt.Errorf("no presentation loaded - please load a presentation first")
		}
	}

	if deleteSlideInput.SlideNumber < 1 {
		return "", fmt.Errorf("slide_number must be 1 or greater")
	}

	fmt.Printf("Deleting slide %d from: %s\n", deleteSlideInput.SlideNumber, deleteSlideInput.PresentationPath)

	// Call Python UNO script
	cmd := exec.Command("python3", "scripts/uno_delete_slide.py", deleteSlideInput.PresentationPath, fmt.Sprintf("%d", deleteSlideInput.SlideNumber))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to delete slide: %v\nOutput: %s", err, string(output))
	}

	// Validate that the output is valid JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("invalid JSON output from UNO script: %v", err)
	}

	// Parse the result to get slide information
	var deleteResult map[string]interface{}
	if err := json.Unmarshal(output, &deleteResult); err != nil {
		return "", fmt.Errorf("failed to parse delete slide result: %v", err)
	}

	// Automatically export slides for visual verification (like add_slide does)
	fmt.Printf("Exporting slides for visual verification...\n")
	slides, exportErr := ConvertPPTXToJPEG(deleteSlideInput.PresentationPath, "slides")
	if exportErr != nil {
		// Don't fail the delete operation if export fails, just warn
		fmt.Printf("Warning: Failed to export slides for preview: %v\n", exportErr)
	} else {
		// Add export information to the result
		deleteResult["exported_slides"] = slides
		deleteResult["slides_directory"] = "slides"

		// Re-marshal the enhanced result
		enhancedResult, _ := json.Marshal(deleteResult)
		return string(enhancedResult), nil
	}

	return string(output), nil
}
