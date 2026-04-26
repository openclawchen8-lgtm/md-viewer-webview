package core

/*
#cgo LDFLAGS: -L../.build/release -lMarkdownEngine -Wl,-rpath,./
#include <stdlib.h>

extern char* render_markdown_to_html(const char* input);
*/
import "C"
import (
	"errors"
	"unsafe"
)

// MarkdownRenderer wraps the Swift-based markdown engine.
type MarkdownRenderer struct {
}

// NewMarkdownRenderer creates a new MarkdownRenderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

// Render converts markdown to HTML using swift-markdown via FFI.
func (r *MarkdownRenderer) Render(content string) (string, error) {
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cContent))

	cResult := C.render_markdown_to_html(cContent)
	if cResult == nil {
		return "", errors.New("swift-markdown rendering failed")
	}
	defer C.free(unsafe.Pointer(cResult))

	return C.GoString(cResult), nil
}
