#import <AppKit/AppKit.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>
#include <stdint.h>

// Forward-declare Go-exported callbacks (cgo generates these in _cgo_export.h)
void goExportHTMLResult(const char *path, const char *errorMsg);
void goExportPDFResult(const char *path, const char *errorMsg);

// ──────────────────────────────────────────────────────────────────────────────
// ExportHelper: hidden WKWebView + NSSavePanel for PDF/HTML export.
// PDF uses WKWebView.createPDF for pixel-perfect rendering.
// ──────────────────────────────────────────────────────────────────────────────
@interface ExportHelper : NSObject <WKNavigationDelegate>

@property (nonatomic, strong) WKWebView *hiddenWV;
@property (nonatomic, copy) NSString *pdfSavePath;

- (void)exportHTML:(NSString *)html defaultName:(NSString *)defaultName;
- (void)exportPDF:(NSString *)html defaultName:(NSString *)defaultName;

@end

@implementation ExportHelper

- (instancetype)init {
    self = [super init];
    if (self) {
        WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];

        NSWindow *hiddenWindow = [[NSWindow alloc]
            initWithContentRect:NSMakeRect(-10000, -10000, 1200, 800)
            styleMask:NSWindowStyleMaskBorderless
            backing:NSBackingStoreBuffered defer:YES];
        hiddenWindow.level = NSNormalWindowLevel;
        hiddenWindow.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorStationary;

        _hiddenWV = [[WKWebView alloc] initWithFrame:NSMakeRect(0, 0, 1200, 800)
                                        configuration:config];
        _hiddenWV.navigationDelegate = self;
        [hiddenWindow.contentView addSubview:_hiddenWV];
    }
    return self;
}

#pragma mark - HTML Export

- (void)exportHTML:(NSString *)html defaultName:(NSString *)defaultName {
    NSSavePanel *panel = [NSSavePanel savePanel];
    panel.nameFieldStringValue = [defaultName stringByAppendingPathExtension:@"html"];
    panel.canCreateDirectories = YES;
    panel.message = @"選擇匯出 HTML 的位置";

    [panel beginSheetModalForWindow:[NSApp keyWindow] completionHandler:^(NSInteger result) {
        if (result != NSModalResponseOK) {
            goExportHTMLResult(NULL, "cancelled");
            return;
        }

        NSURL *url = panel.URL;
        NSError *writeErr = nil;
        NSData *data = [html dataUsingEncoding:NSUTF8StringEncoding];
        BOOL ok = [data writeToURL:url options:NSDataWritingAtomic error:&writeErr];

        if (ok) {
            goExportHTMLResult([url.path UTF8String], NULL);
        } else {
            goExportHTMLResult(NULL, [writeErr.localizedDescription UTF8String]);
        }
    }];
}

#pragma mark - PDF Export

- (void)exportPDF:(NSString *)html defaultName:(NSString *)defaultName {
    NSSavePanel *panel = [NSSavePanel savePanel];
    panel.nameFieldStringValue = [defaultName stringByAppendingPathExtension:@"pdf"];
    panel.canCreateDirectories = YES;
    panel.message = @"選擇匯出 PDF 的位置";

    [panel beginSheetModalForWindow:[NSApp keyWindow] completionHandler:^(NSInteger result) {
        if (result != NSModalResponseOK) {
            goExportPDFResult(NULL, "cancelled");
            return;
        }

        self.pdfSavePath = [panel.URL path];
        [self.hiddenWV loadHTMLString:html baseURL:nil];
    }];
}

#pragma mark - WKNavigationDelegate (PDF)

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    if (self.pdfSavePath == nil) return;

    NSString *savePath = self.pdfSavePath;
    self.pdfSavePath = nil;

    [webView createPDFWithConfiguration:[self pdfConfig]
                      completionHandler:^(NSData *pdfData, NSError *error) {
        if (error || !pdfData) {
            goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
            return;
        }

        NSError *writeErr = nil;
        BOOL ok = [pdfData writeToFile:savePath options:NSDataWritingAtomic error:&writeErr];

        if (ok) {
            goExportPDFResult([savePath UTF8String], NULL);
        } else {
            goExportPDFResult(NULL, [writeErr.localizedDescription UTF8String]);
        }
    }];
}

- (void)webView:(WKWebView *)webView didFailNavigation:(WKNavigation *)navigation withError:(NSError *)error {
    if (self.pdfSavePath != nil) {
        self.pdfSavePath = nil;
        goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
    }
}

- (WKPDFConfiguration *)pdfConfig {
    return [[WKPDFConfiguration alloc] init];
}

@end

// ─── Singleton ───────────────────────────────────────────────────────────────

static ExportHelper *_sharedExportHelper = nil;
static dispatch_once_t _exportOnceToken;

ExportHelper *GetExportHelper(void) {
    dispatch_once(&_exportOnceToken, ^{
        _sharedExportHelper = [[ExportHelper alloc] init];
    });
    return _sharedExportHelper;
}

// ─── C exports ─────────────────────────────────────────────────────────────

// ExportHTML: shows save dialog, writes HTML string.
// Calls goExportHTMLResult(path, NULL) on success or goExportHTMLResult(NULL, error) on failure.
void ExportHTML(const char *htmlUTF8, const char *defaultNameUTF8) {
    ExportHelper *helper = GetExportHelper();
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";
    dispatch_async(dispatch_get_main_queue(), ^{
        [helper exportHTML:html defaultName:name];
    });
}

// ExportPDF: shows save dialog, renders HTML→PDF via WKWebView.createPDF.
// Calls goExportPDFResult(path, NULL) on success or goExportPDFResult(NULL, error) on failure.
void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8) {
    ExportHelper *helper = GetExportHelper();
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";
    dispatch_async(dispatch_get_main_queue(), ^{
        [helper exportPDF:html defaultName:name];
    });
}
