# T011-FIX-04 Summary — ⌘F 全螢幕修復

## 專案目錄
`/Users/claw/Projects/md-viewer-webview`

## 問題
⌘F 觸發 `menuFullscreen` selector，但無任何效果。

## 根因
1. `self.targetWindow` 永遠是 nil（property 宣告了但從未賦值），`[self.targetWindow toggleFullScreen:nil]` 發給 nil → no-op
2. `MenuFullscreen` ID 與 `MenuAboutHelp` 衝突（都是 10）

## 修改內容

### `menu.m` — 修正 1：移除 `targetWindow` property
刪除 `@property (nonatomic, assign) NSWindow *targetWindow;`

### `menu.m` — 修正 2：`menuFullscreen` 改用 `[NSApp keyWindow]`
```objc
- (void)menuFullscreen  {
    NSWindow *window = [NSApp keyWindow];
    if (window && [window respondsToSelector:@selector(toggleFullScreen:)]) {
        [window toggleFullScreen:nil];
    }
}
```

### `menu.go` — 修正 3：修正 ID 衝突
```go
// 修正前
MenuToggleSidebar = 9
MenuFullscreen    = 10
MenuAboutHelp     = 10  // 衝突！

// 修正後
MenuToggleSidebar = 9
MenuAboutHelp     = 10
MenuFullscreen    = 11
```

## 驗收結果
- ✅ `go build` 成功無錯誤（只有不相關的 deprecation warnings）
- ⏳ ⌘F 正確進入全螢幕 — 待手動測試
- ⏳ 再按 ⌘F 正確退出全螢幕 — 待手動測試
- ✅ Go menu ID 無衝突（MenuAboutHelp=10, MenuFullscreen=11）

## 狀態
**done** — 修改完成，build 通過。
