#import <AppKit/AppKit.h>
#import <WebKit/WebKit.h>
#include <stdlib.h>
#include <stdint.h>

void goExportHTMLResult(const char *path, const char *errorMsg);
void goExportPDFResult(const char *path, const char *errorMsg);

static WKWebView *findWKWebView(NSView *view) {
    if ([view isKindOfClass:[WKWebView class]]) return (WKWebView *)view;
    for (NSView *sub in view.subviews) {
        WKWebView *found = findWKWebView(sub);
        if (found) return found;
    }
    return nil;
}

// ─── HTML Export ─────────────────────────────────────────────────────────────

void ExportHTML(const char *htmlUTF8, const char *defaultNameUTF8) {
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";

    dispatch_async(dispatch_get_main_queue(), ^{
        NSSavePanel *panel = [NSSavePanel savePanel];
        panel.nameFieldStringValue = [name stringByAppendingPathExtension:@"html"];
        panel.canCreateDirectories = YES;
        panel.message = @"選擇匯出 HTML 的位置";

        NSInteger result = [panel runModal];
        if (result != NSModalResponseOK) {
            goExportHTMLResult(NULL, "cancelled");
            return;
        }

        NSError *writeErr = nil;
        NSData *data = [html dataUsingEncoding:NSUTF8StringEncoding];
        BOOL ok = [data writeToURL:panel.URL options:NSDataWritingAtomic error:&writeErr];
        if (ok) {
            goExportHTMLResult([panel.URL.path UTF8String], NULL);
        } else {
            goExportHTMLResult(NULL, [writeErr.localizedDescription UTF8String]);
        }
    });
}

// ─── PDF Export ──────────────────────────────────────────────────────────────

@interface PDFExportDelegate : NSObject <WKNavigationDelegate>
@property (nonatomic, copy) NSString *savePath;
@property (nonatomic, copy) NSString *originalHTML;
@property (nonatomic, assign) WKWebView *webView;
@property (nonatomic, assign) id<WKNavigationDelegate> previousDelegate;
@property (nonatomic, assign) BOOL didExport;
@end

@implementation PDFExportDelegate

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    if (self.didExport) return;
    self.didExport = YES;

    // 還原 delegate
    webView.navigationDelegate = self.previousDelegate;

    NSLog(@"[PDF] didFinishNavigation called OK, webView=%@", webView);

    // 還原原本頁面
    if (self.originalHTML) {
        [webView loadHTMLString:self.originalHTML baseURL:nil];
    }

    goExportPDFResult(NULL, "test - not saving");
}

- (void)webView:(WKWebView *)webView didFailNavigation:(WKNavigation *)navigation withError:(NSError *)error {
    if (self.didExport) return;
    self.didExport = YES;

    NSLog(@"[PDF] didFailNavigation: %@", error);
    webView.navigationDelegate = self.previousDelegate;
    if (self.originalHTML) {
        [webView loadHTMLString:self.originalHTML baseURL:nil];
    }
    goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
}

@end

static PDFExportDelegate *_pdfDelegate = nil;

void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8, void *windowPtr) {
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";

    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *win = (__bridge NSWindow *)windowPtr;
        WKWebView *wv = findWKWebView(win.contentView);
        if (!wv) {
            NSLog(@"[PDF] cannot find WKWebView");
            goExportPDFResult(NULL, "cannot find WKWebView");
            return;
        }

        NSLog(@"[PDF] found WKWebView: %@", wv);

        NSSavePanel *panel = [NSSavePanel savePanel];
        panel.nameFieldStringValue = [name stringByAppendingPathExtension:@"pdf"];
        panel.canCreateDirectories = YES;
        panel.message = @"選擇匯出 PDF 的位置";

        NSInteger result = [panel runModal];
        if (result != NSModalResponseOK) {
            goExportPDFResult(NULL, "cancelled");
            return;
        }

        PDFExportDelegate *delegate = [[PDFExportDelegate alloc] init];
        delegate.savePath = panel.URL.path;
        delegate.webView = wv;
        delegate.previousDelegate = wv.navigationDelegate;
        delegate.didExport = NO;

        [wv evaluateJavaScript:@"document.documentElement.outerHTML"
             completionHandler:^(id result, NSError *error) {
            delegate.originalHTML = [result isKindOfClass:[NSString class]] ? result : nil;
            NSLog(@"[PDF] got originalHTML length=%lu", (unsigned long)[delegate.originalHTML length]);
            _pdfDelegate = delegate;
            wv.navigationDelegate = delegate;
            NSLog(@"[PDF] calling loadHTMLString...");
            [wv loadHTMLString:html baseURL:nil];
            NSLog(@"[PDF] loadHTMLString called");
        }];
    });
}
