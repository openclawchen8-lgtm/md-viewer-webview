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
)

// ─── Callback ───────────────────────────────────────────────────────────────

type exportResult struct {
	path string
	err  error
}

var exportCh chan exportResult // non-nil while export is in-flight

//export goExportHTMLResult
func goExportHTMLResult(path, errorMsg *C.char) {
	if exportCh == nil {
		return
	}
	r := exportResult{}
	if path != nil {
		r.path = C.GoString(path)
	} else if errorMsg != nil {
		r.err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else {
		r.err = fmt.Errorf("cancelled")
	}
	select {
	case exportCh <- r:
	default:
		// nobody waiting (shouldn't happen with non-blocking launch)
	}
}

//export goExportPDFResult
func goExportPDFResult(path, errorMsg *C.char) {
	if exportCh == nil {
		return
	}
	r := exportResult{}
	if path != nil {
		r.path = C.GoString(path)
	} else if errorMsg != nil {
		r.err = fmt.Errorf("%s", C.GoString(errorMsg))
	} else {
		r.err = fmt.Errorf("cancelled")
	}
	select {
	case exportCh <- r:
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

// buildExportJS extracts current rendered content as a standalone HTML document.
func buildExportJS() string {
	return `(function(){
		var md = document.getElementById('mdContent');
		var body = document.body;
		if (!md || !body) return '';
		var theme = body.getAttribute ? body.getAttribute('data-theme') : 'auto';
		var lang  = document.documentElement.lang || 'en';
		var zoom  = body.style.zoom || '1';
		var styles = '';
		var s = document.querySelectorAll('style');
		for (var i=0;i<s.length;i++) styles += s[i].textContent + '\n';
		var links = '';
		document.querySelectorAll('link[rel="stylesheet"]').forEach(function(el){
			links += '<link rel="stylesheet" href="' + el.href + '">\n';
		});
		var scripts = document.querySelectorAll('script[src]');
		var hljs = scripts.length > 0
			? '<script src="' + scripts[0].src + '"></script>\n'
			: '';
		return '<!DOCTYPE html>\n<html lang="' + lang + '">\n<head>\n<meta charset="UTF-8">\n<meta name="viewport" content="width=device-width,initial-scale=1">\n<style>' + styles + '</style>\n' + links + hljs + '</head>\n<body data-theme="' + theme + '" style="zoom:' + zoom + ';">' + md.outerHTML + '\n<script>hljs && hljs.highlightAll && hljs.highlightAll();</script>\n</body>\n</html>';
	})()`
}

// ─── exportHTML ─────────────────────────────────────────────────────────────

func exportHTML() {
	if currentWV == nil {
		return
	}
	// Non-blocking: channel is nil or already has a pending result
	if exportCh != nil {
		return
	}
	exportCh = make(chan exportResult, 1)

	// wv.Eval must run on the main WebView thread via Dispatch.
	// We dispatch the JS eval to the WebView, which runs immediately,
	// then schedules the callback on the same thread. The non-blocking
	// select means we return immediately; the callback fires asynchronously.
	wv.Dispatch(func() {
		js := "try{ window.onExportHTML && window.onExportHTML(" + buildExportJS() + "); } catch(e){ window.onExportHTML && window.onExportHTML(''); }"
		currentWV.Eval(js)
	})

	// Wait for callback (or timeout) without blocking the WebView main thread.
	// Since exportCh is buffered (cap=1), this goroutine is independent
	// of wv.Eval and won't cause deadlock.
	go func() {
		select {
		case r := <-exportCh:
			exportCh = nil
			if r.err != nil {
				if r.err.Error() == "cancelled" {
					fmt.Fprintln(os.Stderr, "[ExportHTML] cancelled")
				} else {
					fmt.Fprintf(os.Stderr, "[ExportHTML] Error: %v\n", r.err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "[ExportHTML] Saved: %s\n", r.path)
			}
		case <-exportDone:
			exportCh = nil
		}
	}()
}

// ─── exportPDF ──────────────────────────────────────────────────────────────

func exportPDF() {
	if currentWV == nil {
		fmt.Fprintln(os.Stderr, "[Export] currentWV is nil, returning")
		return
	}
	if exportCh != nil {
		fmt.Fprintln(os.Stderr, "[Export] export already in progress, skipping")
		return
	}
	exportCh = make(chan exportResult, 1)

	wv.Dispatch(func() {
		js := "try{ window.onExportPDF && window.onExportPDF(" + buildExportJS() + "); } catch(e){ window.onExportPDF && window.onExportPDF(''); }"
		currentWV.Eval(js)
	})

	go func() {
		select {
		case r := <-exportCh:
			exportCh = nil
			if r.err != nil {
				if r.err.Error() == "cancelled" {
					fmt.Fprintln(os.Stderr, "[ExportPDF] cancelled")
				} else {
					fmt.Fprintf(os.Stderr, "[ExportPDF] Error: %v\n", r.err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "[ExportPDF] Saved: %s\n", r.path)
			}
		case <-exportDone:
			exportCh = nil
		}
	}()
}

// exportDone is closed when the app is shutting down.
var exportDone = make(chan struct{})
