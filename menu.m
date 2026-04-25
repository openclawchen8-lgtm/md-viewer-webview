#import <AppKit/AppKit.h>
#include <stdlib.h>
#include <stdint.h>

void goMenuCallback(int menuID);

@interface MDAppDelegate : NSObject <NSApplicationDelegate>
- (void)setupMenuWithLang:(NSString *)lang;
@property (nonatomic, strong) NSString *currentLang;
@end

@implementation MDAppDelegate

- (void)setupMenuWithLang:(NSString *)lang {
    self.currentLang = lang.length > 0 ? lang : @"zhTW";

    // i18n dictionary: key -> array [en, zhTW, zhCN, ja, ko]
    NSDictionary *tr = @{
        @"appAbout":      @[@"About md-viewer", @"關於 md-viewer", @"关于 md-viewer", @"md-viewer について", @"md-viewer 정보"],
        @"appPref":       @[@"Preferences...", @"偏好設定...", @"偏好设置...", @"環境設定...", @"설정..."],
        @"appQuit":       @[@"Quit md-viewer", @"結束 md-viewer", @"结束 md-viewer", @"md-viewer を終了", @"md-viewer 종료"],
        @"fileOpen":      @[@"Open...", @"開啟檔案...", @"打开文件...", @"ファイルを開く...", @"파일 열기..."],
        @"fileReload":    @[@"Reload", @"重新載入", @"重新载入", @"再読み込み", @"새로고침"],
        @"viewIn":        @[@"Zoom In", @"放大", @"放大", @"拡大", @"확대"],
        @"viewOut":       @[@"Zoom Out", @"縮小", @"缩小", @"縮小", @"축소"],
        @"viewReset":     @[@"Actual Size", @"實際大小", @"实际大小", @"実際のサイズ", @"실제 크기"],
        @"viewFull":      @[@"Toggle Full Screen", @"切換全螢幕", @"切换全屏", @"フルスクリーン切替", @"전체 화면 전환"],
        @"helpAbout":     @[@"About md-viewer", @"關於 md-viewer", @"关于 md-viewer", @"md-viewer について", @"md-viewer 정보"],
    };
    NSArray *langs = @[@"en", @"zhTW", @"zhCN", @"ja", @"ko"];
    NSUInteger li = [langs indexOfObject:self.currentLang];
    if (li == NSNotFound) li = 0;

    NSString *(^t)(NSString *) = ^(NSString *key) {
        NSArray *arr = tr[key];
        return arr ? arr[li] : key;
    };

    // Rebuild main menu
    NSMenu *mainMenu = [[NSMenu alloc] init];

    // App menu
    NSMenuItem *appItem = [[NSMenuItem alloc] init];
    NSMenu *appMenu = [[NSMenu alloc] init];
    [appMenu addItemWithTitle:t(@"appAbout") action:@selector(orderFrontStandardAboutPanel:) keyEquivalent:@""];
    [appMenu addItem:[NSMenuItem separatorItem]];
    [appMenu addItemWithTitle:t(@"appPref") action:@selector(menuPreferences) keyEquivalent:@","];
    [appMenu addItem:[NSMenuItem separatorItem]];
    [appMenu addItemWithTitle:t(@"appQuit") action:@selector(menuQuit) keyEquivalent:@"q"];
    [appItem setSubmenu:appMenu];
    [mainMenu addItem:appItem];

    // File menu
    NSMenuItem *fileItem = [[NSMenuItem alloc] init];
    NSMenu *fileMenu = [[NSMenu alloc] init];
    [fileMenu addItemWithTitle:t(@"fileOpen") action:@selector(menuOpen) keyEquivalent:@"o"];
    [fileMenu addItemWithTitle:t(@"fileReload") action:@selector(menuReload) keyEquivalent:@"r"];
    [fileItem setSubmenu:fileMenu];
    [mainMenu addItem:fileItem];

    // View menu
    NSMenuItem *viewItem = [[NSMenuItem alloc] init];
    NSMenu *viewMenu = [[NSMenu alloc] init];
    [viewMenu addItemWithTitle:t(@"viewIn") action:@selector(menuZoomIn) keyEquivalent:@"="];
    [viewMenu addItemWithTitle:t(@"viewOut") action:@selector(menuZoomOut) keyEquivalent:@"-"];
    [viewMenu addItemWithTitle:t(@"viewReset") action:@selector(menuZoomReset) keyEquivalent:@"0"];
    [viewMenu addItem:[NSMenuItem separatorItem]];
    [viewMenu addItemWithTitle:t(@"viewFull") action:@selector(menuFullscreen) keyEquivalent:@"f"];
    [viewItem setSubmenu:viewMenu];
    [mainMenu addItem:viewItem];

    // Help menu
    NSMenuItem *helpItem = [[NSMenuItem alloc] init];
    NSMenu *helpMenu = [[NSMenu alloc] init];
    [helpMenu addItemWithTitle:t(@"helpAbout") action:@selector(orderFrontStandardAboutPanel:) keyEquivalent:@""];
    [helpItem setSubmenu:helpMenu];
    [mainMenu addItem:helpItem];

    [NSApp setMainMenu:mainMenu];
}

- (void)menuPreferences { goMenuCallback(2); }
- (void)menuOpen        { goMenuCallback(3); }
- (void)menuReload     { goMenuCallback(4); }
- (void)menuQuit        { [NSApp terminate:nil]; }
- (void)menuZoomIn      { goMenuCallback(6); }
- (void)menuZoomOut     { goMenuCallback(7); }
- (void)menuZoomReset   { goMenuCallback(8); }
- (void)menuFullscreen  {
    NSWindow *window = [NSApp keyWindow];
    if (window && [window respondsToSelector:@selector(toggleFullScreen:)]) {
        [window toggleFullScreen:nil];
    }
}

@end

static MDAppDelegate *_sharedDelegate = nil;
static NSString *currentMenuLang = @"";

void UpdateMenuLanguageTitles(const char *lang) {
    currentMenuLang = [NSString stringWithUTF8String:lang];
    dispatch_async(dispatch_get_main_queue(), ^{
        if (_sharedDelegate) {
            [(MDAppDelegate *)_sharedDelegate setupMenuWithLang:currentMenuLang];
        }
    });
}

void SetupMainMenu(void) {
    static dispatch_once_t once;
    dispatch_once(&once, ^{
        _sharedDelegate = [[MDAppDelegate alloc] init];
        [NSApp setDelegate:_sharedDelegate];
        NSString *lang = currentMenuLang.length > 0 ? currentMenuLang : @"zhTW";
        [(MDAppDelegate *)_sharedDelegate setupMenuWithLang:lang];
    });
}
