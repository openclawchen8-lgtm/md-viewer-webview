# AGENTS.md — md-viewer-webview

## Build Command

```bash
./build.sh
```

This runs: `go mod tidy` → `swift build -c release` → `go build` → creates `md-viewer.app` bundle.

## Tech Stack

- **Go**: 1.26+ (webview_go for WebKit UI)
- **Swift**: 5.7+ (apple/swift-markdown for parsing)
- **CGO**: Bridges Go ↔ Swift dylib

## Architecture

```
main.go (Go) → CGO → libMarkdownEngine.dylib (Swift) → swift-markdown
                ↓
           webview_go (WKWebView)
```

- `main.go` — Go entry point, menu, file handling, settings
- `core/renderer.go` — CGO bridge to Swift
- `Sources/` — Swift MarkdownEngine using apple/swift-markdown
- `libMarkdownEngine.dylib` — Built by Swift, symlinked to `.build/release/`

## Key Files

| File | Purpose |
|------|---------|
| `build.sh` | Full build + bundle script |
| `main.go` | App entry, NSMenu, settings, file watching |
| `core/renderer.go` | CGO bindings to Swift |
| `Sources/MarkdownEngine/` | Swift markdown parsing |
| `config.go` | Settings persistence (`~/.md-viewer/config.json`) |
| `export.go` / `export.m` | HTML/PDF export |

## Development

- Run directly: `./md-viewer` or `./md-viewer somefile.md`
- Run app bundle: `open md-viewer.app`
- After Swift changes: must re-run `./build.sh` (rebuilds dylib)
- After Go changes: `go build -o md-viewer` (faster iteration)

## Notes

- Uses **swift-markdown** (Apple's), NOT goldmark (old AGENTS.md in `old/` is stale)
- Themes and CSS are in the Go code (search for CSS constants in main.go)
- Settings stored at `~/.md-viewer/config.json`