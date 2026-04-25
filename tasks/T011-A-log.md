# T011-A 實作日誌

## 2026-04-25 — Native NSMenu 骨架實作

### 目標
建立 Native macOS NSMenu CGO bridge。

### 實作決策
- menu.go：CGO bridge，var menuCallback func(int)，SetupMainMenu()
- menu.m：ObjC 實作，NSApp.mainMenu，goMenuCallback(int)
- Callback ID 用 uintptr index（與 webview binding 機制一致）
- MenuItem action → ObjC selector → goMenuCallback(id)
