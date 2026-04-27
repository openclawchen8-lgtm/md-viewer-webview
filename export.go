package main

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11 -x objective-c++ -fobjc-arc
#cgo darwin LDFLAGS: -framework AppKit -framework WebKit

#include <stdlib.h>
#include <stdint.h>

extern void ExportHTML(const char *htmlUTF8, const char *defaultNameUTF8);
extern void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8);
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"unicode"
	"unsafe"
)

// ─── State ──────────────────────────────────────────────────────────────────

var exportActive int32

// exportContent holds the in-memory markdown content to support drag-and-drop exports.
var exportContent string

// exportCh carries the result from the ObjC callback to the waiting goroutine.
// Unbuffered so the goroutine must reach the receive BEFORE the callback fires,
// preventing any race condition.
var exportCh chan error

//export goExportHTMLResult
func goExportHTMLResult(path, errorMsg *C.char) {
	var err error
	if errorMsg != nil {
		err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else if path == nil {
		err = fmt.Errorf("cancelled")
	}
	select {
	case exportCh <- err:
	default:
	}
	_ = path // result delivered via exportCh; path available if needed
}

//export goExportPDFResult
func goExportPDFResult(path, errorMsg *C.char) {
	var err error
	if errorMsg != nil {
		err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else if path == nil {
		err = fmt.Errorf("cancelled")
	}
	select {
	case exportCh <- err:
	default:
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func getExportFilename() string {
	base := "untitled"
	if currentFile != "" && !strings.HasPrefix(currentFile, "(dragged)") {
		base = strings.TrimSuffix(filepath.Base(currentFile), filepath.Ext(currentFile))
	}
	return base
}

// buildExportHTML renders the current markdown content into a clean export HTML.
// Supports file-loaded and drag-and-drop via in-memory exportContent.
func buildExportHTML() string {
	var content string

	if exportContent != "" {
		content = exportContent
	} else if currentFile != "" && !strings.HasPrefix(currentFile, "(dragged)") {
		data, err := os.ReadFile(currentFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Export] failed to read file: %v\n", err)
			return ""
		}
		content = string(data)
	}

	if content == "" {
		fmt.Fprintln(os.Stderr, "[Export] no content available")
		return ""
	}

	html, err := renderer.Render(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Export] failed to render: %v\n", err)
		return ""
	}

	html = fmt.Sprintf(htmlTemplate, cssContent, html)

	// Strip app-only UI elements (drop-zone, keyboard-hint, settings-overlay).
	// Uses nested-tag-aware removal to handle the settings-overlay's inner divs.
	html = removeHTMLBlock(html, `<div class="drop-zone"`, `</div>`)
	html = removeHTMLBlock(html, `<div class="keyboard-hint"`, `</div>`)
	html = removeHTMLBlock(html, `<div class="settings-overlay"`, `</div>`)
	// Remove highlight.js CDN (no JS runtime in standalone HTML export)
	html = removeHTMLBlock(html, `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, `</link>`)
	html = removeHTMLBlock(html, `<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, `</script>`)

	// Apply current theme so the exported file has consistent appearance
	theme := currentConfig.Theme
	if theme == "" || theme == "auto" {
		theme = "light"
	}
	html = strings.Replace(html, `data-theme="auto"`, fmt.Sprintf(`data-theme=%q`, theme), 1)

	return html
}

// removeHTMLBlock removes the first HTML block from startTag to its matching endTag,
// correctly handling nested block-level elements using a depth counter.
func removeHTMLBlock(html, startTag, endTag string) string {
	idx := strings.Index(html, startTag)
	if idx == -1 {
		return html
	}

	depth := 1
	searchFrom := idx + len(startTag)
	for i := searchFrom; i <= len(html)-len(endTag) && depth > 0; i++ {
		if html[i:i+len(endTag)] == endTag {
			depth--
			if depth == 0 {
				return html[:idx] + html[i+len(endTag):]
			}
			continue
		}
		// Check for nested opening block tags
		if html[i] == '<' && i+1 < len(html) {
			c := html[i+1]
			if c != '/' && c != '!' && c != '?' && !unicode.IsSpace(rune(c)) {
				j := i + 1
				for j < len(html) && (unicode.IsLetter(rune(html[j])) || unicode.IsDigit(rune(html[j])) || html[j] == '-' || html[j] == '_') {
					j++
				}
				if j > i+1 && isBlockTag(html[i+1:j]) {
					depth++
				}
			}
		}
	}
	return html[:idx]
}

// isBlockTag reports whether tagName is a block-level HTML tag.
func isBlockTag(name string) bool {
	switch strings.ToLower(name) {
	case "div", "section", "article", "aside", "header", "footer", "nav",
		"main", "figure", "figcaption", "details", "summary",
		"blockquote", "fieldset", "form", "table", "thead", "tbody",
		"tfoot", "tr", "ul", "ol", "li", "dl", "dt", "dd", "p", "h1",
		"h2", "h3", "h4", "h5", "h6", "hr", "pre", "address":
		return true
	}
	return false
}

// ─── exportHTML ─────────────────────────────────────────────────────────────
func exportHTML() {
	if !atomic.CompareAndSwapInt32(&exportActive, 0, 1) {
		fmt.Fprintln(os.Stderr, "[ExportHTML] already in progress, skipping")
		return
	}
	defer atomic.StoreInt32(&exportActive, 0)

	html := buildExportHTML()
	if html == "" {
		fmt.Fprintln(os.Stderr, "[ExportHTML] empty HTML, nothing to export")
		return
	}

	// Unbuffered: goroutine reaches receive BEFORE callback fires → no race.
	exportCh = make(chan error)

	fmt.Fprintf(os.Stderr, "[ExportHTML] calling C.ExportHTML (html len=%d)\n", len(html))
	cName := C.CString(getExportFilename())
	cHTML := C.CString(html)
	C.ExportHTML(cHTML, cName)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cHTML))

	// Wait for ObjC callback. It fires asynchronously on the main thread.
	err := <-exportCh
	exportCh = nil

	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Fprintln(os.Stderr, "[ExportHTML] cancelled")
		} else {
			fmt.Fprintf(os.Stderr, "[ExportHTML] Error: %v\n", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "[ExportHTML] Saved")
	}
}

// ─── exportPDF ──────────────────────────────────────────────────────────────
func exportPDF() {
	if !atomic.CompareAndSwapInt32(&exportActive, 0, 1) {
		fmt.Fprintln(os.Stderr, "[ExportPDF] already in progress, skipping")
		return
	}
	defer atomic.StoreInt32(&exportActive, 0)

	html := buildExportHTML()
	if html == "" {
		fmt.Fprintln(os.Stderr, "[ExportPDF] empty HTML, nothing to export")
		return
	}

	exportCh = make(chan error)

	fmt.Fprintf(os.Stderr, "[ExportPDF] calling C.ExportPDF (html len=%d)\n", len(html))
	cName := C.CString(getExportFilename())
	cHTML := C.CString(html)
	C.ExportPDF(cHTML, cName)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cHTML))

	err := <-exportCh
	exportCh = nil

	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Fprintln(os.Stderr, "[ExportPDF] cancelled")
		} else {
			fmt.Fprintf(os.Stderr, "[ExportPDF] Error: %v\n", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "[ExportPDF] Saved")
	}
}
