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
- **Streaming real-time chat interface** for natural language slide editing
- **Autonomous AI workflow** - Claude works through complex tasks independently
- **Real-time tool status indicators** with emojis (📋 Listing slides..., ✏️ Editing slide text...)
- Tool-based editing system with the following capabilities:
  - List slides
  - Read slide content
  - Edit slide text
  - Add new slides
  - Delete slides
  - Export slides to images

### UI Features
- Responsive slide viewer with thumbnails
- Collapsible chat panel with **streaming message bubbles**
- **Real-time slide updates** after AI modifications via auto-export
- Loading states and error handling
- **Live progress visibility** - see Claude working step-by-step

## Environment Variables
Set `ANTHROPIC_API_KEY` environment variable for AI functionality.

## Architecture

### Streaming Real-Time Chat System
- **Backend**: AI agent emits Wails events (`"ai-message"`) for each message chunk and tool status
- **Frontend**: Listens for events and creates separate chat bubbles in real-time
- **Tool Status**: Live indicators show Claude's progress: "📋 Listing slides...", "👀 Reading slide content...", "✏️ Editing slide text..."
- **Autonomous Operation**: Claude continues working until task completion without user intervention

### AI Agent Flow
1. User sends message → Enhanced with current presentation context
2. Claude processes and makes tool calls as needed
3. Each text response and tool status emitted as separate events
4. Loop continues until Claude provides final response with no more tool calls
5. Slides auto-export after successful edits to refresh UI

## Current Status
- ✅ **Streaming real-time chat** with live tool status indicators
- ✅ **Autonomous AI workflow** - Claude works through complex tasks independently  
- ✅ **Real-time UI updates** - slides refresh automatically after edits
- ✅ **Multi-round conversation** - Claude continues until task completion
- ✅ **Robust error handling** and comprehensive logging

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
- **Watch real-time streaming**: Claude will show live progress with tool status indicators
- **Try complex tasks**: "Update both slides with interesting facts about AI" - see autonomous workflow
- Slides auto-refresh in UI after successful edits

## Key Implementation Details
- **Event System**: Uses Wails `runtime.EventsEmit(ctx, "ai-message", message)` for real-time streaming
- **Tool Status Format**: `"📋 Listing slides..."` (no markdown, just emoji + text + ellipsis)
- **Context Injection**: Each user message enhanced with current presentation path
- **Auto-export**: Successful slide edits trigger immediate JPEG export for UI refresh
- **Autonomous Loop**: Continues until Claude responds with no tool calls
