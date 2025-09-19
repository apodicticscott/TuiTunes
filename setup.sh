#!/bin/bash

# TuiTunes Setup Script
echo "🎵 Setting up TuiTunes..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'o' -f2)
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ Failed to download dependencies"
    exit 1
fi

echo "✅ Dependencies downloaded"

# Build the application
echo "🔨 Building TuiTunes..."
go build -o tuitunes .

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Make the binary executable
chmod +x tuitunes

echo ""
echo "🎉 TuiTunes is ready!"
echo ""
echo "Usage:"
echo "  ./tuitunes                    # Use current directory"
echo "  ./tuitunes /path/to/music     # Use specific directory"
echo ""
echo "Controls:"
echo "  Space - Play/Pause"
echo "  N     - Next track"
echo "  P     - Previous track"
echo "  H     - Show help"
echo "  Q     - Quit"
echo ""
echo "Enjoy your music! 🎵"
