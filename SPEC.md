github_issue: https://github.com/openclawchen8-lgtm/openclaw-tasks/issues/91

# md-viewer-app 規格文件

**版本：** v0.2.0  
**狀態：** 草稿  
**建立：** 2026-04-23  
**更新：** 2026-04-23（新增 zserge/webview + CSS 渲染方案）

---

## 1. 產品概述

- **名稱：** md-viewer
- **類型：** macOS 原生桌面應用
- **定位：** 純閱讀導向的 Markdown 預覽器（零編輯功能）
- **目標用戶：** 需要快速預覽 .md 檔案、不需要編輯功能的使用者

---

## 2. 功能需求

### MVP（MVP）

| 功能 | 說明 |
|------|------|
| **開啟檔案** | ⌘O 開啟 .md 檔，支援拖放 |
| **Markdown 渲染** | 完整 GFM（表格、任務清單、刪除線、程式碼區塊、連結） |
| **側邊欄檔案列表** | 開啟資料夾時自動掃描所有 .md |
| **深色模式** | 自動跟隨 macOS 系統主題 |
| **⌘W 關閉** | 關閉 App |

### 未来功能（P1-P3）

| 功能 | 優先級 |
|------|--------|
| 雙擊 .md 直接開啟 | P1 |
| 書籤 | P2 |
| 關鍵字搜尋 | P2 |
| Copy HTML | P3 |

---

## 3. UI 規格

### 3.1 視窗佈局

```
┌────────────────────────────────────────────────────────────┐
│  📄 md-viewer                              [─] [□] [×]    │
├──────────────┬─────────────────────────────────────────────┤
│  📁 側邊欄   │         Markdown 預覽區（WebView）         │
│              │                                             │
│  ▸ folder1/ │  # 標題                                      │
│    ▸ notes  │  內文內容...                                │
│      readme │  ```go                                       │
│      guide  │  code block                                  │
│    ▸ docs   │  ```                                        │
│              │                                             │
├──────────────┴─────────────────────────────────────────────┤
│  📄: readme.md                              🌙 深色模式      │
└────────────────────────────────────────────────────────────┘
```

### 3.2 預覽區技術規格

- **渲染引擎：** zserge/webview（macOS WebKit 底層）
- **最大寬度：** 800px，內容居中
- **樣式：** GitHub 官方 Markdown 樣式（含深色模式）
- **字體：** `-apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif`（內文）/ monospace（程式碼）
- **滾動：** 獨立滾動條，支援平滑滾動

### 3.3 選單列

| 選單 | 項目 |
|------|------|
| **檔案** | 開啟檔案... (⌘O)、開啟資料夾... (⌘⇧O)、關閉 (⌘W) |
| **顯示** | 切換側邊欄 (⌘B) |
| **說明** | 關於 md-viewer |

---

## 4. 技術規格

### 4.1 技術棧（v0.2）

| 層面 | 技術選擇 | 理由 |
|------|----------|------|
| **語言** | Go 1.21+ | 主要開發語言 |
| **WebView** | zserge/webview | 輕量跨平台，直接控制 HTML/CSS |
| **Markdown** | goldmark + GFM extensions | Go 生態品質解析器 |
| **樣式** | 自訂 CSS（含深色模式） | 完整控制渲染外觀 |
| **打包** | `go build` + 手動 .app | 避免 Fyne UI 依賴 |

### 4.2 依賴

```go
require (
    github.com/zserge/webview v0.0.0-20260309075125-cbbdee44afff
    github.com/yuin/goldmark v1.8.2
)
```

### 4.3 資料夾結構

```
md-viewer-app/
├── main.go              ← 入口
├── go.mod / go.sum
├── SPEC.md
├── AGENTS.md
├── assets/
│   ├── template.html    ← HTML 殼（含 dark mode 支援）
│   └── markdown.css    ← GitHub 風格（含深色模式）
├── ui/
│   ├── app.go          ← 主程式邏輯
│   ├── sidebar.go      ← 側邊欄
│   └── webview.go      ← zserge/webview 包裝
└── core/
    ├── markdown.go      ← goldmark → HTML
    └── file_tree.go     ← 資料夾掃描
```

---

## 5. 驗收標準

### MVP 驗收

- [ ] ⌘O 開啟 .md 檔，正確渲染 GFM
- [ ] 拖放 .md 到視窗可開啟預覽
- [ ] 側邊欄掃描並列出目錄中的 .md
- [ ] 深色模式下樣式正確切換
- [ ] ⌘W 關閉 App
- [ ] .app bundle < 30MB

### 視覺標準（★★★★★）

- [ ] 程式碼區塊：monospace 字體 + 深色背景 + 邊框
- [ ] 表格：完整边框 + 斑馬紋
- [ ] 區塊引用：左側彩色邊線
- [ ] 任務清單：checkbox 樣式
- [ ] 圖片：自適應最大寬度
- [ ] 深色模式：所有元素正確反轉

---

*規格作者：寶寶 | 最後更新：2026-04-23*
