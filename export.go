package main

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11 -x objective-c++ -fobjc-arc
#cgo darwin LDFLAGS: -framework AppKit -framework WebKit

#include <stdlib.h>
#include <stdint.h>

extern void ExportHTML(const char *htmlUTF8, const char *defaultNameUTF8);
extern void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8, void *windowPtr);
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	exportCond = sync.NewCond(&exportMu)
	exportDone bool
	exportRes  exportResult
)

//export goExportHTMLResult
func goExportHTMLResult(path, errorMsg *C.char) {
	exportMu.Lock()
	defer exportMu.Unlock()
	exportRes = buildExportResult(path, errorMsg)
	exportDone = true
	exportCond.Signal()
}

//export goExportPDFResult
func goExportPDFResult(path, errorMsg *C.char) {
	exportMu.Lock()
	defer exportMu.Unlock()
	exportRes = buildExportResult(path, errorMsg)
	exportDone = true
	exportCond.Signal()
}

func buildExportResult(path, errorMsg *C.char) exportResult {
	if errorMsg != nil {
		return exportResult{err: fmt.Errorf("%s", C.GoString(errorMsg))}
	}
	if path == nil {
		return exportResult{err: fmt.Errorf("cancelled")}
	}
	return exportResult{path: C.GoString(path)}
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
	html = removeHTMLBlock(html, `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, "")
	html = removeHTMLBlock(html, `<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js`, "")

	theme := currentConfig.Theme
	if theme == "" || theme == "auto" {
		theme = "light"
	}
	html = strings.Replace(html, `data-theme="auto"`, fmt.Sprintf(`data-theme=%q`, theme), 1)

	return html
}

func removeHTMLBlock(html, startTag, endTag string) string {
	startIdx := strings.Index(html, startTag)
	if startIdx == -1 {
		return html
	}

	// 找到 opening tag 的結束位置（第一個 '>'）
	openEnd := strings.Index(html[startIdx:], ">")
	if openEnd == -1 {
		return html[:startIdx]
	}
	openEnd += startIdx + 1 // 指向 '>' 之後

	// endTag 為空代表 self-closing tag，直接砍掉 opening tag 本身
	if endTag == "" {
		return html[:startIdx] + html[openEnd:]
	}

	// 用深度追蹤找對應的 closing tag，從 opening tag 結束後開始掃
	depth := 1
	i := openEnd
	for i <= len(html)-len(endTag) {
		if html[i] == '<' {
			k := i + 1
			isClose := k < len(html) && html[k] == '/'
			if isClose {
				k++
			}
			tagEnd := k
			for tagEnd < len(html) && isLetter(html[tagEnd]) {
				tagEnd++
			}
			tagName := html[k:tagEnd]
			if isBlockTag(tagName) {
				if !isClose {
					depth++
				} else {
					depth--
					if depth == 0 {
						closeEnd := strings.Index(html[i:], ">")
						if closeEnd == -1 {
							break
						}
						return html[:startIdx] + html[i+closeEnd+1:]
					}
				}
			}
		}
		i++
	}
	return html[:startIdx]
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isBlockTag(name string) bool {
	switch name {
	case "div", "section", "article", "aside", "header", "footer", "nav",
		"main", "figure", "figcaption", "details", "summary",
		"blockquote", "fieldset", "form", "table", "thead", "tbody",
		"tfoot", "tr", "ul", "ol", "li", "dl", "dt", "dd", "p", "h1",
		"h2", "h3", "h4", "h5", "h6", "hr", "pre", "address":
		return true
	}
	return false
}

// ─── doExport ───────────────────────────────────────────────────────────────

func doExport(label string, callC func(*C.char, *C.char)) {
	exportMu.Lock()
	exportDone = false
	exportRes = exportResult{}
	exportMu.Unlock()

	html := buildExportHTML()
	if html == "" {
		fmt.Fprintf(os.Stderr, "[%s] empty HTML\n", label)
		return
	}

	fmt.Fprintf(os.Stderr, "[%s] calling C fn (len=%d)\n", label, len(html))
	cName := C.CString(getExportFilename())
	cHTML := C.CString(html)
	callC(cHTML, cName)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cHTML))

	exportMu.Lock()
	for !exportDone {
		exportCond.Wait()
	}
	res := exportRes
	exportMu.Unlock()

	if res.err != nil {
		fmt.Fprintf(os.Stderr, "[%s] %v\n", label, res.err)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] Saved\n", label)
	}
}

func exportHTML() { go doExport("ExportHTML", func(cHTML, cName *C.char) { C.ExportHTML(cHTML, cName) }) }
func exportPDF()  { go doExport("ExportPDF", func(cHTML, cName *C.char) { C.ExportPDF(cHTML, cName, wv.Window()) }) }
