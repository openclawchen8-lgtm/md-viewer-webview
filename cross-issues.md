✦ 簡單來說：這個專案目前無法直接編譯成 Windows 或 Linux 版本。

  原因在於我們為了達成 macOS 的深度整合，使用了大量macOS 專用技術，這些技術在 Windows/Linux 上是完全無法運作的：

  1. 核心技術依賴 (OS-Specific Dependencies)
   * Objective-C (menu.m, dragdrop.m)：這部分代碼是 macOS 的 Cocoa API，它直接控制 Mac 的選單欄、視窗與拖放邏輯。這些檔案在 Windows
     上是無法編譯的，也沒有對應的替代品。
   * Swift (MarkdownEngine)：雖然 Swift 可以跨平台，但 Apple 的 swift-markdown 庫高度依賴 Foundation 和 Apple 的系統函式庫，要移植到 Windows/Linux
     非常困難且不穩定。
   * CGO 橋接：我們大量使用 LDFLAGS 連結 AppKit 與 WebKit (macOS 框架)，這些函式庫在 Windows/Linux 上不存在。

  2. 如果您希望跨平台，該怎麼做？
  如果您未來有跨平台的需求，架構需要做較大的調整：

   * 前端層 (UI)：目前的 webview 庫 (webview_go) 本身是跨平台的，這部分可以保留。
   * 解析層 (Markdown Engine)：您需要放棄 swift-markdown，改回使用純 Go 語言編寫的解析器（如 goldmark），因為 Go 是跨平台的。
   * 原生層 (Menu & System)：這部分必須完全重寫。
       * macOS：保持現狀。
       * Windows/Linux：需要使用該作業系統對應的選單與視窗 API（例如 Windows 上的 Win32 API 或使用跨平台的 GUI 框架如 Wails、Fyne 或 Electron）。

  結論
  現在的 md-viewer 是一個「純血統的 macOS 應用」，這也是為什麼它能擁有這麼好的原生體驗（例如雙擊開檔、選單列整合、極快的 Swift 解析速度）。

  如果您有跨平台的需求，最好的方式是：
  保持這個 md-viewer 作為您的 macOS 原生產品，如果未來要擴展到其他平台，我建議開發一個獨立的分支，使用 Wails (Go + HTML/CSS) 框架，這樣可以最大程度地共用
  Go 的業務邏輯代碼，同時實現真正的跨平台原生介面。

分析一下如何達成：

  1. 什麼功能可以 100% 複製？
   * Markdown 解析邏輯：使用 Go 的 goldmark (甚至是重寫一個)，可以完全無痛跨平台。
   * Web 渲染邏輯：HTML/CSS/JS 的代碼幾乎不需要改動，因為所有的瀏覽器核心（WebKit/Blink）在各平台上對 Web 標準的支援度已經趨於一致。
   * 設定管理：Go 的 encoding/json 處理與檔案讀寫是完全跨平台的。
   * 國際化 (i18n)：JSON 翻譯檔案與 JS 處理邏輯可以完全共用。

  2. 什麼功能需要改寫？（這是跨平台最大的成本）
   * 選單欄 (Menu Bar)：
       * macOS 的 NSMenu 是系統級的。
       * Windows 需要 Win32 API (HMENU)，Linux (GNOME) 則是透過 D-Bus 或 GTK 實作。
       * 解法：使用像 Wails 或 Fyne 這類框架，它們會幫你封裝好各平台的原生選單。
   * 檔案關聯 (File Association)：
       * Windows 使用 Registry (登錄檔)。
       * macOS 使用 Info.plist。
       * Linux 使用 .desktop 檔案。
       * 解法：安裝程式 (Installer) 或打包工具（如 wails build）會處理這些平台差異。
   * 原生視窗操作：例如「全螢幕」或「開檔選擇器」，必須由跨平台 GUI 庫提供，不能再寫 Objective-C。

  ---

  與您現在的 md-viewer 相比，Wails 的優劣：

  ┌──────────┬──────────────────────────────────┬─────────────────────────────────┐
  │ 特性     │ 目前的 md-viewer (CGO + WebView) │ Wails (Go + Web)                │
  ├──────────┼──────────────────────────────────┼─────────────────────────────────┤
  │ 視覺效果 │ 極致 (全權由 CSS 控制)           │ 極致 (全權由 CSS 控制)          │
  │ 整合程度 │ 非常高 (原生 Swift/OC)           │ 高 (透過 Go 橋接)               │
  │ 跨平台性 │ 無 (完全綁定 macOS)              │ 高 (一套代碼支援 Mac/Win/Linux) │
  │ 維護成本 │ 高 (需要處理 CGO/Swift 橋接)     │ 低 (標準 Go 開發流程)           │
  └──────────┴──────────────────────────────────┴─────────────────────────────────┘
  ---

