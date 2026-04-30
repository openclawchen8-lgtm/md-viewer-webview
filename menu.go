package main

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11 -x objective-c++ -fobjc-arc
#cgo darwin LDFLAGS: -framework AppKit

#include <stdlib.h>
#include <stdint.h>

extern void goMenuCallback(int menuID);
void SetupMainMenu(void);
void UpdateMenuLanguageTitles(const char *lang);
void UpdateRecentFilesMenu(const char **files, int count);
void SetWindowFrame(void *windowPtr, int x, int y, int width, int height);
void GetWindowSize(void *windowPtr, int *width, int *height);
void GetWindowPosition(void *windowPtr, int *x, int *y);
*/
import "C"
import (
	"unsafe"
)

// Menu item IDs
const (
	MenuAbout         = 1
	MenuPreferences   = 2
	MenuOpen          = 3
	MenuReload        = 4
	MenuQuit          = 5
	MenuZoomIn        = 6
	MenuZoomOut       = 7
	MenuZoomReset     = 8
	MenuToggleSidebar = 9
	MenuAboutHelp     = 10
	MenuFullscreen    = 11
	MenuExportHTML    = 12
	MenuExportPDF     = 13
)

var menuCallback func(int)
var openFileCallback func(string)
var removeRecentFileCallback func(string)
var recentFilesCache []*C.char

//export goMenuCallback
func goMenuCallback(menuID C.int) {
	if menuCallback != nil {
		menuCallback(int(menuID))
	}
}

//export goOpenFileCallback
func goOpenFileCallback(path *C.char) {
	if openFileCallback != nil {
		openFileCallback(C.GoString(path))
	}
}

//export goRemoveRecentFileCallback
func goRemoveRecentFileCallback(path *C.char) {
	if removeRecentFileCallback != nil {
		removeRecentFileCallback(C.GoString(path))
	}
}

// RegisterOpenFileCallback registers a callback for macOS "Open File" events.
func RegisterOpenFileCallback(callback func(string)) {
	openFileCallback = callback
}

// RegisterRemoveRecentFileCallback registers a callback for removing recent files.
func RegisterRemoveRecentFileCallback(callback func(string)) {
	removeRecentFileCallback = callback
}

// SetupMenu registers a callback for menu item clicks and sets up the native NSMenu.
// UpdateMenuLanguage rebuilds the menu with localized titles.
func UpdateMenuLanguage(lang string) {
	C.UpdateMenuLanguageTitles(C.CString(lang))
}

// UpdateRecentFiles updates the recent files menu with the given file list.
func UpdateRecentFiles(files []string) {
	if len(files) == 0 {
		C.UpdateRecentFilesMenu(nil, 0)
		return
	}
	// Convert to C array
	cFiles := make([]*C.char, len(files))
	for i, f := range files {
		cFiles[i] = C.CString(f)
	}
	C.UpdateRecentFilesMenu(&cFiles[0], C.int(len(files)))
	// Note: C strings are intentionally NOT freed here
	// They will be copied by Objective-C's stringWithUTF8String
	// But since we don't use them after this, we need to leak them or fix differently
	// Actually - let's keep them alive in a slice to prevent GC
	recentFilesCache = cFiles
}

func SetupMenu(callback func(int)) {
	menuCallback = callback
	C.SetupMainMenu()
}

// Unused dummy — keeps the unsafe import from being stripped by go vet/compiler.
var _ = unsafe.Pointer(nil)

// SetWindowFrame sets the window position and size, disabling macOS autosave.
func SetWindowFrame(windowPtr unsafe.Pointer, x, y, width, height int) {
	C.SetWindowFrame(windowPtr, C.int(x), C.int(y), C.int(width), C.int(height))
}

// GetWindowSize returns the current window size.
func GetWindowSize(windowPtr unsafe.Pointer) (width, height int) {
	var w, h C.int
	C.GetWindowSize(windowPtr, &w, &h)
	return int(w), int(h)
}

// GetWindowPosition returns the current window position.
func GetWindowPosition(windowPtr unsafe.Pointer) (x, y int) {
	var xPos, yPos C.int
	C.GetWindowPosition(windowPtr, &xPos, &yPos)
	return int(xPos), int(yPos)
}
