# SlidePilot - AI-Powered Slide Editor

## Overview
SlidePilot is a Wails application that combines a Go backend with a React frontend to provide AI-powered PowerPoint slide editing capabilities.

## System Requirements
- LibreOffice (soffice command)
- ImageMagick (convert command)
- Python 3 with UNO bridge
- Go 1.23+
- Node.js and npm

## Development Commands

### Backend
```bash
go mod tidy          # Install dependencies
go run .             # Run in development mode
```

### Frontend
```bash
cd frontend
npm install          # Install dependencies
npm run dev          # Development server
npm run build        # Production build
```

### Wails
```bash
wails dev            # Development mode with hot reload
wails build          # Production build
wails generate module # Regenerate bindings
```

## Project Structure

### Backend (Go)
- `main.go` - Wails application entry point
- `app.go` - Main application struct with frontend bindings
- `ai_agent.go` - Anthropic AI integration and conversation management
- `slide_service.go` - LibreOffice headless service management
- `slide_tools.go` - AI tool definitions for slide operations
- `converter.go` - PowerPoint to JPEG conversion utilities
- `scripts/` - Python UNO scripts for LibreOffice automation

### Frontend (React + TypeScript + Tailwind)
- `src/App.tsx` - Main application component
- `src/components/SlideViewer.tsx` - Slide display and navigation
- `src/components/ChatPanel.tsx` - AI chat interface
- `src/style.css` - Global styles with Tailwind

## Features

### Slide Operations
- Load PowerPoint presentations (.pptx)
- Convert slides to JPEG images
- Display slides in a gallery view
- Navigate between slides

### AI Integration
- Chat interface for natural language slide editing
- Tool-based editing system with the following capabilities:
  - List slides
  - Read slide content
  - Edit slide text
  - Add new slides
  - Delete slides
  - Export slides to images

### UI Features
- Responsive slide viewer with thumbnails
- Collapsible chat panel
- Real-time slide updates after AI modifications
- Loading states and error handling

## Environment Variables
Set `ANTHROPIC_API_KEY` environment variable for AI functionality.

## Recent Fixes & Improvements

### AI Agent Conversation Loop (Fixed)
- **Issue**: Claude would stop mid-conversation after tool calls
- **Solution**: Added proper multi-round inference handling in `ai_agent.go`
- **Result**: Claude now properly acknowledges completed edits

### UI Slide Refresh (Fixed)
- **Issue**: Slides didn't refresh in UI after AI edits
- **Solution**: Auto-export slides after successful edits in `slide_tools.go`
- **Result**: UI updates immediately after slide modifications

### Python Script JSON Output (Fixed)
- **Issue**: Python UNO scripts printed extra text that corrupted JSON parsing
- **Solution**: Removed print statements from `scripts/uno_edit_slide.py`
- **Result**: Clean JSON output that Go can parse properly

### Empty Slides Array (Fixed)
- **Issue**: Empty slides directory caused UI crashes with null array
- **Solution**: Use `make([]string, 0)` instead of `var slides []string` in `app.go`
- **Result**: Graceful handling of empty slide directories

## Current Status
- ✅ AI slide editing fully functional
- ✅ UI updates automatically after edits
- ✅ Proper conversation flow with Claude
- ✅ Robust error handling and logging

## Debugging
- AI conversation logs available in `slides/ai_conversation.log`
- Enhanced debug logging shows inference steps and tool results
- Context injection ensures Claude knows current presentation path

## Known Requirements
- LibreOffice headless service must be running on port 8100
- Python UNO bridge must be properly configured
- `ANTHROPIC_API_KEY` environment variable required

## Testing
- Load any `.pptx` file using "Open Presentation" button
- Use AI chat to edit slides: "Change the title of slide 1 to 'Hello World'"
- Slides auto-refresh in UI after successful edits
