package main

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11 -x objective-c++ -fobjc-arc
#cgo darwin LDFLAGS: -framework AppKit

#include <stdlib.h>
#include <stdint.h>

extern void goMenuCallback(int menuID);
void SetupMainMenu(void);
void UpdateMenuLanguageTitles(const char *lang);
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
)

var menuCallback func(int)

//export goMenuCallback
func goMenuCallback(menuID C.int) {
	if menuCallback != nil {
		menuCallback(int(menuID))
	}
}

// SetupMenu registers a callback for menu item clicks and sets up the native NSMenu.
// UpdateMenuLanguage rebuilds the menu with localized titles.
func UpdateMenuLanguage(lang string) {
	C.UpdateMenuLanguageTitles(C.CString(lang))
}

func SetupMenu(callback func(int)) {
	menuCallback = callback
	C.SetupMainMenu()
}

// Unused dummy — keeps the unsafe import from being stripped by go vet/compiler.
var _ = unsafe.Pointer(nil)
