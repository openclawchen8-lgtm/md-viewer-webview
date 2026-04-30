#import <AppKit/AppKit.h>
#include <stdlib.h>
#include <stdint.h>

void goMenuCallback(int menuID);
void goOpenFileCallback(const char *path);

@interface MDAppDelegate : NSObject <NSApplicationDelegate>
- (void)setupMenuWithLang:(NSString *)lang;
@property (nonatomic, strong) NSString *currentLang;
@property (nonatomic, strong) NSString *pendingFile;
@end

@implementation MDAppDelegate

- (void)applicationWillFinishLaunching:(NSNotification *)notification {
    // This helps ensure we catch openFile: events during startup
}

- (BOOL)application:(NSApplication *)sender openFile:(NSString *)filename {
    if (filename) {
        self.pendingFile = filename;
        goOpenFileCallback([filename UTF8String]);
        return YES;
    }
    return NO;
}

- (void)setupMenuWithLang:(NSString *)lang {
    self.currentLang = lang.length > 0 ? lang : @"zhTW";

    // i18n dictionary: key -> array [en, zhTW, zhCN, ja, ko]
    NSDictionary *tr = @{
        @"menuFile":      @[@"File", @"檔案", @"文件", @"ファイル", @"파일"],
        @"menuView":      @[@"View", @"顯示", @"显示", @"表示", @"보기"],
        @"menuExport":    @[@"Export", @"匯出", @"导出", @"エクスポート", @"내보내기"],
        @"menuHelp":      @[@"Help", @"說明", @"幫助", @"ヘルプ", @"도움말"],
        @"appAbout":      @[@"About md-viewer", @"關於 md-viewer", @"关于 md-viewer", @"md-viewer について", @"md-viewer 정보"],
        @"appPref":       @[@"Preferences...", @"偏好設定...", @"偏好设置...", @"環境設定...", @"설정..."],
        @"appQuit":       @[@"Quit md-viewer", @"結束 md-viewer", @"结束 md-viewer", @"md-viewer を終了", @"md-viewer 종료"],
        @"fileOpen":      @[@"Open...", @"開啟檔案...", @"打开文件...", @"ファイルを開く...", @"파일 열기..."],
        @"fileReload":    @[@"Reload", @"重新載入", @"重新载入", @"再読み込み", @"새로고침"],
        @"fileExportHTML":@[@"Export as HTML...", @"匯出為 HTML...", @"导出为 HTML...", @"HTML としてエクスポート...", @"HTML로 내보내기..."],
        @"fileExportPDF": @[@"Export as PDF...",  @"匯出為 PDF...",  @"导出为 PDF...",  @"PDF としてエクスポート...",  @"PDF로 내보내기..."],
        @"viewIn":        @[@"Zoom In", @"放大", @"放大", @"拡大", @"확대"],
        @"viewOut":       @[@"Zoom Out", @"縮小", @"缩小", @"縮小", @"축소"],
        @"viewReset":     @[@"Actual Size", @"實際大小", @"实际大小", @"實際のサイズ", @"실제 크기"],
        @"viewFull":      @[@"Toggle Full Screen", @"切換全螢幕", @"切换全屏", @"フルスクリーン切替", @"전체 화면 전환"],
        @"viewFocus":     @[@"Focus Mode", @"專注模式", @"专注模式", @"集中モード", @"집중 모드"],
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
    [appItem setTitle:@"md-viewer"];
    NSMenu *appMenu = [[NSMenu alloc] init];
    [appMenu addItemWithTitle:t(@"appAbout") action:@selector(orderFrontStandardAboutPanel:) keyEquivalent:@""];
    [appMenu addItem:[NSMenuItem separatorItem]];
    // ⌘, toggles settings panel (standard macOS shortcut)
    NSMenuItem *prefShortcut = [[NSMenuItem alloc] initWithTitle:t(@"appPref") action:@selector(menuPreferences) keyEquivalent:@","];
    [appMenu addItem:prefShortcut];
    [appMenu addItem:[NSMenuItem separatorItem]];
    [appMenu addItemWithTitle:t(@"appQuit") action:@selector(menuQuit) keyEquivalent:@"q"];
    [appItem setSubmenu:appMenu];
    [mainMenu addItem:appItem];

    // File menu
    NSMenuItem *fileItem = [[NSMenuItem alloc] init];
    [fileItem setTitle:t(@"menuFile")];
    NSMenu *fileMenu = [[NSMenu alloc] initWithTitle:t(@"menuFile")];
    [fileMenu addItemWithTitle:t(@"fileOpen") action:@selector(menuOpen) keyEquivalent:@"o"];
    [fileMenu addItemWithTitle:t(@"fileReload") action:@selector(menuReload) keyEquivalent:@"r"];
    [fileMenu addItem:[NSMenuItem separatorItem]];
    
    // Recent files submenu
    NSMenuItem *recentItem = [[NSMenuItem alloc] init];
    recentItem.title = @"最近開啟";
    NSMenu *recentMenu = [[NSMenu alloc] init];
    recentItem.submenu = recentMenu;
    [fileMenu addItem:recentItem];
    
    // Export submenu
    NSMenuItem *exportItem = [[NSMenuItem alloc] init];
    [exportItem setTitle:t(@"menuExport")];
    NSMenu *exportMenu = [[NSMenu alloc] init];
    [exportMenu addItemWithTitle:t(@"fileExportHTML") action:@selector(menuExportHTML) keyEquivalent:@""];
    [exportMenu addItemWithTitle:t(@"fileExportPDF") action:@selector(menuExportPDF) keyEquivalent:@""];
    [exportItem setSubmenu:exportMenu];
    [fileMenu addItem:exportItem];
    [fileItem setSubmenu:fileMenu];
    [mainMenu addItem:fileItem];

    // View menu
    NSMenuItem *viewItem = [[NSMenuItem alloc] init];
    [viewItem setTitle:t(@"menuView")];
    NSMenu *viewMenu = [[NSMenu alloc] initWithTitle:t(@"menuView")];
    [viewMenu addItemWithTitle:t(@"viewIn") action:@selector(menuZoomIn) keyEquivalent:@"="];
    [viewMenu addItemWithTitle:t(@"viewOut") action:@selector(menuZoomOut) keyEquivalent:@"-"];
    [viewMenu addItemWithTitle:t(@"viewReset") action:@selector(menuZoomReset) keyEquivalent:@"0"];
    [viewMenu addItem:[NSMenuItem separatorItem]];
    NSMenuItem *focusItem = [[NSMenuItem alloc] initWithTitle:t(@"viewFocus") action:@selector(menuFocusMode) keyEquivalent:@"m"];
    focusItem.keyEquivalentModifierMask = NSEventModifierFlagCommand | NSEventModifierFlagShift;
    [viewMenu addItem:focusItem];
    [viewMenu addItem:[NSMenuItem separatorItem]];
    [viewMenu addItemWithTitle:t(@"viewFull") action:@selector(menuFullscreen) keyEquivalent:@"f"];
    [viewItem setSubmenu:viewMenu];
    [mainMenu addItem:viewItem];

    // Help menu
    NSMenuItem *helpItem = [[NSMenuItem alloc] init];
    [helpItem setTitle:t(@"menuHelp")];
    NSMenu *helpMenu = [[NSMenu alloc] initWithTitle:t(@"menuHelp")];
    [helpMenu addItemWithTitle:t(@"helpAbout") action:@selector(orderFrontStandardAboutPanel:) keyEquivalent:@""];
    [helpItem setSubmenu:helpMenu];
    [mainMenu addItem:helpItem];

    [NSApp setMainMenu:mainMenu];
}

- (void)menuPreferences { goMenuCallback(2); }
- (void)menuOpen        { goMenuCallback(3); }
- (void)menuReload     { goMenuCallback(4); }
- (void)menuOpenRecent:(NSMenuItem *)sender {
    NSString *path = sender.representedObject;
    if (path) {
        goOpenFileCallback([path UTF8String]);
    }
}
void goRemoveRecentFileCallback(const char *path);

- (void)menuRemoveRecent:(NSMenuItem *)sender {
    NSString *path = sender.representedObject;
    if (path) {
        goRemoveRecentFileCallback([path UTF8String]);
    }
}
- (void)menuQuit        { [NSApp terminate:nil]; }
- (void)menuZoomIn      { goMenuCallback(6); }
- (void)menuZoomOut     { goMenuCallback(7); }
- (void)menuZoomReset   { goMenuCallback(8); }
- (void)menuExportHTML  { goMenuCallback(12); }
- (void)menuExportPDF   { goMenuCallback(13); }
- (void)menuFocusMode  { 
    NSLog(@"menuFocusMode triggered");
    goMenuCallback(14); 
}
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

// Update recent files menu
void UpdateRecentFilesMenu(const char **files, int count) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSMenu *mainMenu = [NSApp mainMenu];
        if (!mainMenu) return;
        
        NSArray *items = mainMenu.itemArray;
        NSMenuItem *fileItem = nil;
        for (NSMenuItem *item in items) {
            if ([item.title isEqualToString:@"檔案"] || [item.title isEqualToString:@"File"]) {
                fileItem = item;
                break;
            }
        }
        if (!fileItem) return;
        NSMenu *fileMenu = fileItem.submenu;
        if (!fileMenu) return;
        
        NSMenuItem *recentItem = nil;
        for (int i = 0; i < fileMenu.numberOfItems; i++) {
            NSMenuItem *item = [fileMenu itemAtIndex:i];
            if ([item.title isEqualToString:@"最近開啟"] || [item.title isEqualToString:@"Open Recent"]) {
                recentItem = item;
                break;
            }
        }
        
        if (!recentItem) return;
        
        NSMenu *recentMenu = recentItem.submenu;
        [recentMenu removeAllItems];
        
        for (int i = 0; i < count; i++) {
            NSString *filePath = [NSString stringWithUTF8String:files[i]];
            if (!filePath) continue;
            
            BOOL fileExists = [[NSFileManager defaultManager] fileExistsAtPath:filePath];
            NSString *fileName = [filePath lastPathComponent];
            
            NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:fileName 
                                                           action:fileExists ? @selector(menuOpenRecent:) : @selector(menuRemoveRecent:) 
                                                    keyEquivalent:@""];
            item.representedObject = [filePath copy];
            
            if (!fileExists) {
                // File doesn't exist - show as gray but still clickable to remove
                NSMutableAttributedString *attrTitle = [[NSMutableAttributedString alloc] initWithString:fileName];
                [attrTitle addAttribute:NSForegroundColorAttributeName 
                                  value:[NSColor disabledControlTextColor] 
                                  range:NSMakeRange(0, fileName.length)];
                item.attributedTitle = attrTitle;
                item.toolTip = @"檔案不存在，點擊移除";
            }
            
            [recentMenu addItem:item];
        }
        
        if (count == 0) {
            NSMenuItem *emptyItem = [[NSMenuItem alloc] initWithTitle:@"無最近檔案" 
                                                               action:nil 
                                                        keyEquivalent:@""];
            emptyItem.enabled = NO;
            [recentMenu addItem:emptyItem];
        }
    });
}

// Set window frame (position + size)
void SetWindowFrame(void *windowPtr, int x, int y, int width, int height) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *window = (__bridge NSWindow *)windowPtr;
        if (window) {
            NSRect frame = NSMakeRect(x, y, width, height);
            [window setFrame:frame display:YES animate:NO];
        }
    });
}

// Get current window size
void GetWindowSize(void *windowPtr, int *width, int *height) {
    NSWindow *window = (__bridge NSWindow *)windowPtr;
    if (window) {
        NSRect frame = [window frame];
        *width = (int)frame.size.width;
        *height = (int)frame.size.height;
    } else {
        *width = 0;
        *height = 0;
    }
}

// Get current window position
void GetWindowPosition(void *windowPtr, int *x, int *y) {
    NSWindow *window = (__bridge NSWindow *)windowPtr;
    if (window) {
        NSRect frame = [window frame];
        *x = (int)frame.origin.x;
        *y = (int)frame.origin.y;
    } else {
        *x = 0;
        *y = 0;
    }
}
