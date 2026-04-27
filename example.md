# md-viewer 完整功能測試

這是一個用於驗證 `swift-markdown` 解析引擎與 `md-viewer` 渲染效果的測試文件。

---

## 1. 文字樣式與段落

這是普通文字。這是 **粗體文字**，這是 *斜體文字*。
您也可以使用 `行內程式碼` 來標註指令。

> 這是一個區塊引用 (Blockquote)。
> 它可以用來引用他人的話語或特別標註內容。
> test
> test
> test
---

## 2. 清單測試

### 無序清單
- 項目一
- 項目二
  - 子項目 A
  - 子項目 B
- 項目三

### 有序清單
1. 第一步
2. 第二步
3. 第三步

---

## 3. 表格測試

| 功能 | 狀態 | 備註 |
| :--- | :--- | :--- |
| Swift 解析 | ✅ 支援 | 透過 swift-markdown |
| 異步渲染 | ✅ 支援 | 解決大檔案卡頓 |
| 多主題 | ✅ 支援 | 包含 Nord, Sepia 等 |
| 10段縮放 | ✅ 支援 | ⌘+ / ⌘- |

---

## 4. 連結與圖片

### 連結行為測試
- **外部網址** (瀏覽器開啟): [點我訪問 Google](https://www.google.com)
- **本地資料夾** (Finder 開啟): [開啟 /Users 目錄](file:///Users)
- **相對路徑** (App 內跳轉): [開啟同目錄的 test.md](test.md)

### 圖片顯示測試
![Markdown Logo](https://upload.wikimedia.org/wikipedia/commons/4/48/Markdown-mark.svg)

![MarkdownUI](https://raw.githubusercontent.com/gonzalezreal/swift-markdown-ui/refs/heads/main/Examples/Demo/Screenshot.png)

---

## 5. 程式碼區塊

```swift
import Foundation
import Markdown

func helloWorld() {
    print("Hello, md-viewer!")
}
```

```go
package main

func main() {
    println("Go + Swift is awesome!")
}
```

```console
ls -l
pwd
```

```bash
#!/bin/bash

echo "hello, world~"
```

---

## 6. 特殊字元轉義測試
測試 HTML 轉義：`<script>alert('xss')</script>` & `&` 符號。

---
*文件生成時間：2026-04-26*
