#!/bin/bash

# md-viewer build & package script
set -e

APP_NAME="md-viewer"
BUNDLE_DIR="md-viewer.app"
CONTENTS_DIR="$BUNDLE_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
FRAMEWORKS_DIR="$CONTENTS_DIR/Frameworks"

echo "🚀 Starting build process..."

# 0. Tidy Go modules
echo "🧹 Tidying Go modules..."
go mod tidy

# 1. Build Swift MarkdownEngine
echo "📦 Building Swift MarkdownEngine..."
##rm -rf .build/
swift build -c release

# 2. Build Go Application
echo "🐹 Building Go application..."
go build -o $APP_NAME

# 3. Prepare .app structure
echo "📂 Preparing .app bundle..."
mkdir -p "$MACOS_DIR"
mkdir -p "$FRAMEWORKS_DIR"

# 4. Copy and fix dynamic library
echo "🔗 Linking Swift library..."
cp .build/release/libMarkdownEngine.dylib "$FRAMEWORKS_DIR/"
# Also keep a copy or symlink in root for direct execution
ln -sf .build/release/libMarkdownEngine.dylib libMarkdownEngine.dylib

# Fix rpath in the executable
echo "🛠 Fixing rpath..."
# Remove old rpaths if they exist to prevent duplication errors
install_name_tool -delete_rpath "@executable_path/../Frameworks/" "$APP_NAME" 2>/dev/null || true
install_name_tool -delete_rpath "./" "$APP_NAME" 2>/dev/null || true
install_name_tool -delete_rpath ".build/release/" "$APP_NAME" 2>/dev/null || true

# Add rpaths
install_name_tool -add_rpath "@executable_path/../Frameworks/" "$APP_NAME"
install_name_tool -add_rpath "./" "$APP_NAME"
install_name_tool -add_rpath ".build/release/" "$APP_NAME"

# 5. Move executable to bundle
echo "🚚 Moving executable to bundle..."
cp $APP_NAME "$MACOS_DIR/"

echo "✅ Build complete!"
echo "👉 You can now run the app via: ./$APP_NAME"
echo "👉 Or use the bundle: open $BUNDLE_DIR"
