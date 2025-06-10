# SlidePilot - AI-Powered Slide Editor

## Setup (macOS)

Install dependencies:

```bash
# Core dependencies
brew install go node # if you don't already have it
brew install --cask libreoffice
brew install ghostscript imagemagick

# Add LibreOffice to PATH
echo 'export PATH="/Applications/LibreOffice.app/Contents/MacOS:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Set Anthropic API key
export ANTHROPIC_API_KEY="your-api-key"
```

## Development

```bash
wails dev    # Run with hot reload
```

## Building

```bash
wails build  # Production build
```
