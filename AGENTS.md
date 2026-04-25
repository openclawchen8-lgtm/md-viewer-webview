# AGENTS.md — md-viewer-app 團隊協作規範

## 專案目錄
`/Users/claw/Projects/md-viewer-app`

## 團隊分工

| 成員 | 角色 | 負責任務 |
|------|------|---------|
| 寶寶 | Planner | 規劃、統籌 |
| 碼農2號 | Coder | 實作 T003 系列 |

## 實作規範

### 1. zserge/webview 使用原則
- `webview.WebView` 為唯讀展示，不做輸入框
- HTML 內容由 goldmark 生成，注入 `template.html` 殼
- 深色模式透過 CSS 變數切換，不重建 WebView

### 2. goldmark 配置
- 啟用 GFM extensions：表格、任務清單、刪除線、連結自動轉換、程式碼區塊
- AutoHeadingID：每個標題有穩定錨點（供未來書籤功能用）

### 3. CSS 樣式優先順序
1. `assets/markdown.css`（主要）
2. `@media (prefers-color-scheme: dark)`（深色模式覆蓋）
3. `body[data-theme="dark"]`（JS 動態切換覆蓋）

### 4. macOS 深色模式偵測
```go
// 方法1：讀取系統偏好（推薦）
cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
out, _ := cmd.Output()
dark := strings.Contains(string(out), "Dark")

// 方法2：CSS prefers-color-scheme（由 WebView 自動處理）
// 無需 Go 端偵測，HTML CSS 內建自動切換
```

### 5. 構建命令
```bash
# 開發构建
go build -o md-viewer .

# macOS .app（用 Fyne CLI 打包）
go build -o md-viewer .
fyne package --executable ./md-viewer --name md-viewer --os darwin --app-id com.mdviewer.app
```

### 6. Sub-Agent 規範
- Spawn 時 `cwd` 必填：`/Users/claw/Projects/md-viewer-app`
- T003 各子任務改不同檔案，可並行實作
- `main.go` / `ui/app.go` 只由一個 sub-agent 負責（避免衝突）

---

*建立時間：2026-04-23 by 寶寶*
