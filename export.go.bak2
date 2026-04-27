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
	"sync"
	"unicode"
	"unsafe"
)

// ─── State ──────────────────────────────────────────────────────────────────

// exportContent holds the in-memory markdown content to support drag-and-drop exports.
var exportContent string

// exportResult holds the result from the ObjC callback.
type exportResult struct {
	path string
	err  error
}

var (
	exportMu   sync.Mutex
	exportDone bool
	exportRes  exportResult
)

//export goExportHTMLResult
func goExportHTMLResult(path, errorMsg *C.char) {
	exportMu.Lock()
	defer exportMu.Unlock()

	var err error
	if errorMsg != nil {
		err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else if path == nil {
		err = fmt.Errorf("cancelled")
	}
	exportRes = exportResult{path: GoStringOrEmpty(path), err: err}
	exportDone = true
	fmt.Fprintf(os.Stderr, "[CB] goExportHTMLResult done=true err=%v\n", err)
}

//export goExportPDFResult
func goExportPDFResult(path, errorMsg *C.char) {
	exportMu.Lock()
	defer exportMu.Unlock()

	var err error
	if errorMsg != nil {
		err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else if path == nil {
		err = fmt.Errorf("cancelled")
	}
	exportRes = exportResult{path: GoStringOrEmpty(path), err: err}
	exportDone = true
	fmt.Fprintf(os.Stderr, "[CB] goExportPDFResult done=true err=%v\n", err)
}

func GoStringOrEmpty(p *C.char) string {
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func getExportFilename() string {
	base := "untitled"
	if currentFile != "" && !strings.HasPrefix(currentFile, "(dragged)") {
		base = strings.TrimSuffix(filepath.Base(currentFile), filepath.Ext(currentFile))
	}
	return base
}

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

	html = removeHTMLBlock(html, `<div class="drop-zone"`, `</div>`)
	html = removeHTMLBlock(html, `<div class="keyboard-hint"`, `</div>`)
	html = removeHTMLBlock(html, `<div class="settings-overlay"`, `</div>`)
	html = removeHTMLBlock(html, `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, `</link>`)
	html = removeHTMLBlock(html, `<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, `</script>`)

	theme := currentConfig.Theme
	if theme == "" || theme == "auto" {
		theme = "light"
	}
	html = strings.Replace(html, `data-theme="auto"`, fmt.Sprintf(`data-theme=%q`, theme), 1)

	return html
}

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
// Uses sync.Mutex + flag instead of channels to avoid goroutine-blocking issues.
// goroutine waits on exportDone flag; callback sets it to true and returns.
// sync.Mutex is released during cond.Wait(), so Go scheduler can run other goroutines.
func exportHTML() {
	exportMu.Lock()
	exportDone = false
	exportRes = exportResult{}
	exportMu.Unlock()

	html := buildExportHTML()
	if html == "" {
		fmt.Fprintln(os.Stderr, "[ExportHTML] empty HTML")
		return
	}

	fmt.Fprintf(os.Stderr, "[ExportHTML] calling C.ExportHTML (len=%d)\n", len(html))
	cName := C.CString(getExportFilename())
	cHTML := C.CString(html)
	C.ExportHTML(cHTML, cName)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cHTML))

	// Wait for ObjC callback by polling the exportDone flag.
	// Lock is released during the check, allowing Go scheduler to run.
	for {
		exportMu.Lock()
		if exportDone {
			res := exportRes
			exportMu.Unlock()
			if res.err != nil {
				fmt.Fprintf(os.Stderr, "[ExportHTML] %v\n", res.err)
			} else {
				fmt.Fprintln(os.Stderr, "[ExportHTML] Saved")
			}
			return
		}
		exportMu.Unlock()
	}
}

// ─── exportPDF ──────────────────────────────────────────────────────────────
func exportPDF() {
	exportMu.Lock()
	exportDone = false
	exportRes = exportResult{}
	exportMu.Unlock()

	html := buildExportHTML()
	if html == "" {
		fmt.Fprintln(os.Stderr, "[ExportPDF] empty HTML")
		return
	}

	fmt.Fprintf(os.Stderr, "[ExportPDF] calling C.ExportPDF (len=%d)\n", len(html))
	cName := C.CString(getExportFilename())
	cHTML := C.CString(html)
	C.ExportPDF(cHTML, cName)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cHTML))

	for {
		exportMu.Lock()
		if exportDone {
			res := exportRes
			exportMu.Unlock()
			if res.err != nil {
				fmt.Fprintf(os.Stderr, "[ExportPDF] %v\n", res.err)
			} else {
				fmt.Fprintln(os.Stderr, "[ExportPDF] Saved")
			}
			return
		}
		exportMu.Unlock()
	}
}
