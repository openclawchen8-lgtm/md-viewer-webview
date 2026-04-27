#import <AppKit/AppKit.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>
#include <stdint.h>

void goExportHTMLResult(const char *path, const char *errorMsg);
void goExportPDFResult(const char *path, const char *errorMsg);

// ──────────────────────────────────────────────────────────────────────────────
// ExportHelper: hidden WKWebView + NSSavePanel for PDF/HTML export.
// All ObjC work is dispatched to the main queue. C exports return immediately.
// ──────────────────────────────────────────────────────────────────────────────
@interface ExportHelper : NSObject <WKNavigationDelegate>

@property (nonatomic, strong) WKWebView *hiddenWV;
@property (nonatomic, copy) NSString *pdfSavePath;

- (void)doExportHTML:(NSString *)html name:(NSString *)name;
- (void)doExportPDF:(NSString *)html name:(NSString *)name;

@end

@implementation ExportHelper

- (instancetype)init {
    self = [super init];
    if (self) {
        NSWindow *hiddenWindow = [[NSWindow alloc]
            initWithContentRect:NSMakeRect(-10000, -10000, 1200, 800)
            styleMask:NSWindowStyleMaskBorderless
            backing:NSBackingStoreBuffered defer:YES];
        hiddenWindow.level = NSNormalWindowLevel;
        hiddenWindow.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces
                                        | NSWindowCollectionBehaviorStationary;

        WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
        _hiddenWV = [[WKWebView alloc] initWithFrame:NSMakeRect(0, 0, 1200, 800)
                                        configuration:config];
        _hiddenWV.navigationDelegate = self;
        [hiddenWindow.contentView addSubview:_hiddenWV];
    }
    return self;
}

- (void)doExportHTML:(NSString *)html name:(NSString *)name {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSSavePanel *panel = [NSSavePanel savePanel];
        panel.nameFieldStringValue = [name stringByAppendingPathExtension:@"html"];
        panel.canCreateDirectories = YES;
        panel.message = @"選擇匯出 HTML 的位置";

        NSInteger result = [panel runModal];

        if (result != NSModalResponseOK) {
            goExportHTMLResult(NULL, "cancelled");
        } else {
            NSURL *url = panel.URL;
            NSError *writeErr = nil;
            NSData *data = [html dataUsingEncoding:NSUTF8StringEncoding];
            BOOL ok = [data writeToURL:url options:NSDataWritingAtomic error:&writeErr];
            if (ok) {
                goExportHTMLResult([url.path UTF8String], NULL);
            } else {
                goExportHTMLResult(NULL, [writeErr.localizedDescription UTF8String]);
            }
        }
    });
}

- (void)doExportPDF:(NSString *)html name:(NSString *)name {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSSavePanel *panel = [NSSavePanel savePanel];
        panel.nameFieldStringValue = [name stringByAppendingPathExtension:@"pdf"];
        panel.canCreateDirectories = YES;
        panel.message = @"選擇匯出 PDF 的位置";

        NSInteger result = [panel runModal];

        if (result != NSModalResponseOK) {
            goExportPDFResult(NULL, "cancelled");
            return;
        }

        self.pdfSavePath = [panel.URL path];
        [self.hiddenWV loadHTMLString:html baseURL:nil];
    });
}

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    if (self.pdfSavePath == nil) return;

    NSString *savePath = self.pdfSavePath;
    self.pdfSavePath = nil;

    [webView createPDFWithConfiguration:[self pdfConfig]
                      completionHandler:^(NSData *pdfData, NSError *error) {
        if (error || !pdfData) {
            goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
        } else {
            NSError *writeErr = nil;
            BOOL ok = [pdfData writeToFile:savePath options:NSDataWritingAtomic error:&writeErr];
            if (ok) {
                goExportPDFResult([savePath UTF8String], NULL);
            } else {
                goExportPDFResult(NULL, [writeErr.localizedDescription UTF8String]);
            }
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

static ExportHelper *_sharedExportHelper = nil;
static dispatch_once_t _exportOnceToken;

ExportHelper *GetExportHelper(void) {
    dispatch_once(&_exportOnceToken, ^{
        // Synchronous init so hiddenWV is ready before first call returns.
        _sharedExportHelper = [[ExportHelper alloc] init];
    });
    return _sharedExportHelper;
}

void ExportHTML(const char *htmlUTF8, const char *defaultNameUTF8) {
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";
    [GetExportHelper() doExportHTML:html name:name];
}

void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8) {
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";
    [GetExportHelper() doExportPDF:html name:name];
}
