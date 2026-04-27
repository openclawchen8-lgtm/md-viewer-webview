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
            goExportPDFResult(NULL, "test - not saving"); // Wait, this should call goExportHTMLResult
            goExportHTMLResult([panel.URL.path UTF8String], NULL);
        } else {
            goExportHTMLResult(NULL, [writeErr.localizedDescription UTF8String]);
        }
    });
}

// ─── PDF Export ──────────────────────────────────────────────────────────────

void ExportPDF(const char *htmlUTF8, const char *defaultNameUTF8, const char *baseURLUTF8, void *windowPtr) {
    NSString *name = defaultNameUTF8 ? [NSString stringWithUTF8String:defaultNameUTF8] : @"untitled";

    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *win = (__bridge NSWindow *)windowPtr;
        WKWebView *wv = findWKWebView(win.contentView);
        if (!wv) {
            goExportPDFResult(NULL, "cannot find WKWebView");
            return;
        }

        NSSavePanel *panel = [NSSavePanel savePanel];
        panel.nameFieldStringValue = [name stringByAppendingPathExtension:@"pdf"];
        panel.canCreateDirectories = YES;
        panel.message = @"選擇匯出 PDF 的位置";

        NSInteger result = [panel runModal];
        if (result != NSModalResponseOK) {
            goExportPDFResult(NULL, "cancelled");
            return;
        }

        NSString *savePath = panel.URL.path;

        // 1. 準備環境：將縮放強制設為 1.0 以利計算精確高度
        [wv evaluateJavaScript:@"(function(){\
            var oldZoom = document.body.style.zoom || '1.0';\
            var oldScrollY = window.scrollY;\
            document.body.classList.add('is-exporting');\
            document.body.style.zoom = '1.0';\
            window.scrollTo(0,0);\
            var h = Math.max(document.body.scrollHeight, document.body.offsetHeight, document.documentElement.scrollHeight);\
            return {zoom: oldZoom, height: h, scrollY: oldScrollY};\
        })()" completionHandler:^(id result, NSError *error) {
            
            NSDictionary *data = (NSDictionary *)result;
            NSString *oldZoom = data[@"zoom"];
            CGFloat contentHeight = [data[@"height"] floatValue];
            NSNumber *oldScrollY = data[@"scrollY"];
            
            NSRect originalFrame = wv.frame;
            NSRect exportFrame = originalFrame;
            exportFrame.size.height = contentHeight;
            
            // 2. 暫時調整 Frame
            wv.frame = exportFrame;

            // 3. 確保 WebKit 有時間繪製
            dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.4 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
                WKPDFConfiguration *config = [[WKPDFConfiguration alloc] init];
                config.rect = CGRectMake(0, 0, originalFrame.size.width, contentHeight);

                [wv createPDFWithConfiguration:config completionHandler:^(NSData *pdfData, NSError *error) {
                    
                    // 4. 無論如何都要恢復環境，防止 UI 消失
                    wv.frame = originalFrame;
                    NSString *restoreJS = [NSString stringWithFormat:@"\
                        document.body.classList.remove('is-exporting');\
                        document.body.style.zoom = '%@';\
                        window.scrollTo(0, %@);", oldZoom, oldScrollY];
                    [wv evaluateJavaScript:restoreJS completionHandler:nil];

                    if (error || !pdfData) {
                        goExportPDFResult(NULL, [error.localizedDescription UTF8String]);
                    } else {
                        if ([pdfData writeToFile:savePath options:NSDataWritingAtomic error:nil]) {
                            goExportPDFResult([savePath UTF8String], NULL);
                        } else {
                            goExportPDFResult(NULL, "failed to write to file");
                        }
                    }
                }];
            });
        }];
    });
}
