package main

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11 -x objective-c++ -fobjc-arc
#cgo darwin LDFLAGS: -framework WebKit -ldl -framework AppKit

#include <stdlib.h>
#include <stdint.h>

void* CreateDragDropOverlay(void* mainWindow);
void  ReleaseDragDropOverlay(void* overlay);
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

//export goDragDropCallback
func goDragDropCallback(filePath *C.char) {
	path := C.GoString(filePath)
	fmt.Fprintf(os.Stderr, "[DragDrop Go] path: %s\n", path)
	if dragDropCallback != nil {
		dragDropCallback(path)
	}
}

var dragDropCallback func(string)

// CreateDragDropOverlay creates a transparent overlay window for receiving drag & drop
func CreateDragDropOverlay(mainWindow unsafe.Pointer) unsafe.Pointer {
	if mainWindow == nil {
		return nil
	}
	result := C.CreateDragDropOverlay(mainWindow)
	return result
}

// ReleaseDragDropOverlay hides and releases the overlay
func ReleaseDragDropOverlay(overlay unsafe.Pointer) {
	if overlay != nil {
		C.ReleaseDragDropOverlay(overlay)
	}
}

// RegisterDragDropCallback registers a callback to be called when files are dropped
func RegisterDragDropCallback(callback func(string)) {
	dragDropCallback = callback
}
