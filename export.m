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

static PDFExportDelegate *_pdfDelegate = nil;

@implementation PDFExportDelegate

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    if (self.didExport) return;
    self.didExport = YES;

    NSLog(@"[PDF] didFinishNavigation called, waiting for layout...");

    // 1. 先等待 JS/高亮渲染完成
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.5 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
        [webView evaluateJavaScript:@"Math.max(document.body.scrollHeight, document.body.offsetHeight, document.documentElement.clientHeight, document.documentElement.scrollHeight, document.documentElement.offsetHeight)"
                  completionHandler:^(id result, NSError *error) {
            
            CGFloat height = [result floatValue];
            CGFloat width = webView.bounds.size.width;
            NSLog(@"[PDF] Calculated content height: %f", height);

            // 2. 暫時調整 WebView 的 Frame 高度，確保 WebKit 會繪製所有區域
            NSRect originalFrame = webView.frame;
            NSRect exportFrame = originalFrame;
            exportFrame.size.height = height;
            webView.frame = exportFrame;

            // 3. 調整 Frame 後稍等一下讓 Layout 生效
            dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.2 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
                WKPDFConfiguration *config = [[WKPDFConfiguration alloc] init];
                config.rect = CGRectMake(0, 0, width, height);

                [webView createPDFWithConfiguration:config
                                  completionHandler:^(NSData *pdfData, NSError *error) {
                    
                    // 恢復 Frame
                    webView.frame = originalFrame;

                    if (error || !pdfData) {
                        NSLog(@"[PDF] createPDF error: %@", error);
                        goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
                    } else {
                        NSError *writeErr = nil;
                        BOOL ok = [pdfData writeToFile:self.savePath options:NSDataWritingAtomic error:&writeErr];
                        if (ok) {
                            NSLog(@"[PDF] saved to %@", self.savePath);
                            goExportPDFResult([self.savePath UTF8String], NULL);
                        } else {
                            NSLog(@"[PDF] write error: %@", writeErr);
                            goExportPDFResult(NULL, [writeErr.localizedDescription UTF8String]);
                        }
                    }

                    // 恢復 Delegate 與內容
                    webView.navigationDelegate = self.previousDelegate;
                    if (self.originalHTML) {
                        [webView loadHTMLString:self.originalHTML baseURL:nil];
                    }
                    _pdfDelegate = nil;
                }];
            });
        }];
    });
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

void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8, const char *baseURLUTF8, void *windowPtr) {
    NSString *html = [NSString stringWithUTF8String:htmlUTF8];
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";
    NSURL *baseURL = baseURLUTF8 ? [NSURL URLWithString:[NSString stringWithUTF8String:baseURLUTF8]] : nil;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *win = (__bridge NSWindow *)windowPtr;
        WKWebView *wv = findWKWebView(win.contentView);
        if (!wv) {
            NSLog(@"[PDF] cannot find WKWebView");
            goExportPDFResult(NULL, "cannot find WKWebView");
            return;
        }

        NSLog(@"[PDF] found WKWebView: %@, baseURL: %@", wv, baseURL);

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
            NSLog(@"[PDF] calling loadHTMLString with baseURL: %@", baseURL);
            [wv loadHTMLString:html baseURL:baseURL];
            NSLog(@"[PDF] loadHTMLString called");
        }];
    });
}
