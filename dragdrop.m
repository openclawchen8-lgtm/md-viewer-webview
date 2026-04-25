#import <AppKit/AppKit.h>
#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>
#include <stdlib.h>
#include <stdint.h>

void goDragDropCallback(const char* filePath);

@interface DragDropView : NSView <NSDraggingDestination>
@property (nonatomic, copy) void (^onFileDrop)(NSString *filePath);
@property (nonatomic, assign) BOOL isDragging;
@end

@implementation DragDropView

- (instancetype)initWithFrame:(NSRect)frameRect {
    self = [super initWithFrame:frameRect];
    if (self) {
        _isDragging = NO;
        [self registerForDraggedTypes:@[NSPasteboardTypeFileURL, NSFilenamesPboardType]];
    }
    return self;
}

// Pass through all mouse events to underlying view when not dragging
- (NSView *)hitTest:(NSPoint)point {
    if (!self.isDragging) {
        // Not dragging - pass event down the responder chain (to WebView below)
        return nil;
    }
    return self; // During drag, capture events
}

- (NSDragOperation)draggingEntered:(id<NSDraggingInfo>)sender {
    self.isDragging = YES;
    return NSDragOperationCopy;
}

- (void)draggingExited:(id<NSDraggingInfo>)sender {
    self.isDragging = NO;
}

- (BOOL)prepareForDragOperation:(id<NSDraggingInfo>)sender {
    NSLog(@"[DragDrop] prepareForDragOperation");
    return YES;
}

- (BOOL)performDragOperation:(id<NSDraggingInfo>)sender {
    self.isDragging = NO;
    NSPasteboard *pboard = [sender draggingPasteboard];
    
    // Try NSPasteboardTypeFileURL first (modern API)
    if ([pboard.types containsObject:NSPasteboardTypeFileURL]) {
        NSArray *urls = [pboard readObjectsForClasses:@[[NSURL class]] 
                                             options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
        if (urls.count > 0) {
            NSURL *url = urls[0];
            goDragDropCallback([url.path UTF8String]);
            return YES;
        }
    }
    
    // Fallback to NSFilenamesPboardType (legacy but more compatible)
    if ([pboard.types containsObject:NSFilenamesPboardType]) {
        NSArray *files = [pboard propertyListForType:NSFilenamesPboardType];
        if (files.count > 0) {
            goDragDropCallback([files[0] UTF8String]);
            return YES;
        }
    }
    
    return NO;
}

- (BOOL)wantsPeriodicDraggingUpdates {
    return NO;
}

@end

void* EnableDragDrop(NSWindow *window) {
    NSLog(@"[DragDrop] EnableDragDrop called with window: %@", window);
    
    if (!window) {
        NSLog(@"[DragDrop] Window is NULL");
        return NULL;
    }
    
    NSView *contentView = window.contentView;
    if (!contentView) {
        NSLog(@"[DragDrop] ContentView is NULL");
        return NULL;
    }
    
    NSLog(@"[DragDrop] ContentView frame: %@", NSStringFromRect(contentView.frame));
    
    DragDropView *dropView = [[DragDropView alloc] initWithFrame:contentView.bounds];
    dropView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    
    // Make the view transparent and click-through
    dropView.wantsLayer = YES;
    dropView.layer.backgroundColor = [[NSColor clearColor] CGColor];
    dropView.layer.opacity = 0.0; // Invisible but receives events
    
    // Add as TOPMOST sibling (above WebView)
    [contentView addSubview:dropView];
    
    NSLog(@"[DragDrop] DropView added as subview");
    
    return (__bridge void *)dropView;
}
