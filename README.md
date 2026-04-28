# md-viewer

**一款為 macOS 打造的高效能 Markdown 閱讀器。**

[![Platform](https://img.shields.io/badge/platform-macOS%2012+-lightgrey?style=flat-square)](https://www.apple.com/macos)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Swift](https://img.shields.io/badge/Swift-5.7+-FA7343?style=flat-square&logo=swift)](https://swift.org)

---

`md-viewer` 以 **WebKit** 為渲染核心，結合 **Apple swift-markdown** 解析引擎，為 Markdown 文件提供精準、快速的閱讀體驗。介面簡潔，支援豐富的主題與快捷鍵，是 macOS 上日常閱讀 `.md` 檔案的理想工具。

---

## ✨ 功能特色

### 精準的 Markdown 渲染
- **Swift-markdown 解析引擎**：由 Apple 官方 `swift-markdown` 程式庫提供支援，GFM（GitHub Flavored Markdown）語法完整支援，包括標題、段落、表格、任務清單、刪除線、程式碼區塊、Blockquote 等。
- **GFM 表格**：支援 Markdown 表格，含邊框與交替列底色。
- **任務清單**：`- [ ]` / `- [x]` 語法直接渲染為可見的核取方塊。

### 程式碼體驗
- **語法高亮**：整合 highlight.js，支援 190+ 種語言的即時著色。
- **一鍵複製**：滑鼠移過程式碼區塊，右上角浮現複製按鈕，點擊即可複製內容。
- **行號顯示**：可在設定面板開關程式碼行號。

### 多元主題
- **6 種精心設計的主題**：
  | 主題 | 說明 |
  |------|------|
  | Auto | 自動跟隨 macOS 系統外觀（Light / Dark） |
  | GitHub Light | GitHub 官方淺色文件風格 |
  | GitHub Dark | GitHub 官方深色文件風格 |
  | Sepia | 羊皮紙護眼暖色調 |
  | Solarized | Solarized 極簡配色 |
  | Nord | 北歐極地冷色調 |
- **主題即時切換**：無需重啟，變更立即套用。
- **語法高亮同步**：淺色主題自動使用淺色高亮配色，深色主題則切換為深色配色。

### 設定面板
- **縮放控制**：支援 `⌘+` / `⌘-` 縮放（50%–200%），縮放靈敏度可自訂。
- **字型與大小**：可選擇字體與尺寸。
- **i18n**：完整支援繁體中文、簡體中文、English、日本語、한국어，選單文字同步切換。
- **所有設定自動儲存**，下次開啟自動還原。

### 檔案互動
- **拖放開啟**：直接將 `.md` 檔案拖入視窗即可開啟，視覺化放置區域提示。
- **檔案監控自動重載**：外部編輯器存檔後，視窗自動重新渲染，並帶有粉紅色閃爍邊框動畫提示。
- **相對路徑智慧解析**：本地圖片或連結可自動解析為正確路徑。
- **外部連結**：文件中的 URL 點擊後自動以預設瀏覽器開啟。

### macOS 原生整合
- **原生選單列**：標準 macOS NSMenu，含應用程式、檔案、檢視、說明選單。
- **完整快捷鍵**：⌘O 開檔、⌘R 重載、⌘+/- 縮放、⌘0 重置、⌘, 設定、⌘F 全螢幕、⌘Q 結束。
- **文件關聯**：支援 `.md`、`.markdown`、`.txt` 檔案關聯，Finder 雙擊即可用 md-viewer 開啟。
- **匯出功能**：支援匯出為獨立 HTML（含完整內嵌樣式）與 PDF（原生列印流程）。

---

## ⌨️ 快捷鍵

| 功能 | 快捷鍵 |
|:---|:---|
| 開啟檔案 | `⌘O` |
| 重新載入 | `⌘R` |
| 放大 | `⌘+` 或 `⌘=` |
| 縮小 | `⌘-` |
| 重置縮放 | `⌘0` 或 `⇧⌘R` |
| 設定面板 | `⌘,` |
| 全螢幕 | `⌘F` |
| 匯出 HTML | `⌘⇧E` |
| 匯出 PDF | `⌘⇧P` |
| 結束程式 | `⌘Q` |

> **Trackpad 手勢**：`Ctrl` + 雙指滾動同樣支援縮放。

---

## 📦 安裝

### 方式一：直接使用（推薦）

```bash
# 複製 App bundle
cp -r ./md-viewer.app /Applications/

# 設定為 .md 檔案的預設開啟程式
# 對任意 .md 檔案按右鍵 → 開啟方式 → 選擇 md-viewer → 點選「全部替換」
```

### 方式二：命令列執行

```bash
./md-viewer.app/Contents/MacOS/md-viewer           # 開啟空視窗
./md-viewer.app/Contents/MacOS/md-viewer readme.md  # 直接開啟指定檔案
```

---

## 🔨 從原始碼編譯

### 環境需求

| 工具 | 最低版本 |
|------|----------|
| macOS | 12.0+ (Monterey) |
| Go | 1.26+ |
| Swift | 5.7+ |
| Xcode Command Line Tools | 已安裝 |

### 編譯步驟

```bash
git clone https://github.com/openclawchen8-lgtm/openclaw-tasks.git
cd openclaw-tasks/md-viewer-webview

# 一次性編譯 + 打包
./build.sh
```

`build.sh` 會依序執行：
1. `go mod tidy` — 整理 Go 依賴
2. `swift build -c release` — 編譯 Swift MarkdownEngine → `libMarkdownEngine.dylib`
3. `go build -o md-viewer` — 編譯 Go 主程式（含 CGO）
4. 建立 `md-viewer.app/` bundle 結構
5. 複製 Swift dylib → `Contents/Frameworks/`
6. 修正 `@rpath` 以確保動態連結正確

---

## 🏗 技術架構

```
┌─────────────────────────────────────────────┐
│                 macOS App                    │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │            webview_go (WKWebView)       │ │
│  │   HTML Template + CSS + JavaScript      │ │
│  │   • 主題切換  • 縮放  • i18n           │ │
│  │   • 程式碼高亮  • 複製按鈕             │ │
│  └──────────────┬───────────────────────────┘ │
│                 │ JS Bindings                │
│  ┌──────────────▼───────────────────────────┐ │
│  │         main.go (Go)                     │ │
│  │   • 設定持久化  • NSMenu 回调           │ │
│  │   • 檔案監控  • 匯出邏輯                │ │
│  └──────────────┬───────────────────────────┘ │
│                 │ CGO FFI                     │
│  ┌──────────────▼───────────────────────────┐ │
│  │     core/renderer.go → libMarkdownEngine │ │
│  │            (C shared library)             │ │
│  └──────────────┬───────────────────────────┘ │
│                 │ Swift C Export               │
│  ┌──────────────▼───────────────────────────┐ │
│  │  Sources/MarkdownEngine/Engine.swift     │ │
│  │       swift-markdown (AST → HTML)        │ │
│  └──────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

### 核心技術棧

| 層級 | 技術 |
|------|------|
| 渲染引擎 | WebKit（WKWebView via webview_go） |
| Markdown 解析 | swift-markdown（Apple 官方） |
| UI 樣式 | GitHub 風格 CSS 變數 + highlight.js |
| 原生功能 | CGO + Objective-C（NSMenu、Drag&Drop、PDF Export） |
| 設定儲存 | JSON 檔案（`~/.md-viewer/config.json`） |
| 檔案監控 | Go fsnotify |

---

## 🔌 設定檔

設定儲存於 `~/.md-viewer/config.json`，手動編輯亦可：

```json
{
  "zoomSensitivity": 5,
  "theme": "auto",
  "zoomLevel": 1.0,
  "fontFamily": "-apple-system, BlinkMacSystemFont, ...",
  "fontSize": 16,
  "language": "zhTW",
  "showLineNumbers": false
}
```

---

## ❓ 常見問題

**Q: 為什麼使用 Swift-markdown 而非純 Go 方案？**
Apple 的 `swift-markdown` 解析品質極高，GFM 語法支援完整，與 macOS 原生整合緊密，適合作為 macOS 專屬閱讀器的核心。

**Q: 可以跨平台嗎？**
目前僅支援 macOS。若需 Linux/Windows 版本，可將 Swift-markdown 替換為 goldmark（純 Go），但渲染品質會有所差異。

**Q: .md 檔案關聯失效怎麼辦？**
執行 `open -a md-viewer.app` 一次，macOS 會重新註冊文件類型。

---

## 📄 授權

本專案為個人學習與使用目的建立。
