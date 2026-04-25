#import <AppKit/AppKit.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>
#include <stdint.h>

void goDragDropCallback(const char* filePath);

// ──────────────────────────────────────────────────────────────────────────────
// DragDropWindowHelper: registers the main window for native drag & drop
// (as fallback for intra-app drags; Finder→App uses JS HTML5 drop + FileReader)
// Also provides visual feedback overlay during drag operations.
// ──────────────────────────────────────────────────────────────────────────────
@interface DragDropWindowHelper : NSObject <NSDraggingDestination>
@property (nonatomic, assign) NSWindow *targetWindow;
@property (nonatomic, strong) NSView *feedbackView;
@property (nonatomic, strong) NSTextField *dropLabel;
@property (nonatomic, assign) BOOL isDragging;
@end

@implementation DragDropWindowHelper

- (instancetype)initWithWindow:(NSWindow *)window {
    self = [super init];
    if (self) {
        _targetWindow = window;
        _isDragging = NO;

        [window registerForDraggedTypes:@[
            NSPasteboardTypeFileURL,
            NSFilenamesPboardType
        ]];

        // Create visual feedback view (hidden by default)
        NSView *cv = window.contentView;
        _feedbackView = [[NSView alloc] initWithFrame:cv.bounds];
        _feedbackView.wantsLayer = YES;
        _feedbackView.layer.backgroundColor = [[NSColor colorWithRed:0.035 green:0.376 blue:0.855 alpha:0.18] CGColor];
        _feedbackView.layer.opacity = 0.0;
        _feedbackView.layer.cornerRadius = 12;
        _feedbackView.layer.borderWidth = 3;
        _feedbackView.layer.borderColor = [[NSColor colorWithRed:0.035 green:0.376 blue:0.855 alpha:0.9] CGColor];
        _feedbackView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;

        _dropLabel = [[NSTextField alloc] initWithFrame:NSZeroRect];
        _dropLabel.stringValue = @"Drop .md file here";
        _dropLabel.font = [NSFont systemFontOfSize:22 weight:NSFontWeightSemibold];
        _dropLabel.textColor = [NSColor colorWithRed:0.035 green:0.376 blue:0.855 alpha:1.0];
        _dropLabel.alignment = NSTextAlignmentCenter;
        _dropLabel.backgroundColor = [[NSColor windowBackgroundColor] colorWithAlphaComponent:0.92];
        _dropLabel.wantsLayer = YES;
        _dropLabel.layer.cornerRadius = 10;
        _dropLabel.layer.borderWidth = 1;
        _dropLabel.layer.borderColor = [[NSColor colorWithRed:0.035 green:0.376 blue:0.855 alpha:0.4] CGColor];
        _dropLabel.translatesAutoresizingMaskIntoConstraints = NO;
        [_feedbackView addSubview:_dropLabel];
        [NSLayoutConstraint activateConstraints:@[
            [_dropLabel.centerXAnchor constraintEqualToAnchor:_feedbackView.centerXAnchor],
            [_dropLabel.centerYAnchor constraintEqualToAnchor:_feedbackView.centerYAnchor],
        ]];

        // Insert feedback view ABOVE all other content (including WKWebView)
        [cv addSubview:_feedbackView positioned:NSWindowAbove relativeTo:nil];
    }
    return self;
}

- (void)showFeedback:(BOOL)show {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSAnimationContext currentContext].duration = show ? 0.15 : 0.2;
        [[NSAnimationContext currentContext] setAllowsImplicitAnimation:YES];
        self.feedbackView.layer.opacity = show ? 1.0 : 0.0;
        self.dropLabel.alphaValue = show ? 1.0 : 0.0;
    });
}

// ── NSDraggingDestination ──
- (NSDragOperation)draggingEntered:(id<NSDraggingInfo>)sender {
    self.isDragging = YES;
    [self showFeedback:YES];
    return NSDragOperationCopy;
}

- (void)draggingExited:(id<NSDraggingInfo>)sender {
    self.isDragging = NO;
    [self showFeedback:NO];
}

- (BOOL)prepareForDragOperation:(id<NSDraggingInfo>)sender {
    return YES;
}

- (BOOL)performDragOperation:(id<NSDraggingInfo>)sender {
    self.isDragging = NO;
    [self showFeedback:NO];

    NSPasteboard *pboard = [sender draggingPasteboard];

    if ([pboard.types containsObject:NSPasteboardTypeFileURL]) {
        NSArray *urls = [pboard readObjectsForClasses:@[[NSURL class]]
                                             options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
        if (urls.count > 0) {
            NSURL *fileURL = urls[0];
            goDragDropCallback([fileURL.path UTF8String]);
            return YES;
        }
    }

    if ([pboard.types containsObject:NSFilenamesPboardType]) {
        NSArray *files = [pboard propertyListForType:NSFilenamesPboardType];
        if (files.count > 0) {
            goDragDropCallback([files[0] UTF8String]);
            return YES;
        }
    }

    return NO;
}

- (BOOL)wantsPeriodicDraggingUpdates { return NO; }

@end

// ─── C exports ────────────────────────────────────────────────────────────────

void* CreateDragDropOverlay(void* mainWindowPtr) {
    NSWindow *mainWindow = (__bridge NSWindow *)mainWindowPtr;
    if (!mainWindow) return NULL;

    DragDropWindowHelper *helper = [[DragDropWindowHelper alloc] initWithWindow:mainWindow];
    return (__bridge_retained void *)helper;
}

void ReleaseDragDropOverlay(void* overlayPtr) {
    if (!overlayPtr) return;
    id helper = (__bridge_transfer void *)overlayPtr;
    if ([helper respondsToSelector:@selector(targetWindow)]) {
        NSWindow *w = [(DragDropWindowHelper*)helper targetWindow];
        [w unregisterDraggedTypes];
    }
}
