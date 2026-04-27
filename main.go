package main

import (
	"C"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"md-viewer/core"

	"github.com/fsnotify/fsnotify"
	"github.com/webview/webview_go"
)

var (
	watcher      *fsnotify.Watcher
	watchMu      sync.Mutex
	debounceTimer *time.Timer
	pendingFlash bool
)

const cssContent = `
:root, [data-theme='light'] {
  --color-canvas-default: #ffffff;
  --color-canvas-subtle: #f6f8fa;
  --color-fg-default: #1f2328;
  --color-fg-muted: #656d76;
  --color-border-default: #d0d7de;
  --color-accent-fg: #0969da;
  --color-accent-emphasis: #0550ae;
  --color-danger-fg: #d1242f;
  --color-success-fg: #1a7f37;
}
[data-theme='dark'] {
  --color-canvas-default: #0d1117;
  --color-canvas-subtle: #161b22;
  --color-fg-default: #e6edf3;
  --color-fg-muted: #8b949e;
  --color-border-default: #30363d;
  --color-accent-fg: #58a6ff;
  --color-accent-emphasis: #388bfd;
  --color-danger-fg: #f85149;
  --color-success-fg: #3fb950;
}
[data-theme='sepia'] {
  --color-canvas-default: #f4ecd8;
  --color-canvas-subtle: #eaddc0;
  --color-fg-default: #5b4636;
  --color-fg-muted: #7d6b5d;
  --color-border-default: #d3c4a9;
  --color-accent-fg: #a65d00;
  --color-accent-emphasis: #874b00;
  --color-danger-fg: #d1242f;
  --color-success-fg: #1a7f37;
}
[data-theme='solarized'] {
  --color-canvas-default: #002b36;
  --color-canvas-subtle: #073642;
  --color-fg-default: #839496;
  --color-fg-muted: #586e75;
  --color-border-default: #073642;
  --color-accent-fg: #268bd2;
  --color-accent-emphasis: #2aa198;
  --color-danger-fg: #dc322f;
  --color-success-fg: #859900;
}
[data-theme='nord'] {
  --color-canvas-default: #2e3440;
  --color-canvas-subtle: #3b4252;
  --color-fg-default: #d8dee9;
  --color-fg-muted: #949fb1;
  --color-border-default: #434c5e;
  --color-accent-fg: #88c0d0;
  --color-accent-emphasis: #81a1c1;
  --color-danger-fg: #bf616a;
  --color-success-fg: #a3be8c;
}
@media (prefers-color-scheme: dark) {
  [data-theme='auto'] {
    --color-canvas-default: #0d1117;
    --color-canvas-subtle: #161b22;
    --color-fg-default: #e6edf3;
    --color-fg-muted: #8b949e;
    --color-border-default: #30363d;
    --color-accent-fg: #58a6ff;
    --color-accent-emphasis: #388bfd;
    --color-danger-fg: #f85149;
    --color-success-fg: #3fb950;
  }
}
* { box-sizing: border-box; margin: 0; padding: 0; }
html, body { height: 100%; overflow: auto; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
  font-size: 16px; line-height: 1.6;
  color: var(--color-fg-default);
  background: var(--color-canvas-default);
  padding: 2rem;
}
@keyframes pink-flash {
  0%, 100% { border: 0px solid rgba(255, 20, 147, 0); box-shadow: inset 0 0 0 0 rgba(255, 20, 147, 0); }
  20%, 40% { border: 10px solid rgba(255, 20, 147, 0.8); box-shadow: inset 0 0 50px 20px rgba(255, 20, 147, 0.7); }
  50% { border: 0px solid rgba(255, 20, 147, 0); box-shadow: inset 0 0 0 0 rgba(255, 20, 147, 0); }
  70%, 90% { border: 10px solid rgba(255, 20, 147, 0.8); box-shadow: inset 0 0 50px 20px rgba(255, 20, 147, 0.7); }
}
.reload-flash #reloadOverlay {
  animation: pink-flash 1.5s ease-in-out;
}
#reloadOverlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  pointer-events: none; z-index: 2147483647;
  border: 0px solid transparent;
  box-sizing: border-box;
}
.markdown-body { max-width: 900px; margin: 0 auto; }
.markdown-body h1 { font-size: 2em; border-bottom: 1px solid var(--color-border-default); padding-bottom: 0.3em; margin-bottom: 1em; margin-top: 1.5em; }
.markdown-body h2 { font-size: 1.5em; border-bottom: 1px solid var(--color-border-default); padding-bottom: 0.3em; margin-bottom: 0.8em; margin-top: 1.5em; }
.markdown-body h3 { font-size: 1.25em; margin-bottom: 0.6em; margin-top: 1.2em; }
.markdown-body h4 { font-size: 1em; margin-bottom: 0.5em; margin-top: 1em; }
.markdown-body h1:first-child, .markdown-body h2:first-child, .markdown-body h3:first-child { margin-top: 0; }
.markdown-body p { margin-bottom: 1em; }
.markdown-body a { color: var(--color-accent-fg); text-decoration: none; }
.markdown-body a:hover { text-decoration: underline; }
@media print {
  .keyboard-hint, .drop-zone, .settings-overlay, #zoomIndicator { display: none !important; }
}
.is-exporting .keyboard-hint, .is-exporting .drop-zone, .is-exporting .settings-overlay, .is-exporting #zoomIndicator {
  display: none !important;
}
@media print, .is-exporting {
  html, body { height: auto !important; overflow: visible !important; min-height: 0 !important; }
  .markdown-body { height: auto !important; overflow: visible !important; max-width: 100% !important; margin: 0 !important; padding: 2cm !important; }
  * { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
}
.markdown-body code {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 85%;
  background: var(--color-canvas-subtle);
  border: 1px solid var(--color-border-default);
  border-radius: 6px;
  padding: 0.2em 0.4em;
}
.markdown-body pre {
  background: var(--color-canvas-subtle);
  border: 1px solid var(--color-border-default);
  border-radius: 6px;
  padding: 1rem;
  overflow-x: auto;
  margin-bottom: 1.2em;
  position: relative; /* 為複製按鈕定位 */
}
.code-copy-btn {
  position: absolute; top: 8px; right: 8px;
  padding: 4px 8px; font-size: 11px;
  background: var(--color-canvas-default);
  color: var(--color-fg-muted);
  border: 1px solid var(--color-border-default);
  border-radius: 4px; cursor: pointer; opacity: 0;
  transition: opacity 0.2s, background 0.2s;
  z-index: 10;
}
.markdown-body pre:hover .code-copy-btn { opacity: 1; }
.code-copy-btn:hover { background: var(--color-accent-fg); color: white; border-color: var(--color-accent-fg); }
.code-copy-btn.copied { background: var(--color-success-fg); color: white; border-color: var(--color-success-fg); }

/* 行號樣式 */
.line-numbers-mode code {
  display: grid;
  grid-template-columns: minmax(30px, auto) 1fr;
  column-gap: 15px;
}
.line-number {
  color: var(--color-fg-muted);
  text-align: right;
  user-select: none;
  border-right: 1px solid var(--color-border-default);
  padding-right: 10px;
  opacity: 0.5;
}
.markdown-body pre code { background: none; border: none; padding: 0; font-size: 13px; line-height: 1.6; }
.markdown-body blockquote { border-left: 4px solid var(--color-border-default); padding-left: 1rem; color: var(--color-fg-muted); margin-bottom: 1em; }
.markdown-body ul, .markdown-body ol { padding-left: 2rem; margin-bottom: 1em; }
.markdown-body li { margin-bottom: 0.25em; }
.markdown-body table { border-collapse: collapse; width: 100%; margin-bottom: 1.2em; }
.markdown-body table th, .markdown-body table td { border: 1px solid var(--color-border-default); padding: 0.5rem 0.75rem; text-align: left; }
.markdown-body table th { background: var(--color-canvas-subtle); font-weight: 600; }
.markdown-body table tr:nth-child(even) { background: var(--color-canvas-subtle); }
.markdown-body hr { border: none; border-top: 1px solid var(--color-border-default); margin: 2em 0; }
.markdown-body img { max-width: 100%; border-radius: 6px; }
.markdown-body input[type="checkbox"] { margin-right: 0.5em; accent-color: var(--color-accent-emphasis); }
.markdown-body del { color: var(--color-fg-muted); }
.empty-state { text-align: center; padding: 4rem 2rem; color: var(--color-fg-muted); }
.empty-state-icon { font-size: 4rem; margin-bottom: 1rem; }
.empty-state-title { font-size: 1.5rem; font-weight: 600; color: var(--color-fg-default); margin-bottom: 0.5rem; }
.error { color: var(--color-danger-fg); }
.keyboard-hint {
  position: fixed; bottom: 10px; right: 10px;
  font-size: 11px; color: var(--color-fg-muted);
  background: var(--color-canvas-subtle);
  border: 1px solid var(--color-border-default);
  padding: 4px 8px; border-radius: 4px; opacity: 0.9; z-index: 99999;
  pointer-events: none;
}
@media print, .is-exporting {
  .keyboard-hint, .drop-zone, .settings-overlay, #zoomIndicator { display: none !important; }
}
/* Drop Zone Styles */
.drop-zone {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(9, 105, 218, 0.15);
  border: 4px dashed var(--color-accent-fg);
  display: none; align-items: center; justify-content: center;
  z-index: 99999; pointer-events: none;
}
.drop-zone.active { display: flex; }
.drop-zone-msg {
  font-size: 1.5rem; font-weight: 600;
  color: var(--color-accent-fg);
  background: var(--color-canvas-default);
  padding: 1rem 2rem; border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}
.drop-zone.error { background: rgba(209, 36, 47, 0.15); border-color: var(--color-danger-fg); }
.drop-zone.error .drop-zone-msg { color: var(--color-danger-fg); }
/* Settings Panel */
.settings-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 99999;
}
.settings-panel {
  background: var(--color-canvas-default);
  border: 1px solid var(--color-border-default);
  border-radius: 12px;
  padding: 2rem;
  width: 380px; max-width: 90vw;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
}
.settings-title {
  font-size: 1.2rem; font-weight: 600; margin-bottom: 1.5rem;
  color: var(--color-fg-default);
  display: flex; justify-content: space-between; align-items: center;
}
.settings-close {
  background: none; border: none; font-size: 1.2rem; cursor: pointer;
  color: var(--color-fg-muted); padding: 0 4px;
}
.settings-close:hover { color: var(--color-fg-default); }
.setting-row { margin-bottom: 1.2rem; }
.setting-label { font-size: 0.9rem; font-weight: 500; color: var(--color-fg-default); margin-bottom: 0.4rem; }
.setting-desc { font-size: 0.75rem; color: var(--color-fg-muted); margin-bottom: 0.5rem; }
input[type="range"] { width: 100%; accent-color: var(--color-accent-fg); }
.slider-labels { display: flex; justify-content: space-between; font-size: 0.7rem; color: var(--color-fg-muted); margin-top: 0.3rem; }
`

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=5.0, user-scalable=yes">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css" id="hljsLight">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css" id="hljsDark" disabled>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
<style>%s</style>
</head>
<body data-theme="auto">
<div id="reloadOverlay"></div>
<div class="drop-zone" id="dropZone"><div class="drop-zone-msg">Drop .md file here</div></div>
<div class="markdown-body" id="mdContent">%s</div>
<div class="keyboard-hint" id="keyboardHint"><span id="zoomText" style="font-weight:700; color:var(--color-fg-default); margin-right:4px;">100%%</span> | ⌘O Open | ⌘R Reload | ⌘+/- Zoom | ⇧⌘R/⌘0 Reset | ⌘, Settings | ⌘Q Quit</div>
<div class="settings-overlay" id="settingsOverlay" style="display:none">
  <div class="settings-panel">
    <div class="settings-title">
      設定
      <button class="settings-close" id="settingsClose" title="關閉 (Esc)">×</button>
    </div>
    <div class="setting-row">
      <div class="setting-label">佈景主題</div>
      <select id="themeSelect" style="width:100%;padding:6px;border-radius:6px;border:1px solid var(--color-border-default);background:var(--color-canvas-default);color:var(--color-fg-default);">
        <option value="auto">跟隨系統 (Auto)</option>
        <option value="light">GitHub 淺色</option>
        <option value="dark">GitHub 深色</option>
        <option value="sepia">羊皮紙 (Sepia)</option>
        <option value="solarized">Solarized 深色</option>
        <option value="nord">Nord 極地</option>
      </select>
    </div>
    <div class="setting-row">
      <div class="setting-label">縮放靈敏度</div>
      <div class="setting-desc">控制 ⌘+ 和 ⌘- 每次調整的幅度</div>
      <input type="range" id="zoomSensitivity" min="1" max="10" value="5" step="1">
      <div class="slider-labels"><span>低</span><span>中</span><span>高</span></div>
    </div>
    <div class="setting-row">
      <div class="setting-label">字型</div>
      <select id="fontFamily" style="width:100%;padding:6px;border-radius:6px;border:1px solid var(--color-border-default);background:var(--color-canvas-default);color:var(--color-fg-default);">
        <option value="-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif">系統預設</option>
        <option value="'PingFang TC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif">蘋方/微軟雅黑</option>
        <option value="'Noto Sans TC', 'WenQuanYi Micro Hei', sans-serif">思源/文泉驛</option>
        <option value="Georgia, 'Times New Roman', serif">襯線字型</option>
        <option value="'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace">等寬字型</option>
      </select>
    </div>
    <div class="setting-row">
      <div class="setting-label">字型大小</div>
      <select id="fontSize" style="width:100%;padding:6px;border-radius:6px;border:1px solid var(--color-border-default);background:var(--color-canvas-default);color:var(--color-fg-default);">
        <option value="14">14px（小）</option>
        <option value="16">16px（中等）</option>
        <option value="18">18px（較大）</option>
        <option value="20">20px（大）</option>
      </select>
    </div>
    <div class="setting-row">
      <div class="setting-label">語言 / Language</div>
      <select id="language" style="width:100%;padding:6px;border-radius:6px;border:1px solid var(--color-border-default);background:var(--color-canvas-default);color:var(--color-fg-default);">
        <option value="zhTW">繁體中文</option>
        <option value="zhCN">簡體中文</option>
        <option value="en">English</option>
        <option value="ja">日本語</option>
        <option value="ko">한국어</option>
      </select>
    </div>
    <div class="setting-row" style="display:flex; align-items:center; justify-content:space-between;">
      <div class="setting-label" style="margin-bottom:0">顯示行號</div>
      <input type="checkbox" id="showLineNumbers" style="width:auto; height:20px; width:20px; cursor:pointer;">
    </div>
  </div>
</div>
<script>
// Global zoom functions (available immediately for Eval calls)
window.zoomState = { level: 1.0, step: 0.1 };
window._zoomTimer = null;
window.applyZoomLevel = function(level) {
  var minZoom = 0.5, maxZoom = 2.0;
  level = Math.max(minZoom, Math.min(maxZoom, level));
  window.zoomState.level = level;
  if (document.body) document.body.style.zoom = level;
  
  // 強制更新所有顯示縮放文字的地方
  var zt = document.getElementById('zoomText');
  if (zt) {
    zt.innerText = Math.round(level * 100) + '%%';
  }
  
  if (window.saveZoomLevel) window.saveZoomLevel(level);
  window.showZoomIndicator(Math.round(level * 100));
};
window.showZoomIndicator = function(pct) {
  // Floating indicator (fades out after 1.5s)
  var el = document.getElementById('zoomIndicator');
  if (!el) {
    el = document.createElement('div');
    el.id = 'zoomIndicator';
    el.style.cssText = 'position:fixed;top:10px;right:10px;background:rgba(0,0,0,0.8);color:white;padding:5px 12px;border-radius:6px;font-size:13px;font-weight:bold;z-index:99999;pointer-events:none;transition:opacity 0.3s;';
    if (document.body) document.body.appendChild(el);
  }
  if (el) {
    el.innerText = pct + '%%';
    el.style.opacity = '1';
    clearTimeout(window._zoomTimer);
    window._zoomTimer = setTimeout(function(){ el.style.opacity = '0'; }, 1500);
  }
};
window.applyTheme = function(theme) {
  document.body.setAttribute('data-theme', theme || 'auto');
  // Switch highlight.js theme
  var isDark = theme === 'dark' || theme === 'solarized' || theme === 'nord';
  if (theme === 'auto') {
    isDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  var lightLink = document.getElementById('hljsLight');
  var darkLink = document.getElementById('hljsDark');
  if (lightLink && darkLink) {
    lightLink.disabled = isDark;
    darkLink.disabled = !isDark;
  }
  if (window.hljs) hljs.highlightAll();
};
window.applyLanguage = function(lang) {
  var translations = {
    'zhTW': { 'settings': '設定', 'theme': '佈景主題', 'zoom': '縮放靈敏度', 'zoomDesc': '控制 ⌘+ 和 ⌘- 每次調整的幅度', 'low': '低', 'mid': '中', 'high': '高', 'font': '字型', 'fontSize': '字型大小', 'lang': '語言 / Language', 'drop': '拖放 .md 檔案到此', 'noFile': '未載入檔案', 'hint': '按 ⌘O 開啟 Markdown 檔案' },
    'zhCN': { 'settings': '设置', 'theme': '佈景主题', 'zoom': '缩放灵敏度', 'zoomDesc': '控制 ⌘+ 和 ⌘- 每次调整的幅度', 'low': '低', 'mid': '中', 'high': '高', 'font': '字体', 'fontSize': '字体大小', 'lang': '语言 / Language', 'drop': '拖放 .md 文件到此', 'noFile': '未加载文件', 'hint': '按 ⌘O 打开 Markdown 文件' },
    'en':   { 'settings': 'Settings', 'theme': 'Theme', 'zoom': 'Zoom Sensitivity', 'zoomDesc': 'Control how much ⌘+/- zooms', 'low': 'Low', 'mid': 'Mid', 'high': 'High', 'font': 'Font Family', 'fontSize': 'Font Size', 'lang': 'Language', 'drop': 'Drop .md file here', 'noFile': 'No file loaded', 'hint': 'Press ⌘O to open a Markdown file' },
    'ja':   { 'settings': '設定', 'theme': 'テーマ', 'zoom': 'ズーム感度', 'zoomDesc': '⌘+/- の調整幅を制御します', 'low': '低', 'mid': '中', 'high': '高', 'font': 'フォント', 'fontSize': 'フォントサイズ', 'lang': '言語 / Language', 'drop': '.md ファイルをここにドロップ', 'noFile': 'ファイルが読み込まれていません', 'hint': '⌘O を押して Markdown ファイルを開く' },
    'ko':   { 'settings': '설정', 'theme': '테마', 'zoom': '확대/축소 민감도', 'zoomDesc': '⌘+/- 확대/축소 정도를 조절합니다', 'low': '낮음', 'mid': '중간', 'high': '높음', 'font': '글꼴', 'fontSize': '글꼴 크기', 'lang': '언어 / Language', 'drop': '.md 파일을 여기에 드롭', 'noFile': '파일이 로드되지 않음', 'hint': '⌘O를 눌러 Markdown 파일을 엽니다' }
  };
  var t = translations[lang] || translations['en'];
  // Settings panel
  var sTitle = document.querySelector('.settings-title'); if (sTitle) sTitle.firstChild.textContent = t.settings + ' ';
  var labels = document.querySelectorAll('.setting-label');
  if (labels.length >= 5) {
    labels[0].textContent = t.theme;
    labels[1].textContent = t.zoom;
    labels[2].textContent = t.font;
    labels[3].textContent = t.fontSize;
    labels[4].textContent = t.lang;
  }
  var zoomDesc = document.querySelector('.setting-desc'); if (zoomDesc) zoomDesc.textContent = t.zoomDesc;
  var sliderLabels = document.querySelectorAll('.slider-labels span');
  if (sliderLabels.length >= 3) {
    sliderLabels[0].textContent = t.low;
    sliderLabels[1].textContent = t.mid;
    sliderLabels[2].textContent = t.high;
  }
  // Empty state & Drop zone
  var dzMsg = document.querySelector('.drop-zone-msg'); if (dzMsg) dzMsg.textContent = t.drop;
  var esTitle = document.querySelector('.empty-state-title'); if (esTitle) esTitle.textContent = t.noFile;
  var esHint = document.querySelector('.empty-state div:last-child'); if (esHint) esHint.textContent = t.hint;
};
window.showSettingsPanel = function() {
  // Delegate to toggle (opens if closed)
  var el = document.getElementById('settingsOverlay');
  if (!el || el.style.display === 'none' || el.style.display === '') {
    window.toggleSettingsPanel && window.toggleSettingsPanel();
  }
};
window.hideSettingsPanel = function() {
  var el = document.getElementById('settingsOverlay');
  if (el) el.style.display = 'none';
};
window.triggerReloadFlash = function() {
  var body = document.body;
  body.classList.remove('reload-flash');
  void body.offsetWidth; // 強制重繪以重啟動畫
  body.classList.add('reload-flash');
};
window.toggleSettingsPanel = function() {
  var el = document.getElementById('settingsOverlay');
  if (el) {
    var wasHidden = el.style.display === 'none' || el.style.display === '';
    el.style.display = wasHidden ? 'flex' : 'none';
    // When opening, sync values with config
    if (wasHidden) {
      // Try to get fresh config from Go
      var cfg = window.mdConfig || window._initConfig;
      
      var slider = document.getElementById('zoomSensitivity');
      if (slider && cfg) {
        var targetVal = cfg.zoomSensitivity || 5;
        slider.value = targetVal;
        slider.setAttribute('value', targetVal);
        // WebKit workaround: remove and re-add to force repaint
        var parent = slider.parentNode;
        if (parent) {
          var next = slider.nextSibling;
          parent.removeChild(slider);
          parent.insertBefore(slider, next);
        }
      }
      var ts = document.getElementById('themeSelect');
      if (ts && cfg) ts.value = cfg.theme || 'auto';
      var ff = document.getElementById('fontFamily');
      if (ff && cfg) {
        var curr = cfg.fontFamily || '';
        var opts = ff.options;
        for (var i = 0; i < opts.length; i++) {
          if (curr.indexOf(opts[i].value) !== -1) { ff.selectedIndex = i; break; }
        }
      }
      var fs = document.getElementById('fontSize');
      if (fs && cfg) fs.value = cfg.fontSize || 16;
      var lang = document.getElementById('language');
      if (lang && cfg) lang.value = cfg.language || 'zhTW';
    }
  }
};
window.applyZoomSensitivity = function(level) {
  window.zoomState.step = 0.02 * level;
};
window.zoomIn = function() { window.applyZoomLevel(window.zoomState.level + window.zoomState.step); };
window.zoomOut = function() { window.applyZoomLevel(window.zoomState.level - window.zoomState.step); };
window.zoomReset = function() { window.applyZoomLevel(1.0); };

document.addEventListener('DOMContentLoaded', function() {
  // File drop handling
  var dropZone = document.getElementById('dropZone');
  var lastEnterTarget = null;

  document.addEventListener('dragenter', function(e) {
    e.preventDefault();
    lastEnterTarget = e.target;
    dropZone.classList.add('active');
    dropZone.classList.remove('error');
    dropZone.querySelector('.drop-zone-msg').textContent = 'Drop .md file here';
  }, false);

  document.addEventListener('dragleave', function(e) {
    e.preventDefault();
    // Only hide if leaving the drop zone (not entering a child element)
    if (e.target === dropZone || !dropZone.contains(e.relatedTarget)) {
      dropZone.classList.remove('active');
    }
  }, false);

  document.addEventListener('dragover', function(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
  }, false);

  document.addEventListener('drop', function(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.classList.remove('active');

    var files = e.dataTransfer.files;
    if (files.length === 0) {
      var fileUrl = e.dataTransfer.getData('text/uri-list') || e.dataTransfer.getData('text/plain');
      if (fileUrl && fileUrl.indexOf('file://') === 0) {
        var fp = decodeURIComponent(fileUrl.replace('file://', ''));
        if (window.openFileByPath) window.openFileByPath(fp);
        return;
      }
      dropZone.classList.add('error');
      dropZone.classList.add('active');
      dropZone.querySelector('.drop-zone-msg').textContent = 'No file detected. Use ⌘O to open.';
      setTimeout(function() { dropZone.classList.remove('active', 'error'); }, 2000);
      return;
    }

    var file = files[0];
    var name = file.name.toLowerCase();
    if (!name.endsWith('.md') && !name.endsWith('.markdown') && !name.endsWith('.txt')) {
      dropZone.classList.add('error');
      dropZone.classList.add('active');
      dropZone.querySelector('.drop-zone-msg').textContent = 'Please drop a .md file';
      setTimeout(function() { dropZone.classList.remove('active', 'error'); }, 1500);
      return;
    }

    // Use FileReader to read content, pass to Go via binding
    var reader = new FileReader();
    reader.onload = function(evt) {
      if (window.loadFileContent) {
        window.loadFileContent(file.name, evt.target.result);
      }
    };
    reader.onerror = function() {
      dropZone.classList.add('error');
      dropZone.classList.add('active');
      dropZone.querySelector('.drop-zone-msg').textContent = 'Failed to read file';
      setTimeout(function() { dropZone.classList.remove('active', 'error'); }, 2000);
    };
    reader.readAsText(file);
  }, false);

  // Trackpad pinch-to-zoom (WebKit gesture events)
  var lastDist = 0;
  document.addEventListener('gesturestart', function(e){ e.preventDefault(); lastDist = 0; }, {passive:false});
  document.addEventListener('gesturechange', function(e){
    e.preventDefault();
    if (lastDist > 0) window.applyZoomLevel(window.zoomState.level * (e.scale / lastDist));
    lastDist = e.scale;
  }, {passive:false});
  document.addEventListener('gestureend', function(){ lastDist = 0; });

  // Ctrl+Wheel zoom
  document.addEventListener('wheel', function(e){
    if (e.ctrlKey || e.metaKey) { e.preventDefault(); window.applyZoomLevel(window.zoomState.level + (e.deltaY > 0 ? -window.zoomState.step : window.zoomState.step)); }
  }, {passive:false});

  // Keyboard shortcuts
  document.addEventListener('keydown', function(e){
    if (!e.metaKey) return;
    switch(e.key) {
      case '=': case '+': case '§': e.preventDefault(); window.zoomIn();  break;
      case '-': case '_':            e.preventDefault(); window.zoomOut(); break;
      case '0':                      e.preventDefault(); window.zoomReset(); break;
      case 'o': case 'O':            e.preventDefault(); window.openFile();   break;
      case 'r': case 'R':
        if (e.shiftKey) { e.preventDefault(); window.zoomReset(); }
        else { e.preventDefault(); window.reloadFile(); }
        break;
      case ',':
        e.preventDefault(); window.toggleSettingsPanel && window.toggleSettingsPanel();
        break;
    }
  });

  // Apply on init: check _initConfig (injected into HTML) or mdConfig (set via Eval)
  var _cfg = window._initConfig || window.mdConfig;
  if (_cfg) {
    window.mdConfig = _cfg; // ensure mdConfig is available
    window.applyZoomSensitivity(_cfg.zoomSensitivity || 5);
    window.applyZoomLevel(_cfg.zoomLevel || 1.0);
    window.applyTheme(_cfg.theme || 'auto');
    window.applyLanguage(_cfg.language || 'zhTW');
  } else {
    window.applyZoomSensitivity(5);
    window.applyZoomLevel(1.0);
    window.applyTheme('auto');
    window.applyLanguage('zhTW');
  }
  if (window.hljs) hljs.highlightAll();
});
  var themeSelect = document.getElementById('themeSelect');
  if (themeSelect) {
    themeSelect.addEventListener('change', function() {
      var val = this.value;
      window.applyTheme(val);
      if (window.saveTheme) window.saveTheme(val);
      if (window.mdConfig) window.mdConfig.theme = val;
    });
  }
  var zoomSlider = document.getElementById('zoomSensitivity');
  if (zoomSlider) {
    zoomSlider.addEventListener('input', function() {
      var level = parseInt(this.value, 10);
      if (window.saveZoomSensitivity) window.saveZoomSensitivity(level);
      if (window.applyZoomSensitivity) window.applyZoomSensitivity(level);
      if (window.mdConfig) window.mdConfig.zoomSensitivity = level;
    });
  }
  var fontFamilySelect = document.getElementById('fontFamily');
  if (fontFamilySelect) {
    fontFamilySelect.addEventListener('change', function() {
      var val = this.value;
      document.body.style.fontFamily = val;
      // Read current fontSize from the select element, not from mdConfig
      var fsEl = document.getElementById('fontSize');
      var currentSize = fsEl ? parseInt(fsEl.value, 10) : 16;
      if (window.saveFont) {
        window.saveFont(val, currentSize);
      }
      // Update local mdConfig
      if (window.mdConfig) window.mdConfig.fontFamily = val;
    });
  }
  var fontSizeSelect = document.getElementById('fontSize');
  if (fontSizeSelect) {
    fontSizeSelect.addEventListener('change', function() {
      var val = parseInt(this.value, 10);
      document.body.style.fontSize = val + 'px';
      // Save both fontFamily AND fontSize together
      var ff = document.getElementById('fontFamily');
      var fv = ff ? ff.value : (window.mdConfig && window.mdConfig.fontFamily) || '-apple-system, BlinkMacSystemFont, Segoe UI, Helvetica, Arial, sans-serif';
      if (window.saveFont) window.saveFont(fv, val);
      // Update local mdConfig
      if (window.mdConfig) { window.mdConfig.fontFamily = fv; window.mdConfig.fontSize = val; }
    });
  }
  var langSelect = document.getElementById('language');
  if (langSelect) {
    langSelect.addEventListener('change', function() {
      var val = this.value;
      if (window.saveLanguage) window.saveLanguage(val);
      if (window.applyLanguage) window.applyLanguage(val);
      if (window.mdConfig) window.mdConfig.language = val;
    });
  }
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') window.hideSettingsPanel && window.hideSettingsPanel();
  });
  var closeBtn = document.getElementById('settingsClose');
  if (closeBtn) closeBtn.addEventListener('click', window.hideSettingsPanel);

  // Intercept all link clicks and handle appropriately
  document.addEventListener('click', function(e) {
    var target = e.target;
    while (target && target.tagName !== 'A') {
      target = target.parentNode;
    }
    if (target && target.tagName === 'A') {
      var href = target.getAttribute('href');
      if (!href || href.startsWith('#')) return;

      if (href.startsWith('http://') || href.startsWith('https://')) {
        e.preventDefault();
        if (window.openExternal) window.openExternal(target.href);
      } else if (href.startsWith('file://')) {
        e.preventDefault();
        var path = decodeURIComponent(href.replace('file://', ''));
        // If it's a markdown file, open it in-app, otherwise use system
        if (path.toLowerCase().endsWith('.md') || path.toLowerCase().endsWith('.markdown')) {
          if (window.openFileByPath) window.openFileByPath(path);
        } else {
          if (window.openExternal) window.openExternal(href);
        }
      } else if (!href.includes('://')) {
        // Likely a relative local link
        e.preventDefault();
        if (window.handleRelativeLink) window.handleRelativeLink(href);
      }
    }
  });
</script>
</body>
</html>`

var (
	currentWV   webview.WebView
	renderer     *core.MarkdownRenderer
	currentFile  string
	wv           webview.WebView
)

func prepareHTML(html string) string {
	// Inject config directly into HTML so IIFE can read it immediately
	html = strings.Replace(html, "</head>", fmt.Sprintf(`<script>window._initConfig=%s;</script></head>`, ConfigToJS()), 1)
	// Set theme directly on body to avoid flash/revert bug
	theme := currentConfig.Theme
	if theme == "" {
		theme = "auto"
	}
	html = strings.Replace(html, `data-theme="auto"`, fmt.Sprintf(`data-theme=%q`, theme), 1)
	return html
}

func renderMD(md string) string {
	html, err := renderer.Render(md)
	if err != nil {
		return renderError(err.Error())
	}
	res := prepareHTML(fmt.Sprintf(htmlTemplate, cssContent, html))
	if pendingFlash {
		// 使用正規表示式或更寬鬆的替換，確保 class 被加上去
		res = strings.Replace(res, "<body ", "<body class=\"reload-flash\" ", 1)
		pendingFlash = false
	}
	return res
}

func renderEmpty() string {
	content := `<div class="empty-state"><div class="empty-state-icon">📄</div><div class="empty-state-title">No file loaded</div><div>Press ⌘O to open a Markdown file</div></div>`
	return prepareHTML(fmt.Sprintf(htmlTemplate, cssContent, content))
}

func renderError(msg string) string {
	content := fmt.Sprintf(`<div class="error"><strong>Error:</strong> %s</div>`, msg)
	return prepareHTML(fmt.Sprintf(htmlTemplate, cssContent, content))
}

// getConfigJS returns the config JS snippet (call after LoadConfig)
func getConfigJS() string {
	return ConfigToJS()
}

func updateWatcher(path string) {
	watchMu.Lock()
	defer watchMu.Unlock()

	if watcher == nil {
		var err error
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating watcher:", err)
			return
		}

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					// 監聽寫入、重新建立或重新命名 (原子存檔) 事件
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
						watchMu.Lock()
						
						// 如果是 Rename，表示舊的 fd 可能失效，需要稍後重新 Add
						isRename := event.Has(fsnotify.Rename)
						
						if debounceTimer != nil {
							debounceTimer.Stop()
						}
						
						debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
							watchMu.Lock()
							defer watchMu.Unlock()
							
							if isRename && currentFile != "" {
								// 重新將檔案加入監聽
								watcher.Remove(event.Name)
								watcher.Add(currentFile)
							}
							
							if currentWV != nil {
								currentWV.Dispatch(func() {
									reloadFile()
								})
							}
						})
						watchMu.Unlock()
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					fmt.Fprintln(os.Stderr, "Watcher error:", err)
				}
			}
		}()
	}

	// 移除舊的監聽
	for _, w := range watcher.WatchList() {
		watcher.Remove(w)
	}

	// 監聽新的檔案
	if path != "" && !strings.HasPrefix(path, "(dragged)") {
		err := watcher.Add(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error watching %s: %v\n", path, err)
		}
	}
}

func openFile() {
	script := `set f to POSIX path of (choose file of type {"md","markdown","txt","public.plain-text","public.text"} with prompt "Select Markdown file") in "::empty::"`
	out, _ := exec.Command("osascript", "-e", script).Output()
	name := strings.TrimSpace(string(out))
	if name != "" && name != "::empty::" {
		loadFile(name)
	}
}

func loadFile(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if currentWV != nil {
			currentWV.SetHtml(renderError("Cannot read: " + err.Error()))
		}
		return
	}

	currentFile = absPath
	updateWatcher(absPath)
	// Keep exportContent in sync for export feature (supports drag-and-drop scenarios)
	exportContent = string(data)
	if currentWV == nil {
		return
	}

	currentWV.SetTitle(filepath.Base(absPath) + " - md-viewer")
	
	if len(data) > 500*1024 {
		currentWV.SetHtml(prepareHTML(`<div class="empty-state"><div class="empty-state-icon">⏳</div><div class="empty-state-title">Rendering Large File...</div><div>Please wait while we process the content.</div></div>`))
	}

	go func(content string) {
		htmlOutput := renderMD(content)
		currentWV.Dispatch(func() {
			if currentWV != nil {
				currentWV.SetHtml(htmlOutput)
			}
		})
	}(string(data))
}

func reloadFile() {
	if currentFile != "" {
		pendingFlash = true
		loadFile(currentFile)
	}
}

func main() {
	renderer = core.NewMarkdownRenderer()

	if err := LoadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
	}
	UpdateMenuLanguage(currentConfig.Language)

	if len(os.Args) > 1 {
		currentFile = os.Args[1]
		abs, err := filepath.Abs(currentFile)
		if err == nil {
			updateWatcher(abs)
		}
	}

	title := "md-viewer"
	if currentFile != "" {
		title = filepath.Base(currentFile) + " - md-viewer"
	}

	wv = webview.New(true)
	if wv == nil {
		fmt.Fprintln(os.Stderr, "Failed to create webview")
		os.Exit(1)
	}
	currentWV = wv
	defer wv.Destroy()
	defer func() {
		watchMu.Lock()
		if watcher != nil {
			watcher.Close()
		}
		watchMu.Unlock()
	}()

	SetupMenu(func(menuID int) {
		fmt.Fprintf(os.Stderr, "[MENU] callback fired: menuID=%d\n", menuID)
		switch menuID {
		case MenuPreferences:
			wv.Eval("window.toggleSettingsPanel && window.toggleSettingsPanel()")
		case MenuOpen:
			wv.Dispatch(func() { openFile() })
		case MenuReload:
			wv.Dispatch(func() { reloadFile() })
		case MenuQuit:
			os.Exit(0)
		case MenuZoomIn:
			wv.Eval("window.zoomIn && window.zoomIn()")
		case MenuZoomOut:
			wv.Eval("window.zoomOut && window.zoomOut()")
		case MenuZoomReset:
			wv.Eval("window.zoomReset && window.zoomReset()")
		case MenuExportHTML:
			exportHTML()
		case MenuExportPDF:
			exportPDF()
		}
	})

	wv.SetTitle(title)
	wv.SetSize(1200, 800, webview.HintNone)

	// Handle macOS double-click / open file events
	RegisterOpenFileCallback(func(path string) {
		wv.Dispatch(func() { loadFile(path) })
	})

	// Enable native macOS drag & drop (for intra-app drags; Finder→App uses JS HTML5 drop)
	RegisterDragDropCallback(func(path string) {
		wv.Dispatch(func() { loadFile(path) })
	})
	windowPtr := wv.Window()
	if windowPtr != nil {
		overlay := CreateDragDropOverlay(windowPtr)
		if overlay != nil {
			// Keep overlay reference alive (it's retained by the window system)
		}
	}

	wv.Bind("openFile",   func() { wv.Dispatch(func(){ openFile() }) })
	wv.Bind("reloadFile", func() { wv.Dispatch(func(){ reloadFile() }) })
	wv.Bind("openFileByPath", func(path string) {
		wv.Dispatch(func() { loadFile(path) })
	})
	wv.Bind("openExternal", func(url string) {
		exec.Command("open", url).Run()
	})
	wv.Bind("handleRelativeLink", func(relPath string) {
		if currentFile == "" || strings.HasPrefix(currentFile, "(dragged)") {
			return
		}
		baseDir := filepath.Dir(currentFile)
		absPath := filepath.Join(baseDir, relPath)
		wv.Dispatch(func() { loadFile(absPath) })
	})
	wv.Bind("loadFileContent", func(filename, content string) {
		currentFile = "(dragged): " + filename
		exportContent = content // keep exportContent in sync for export feature
		if currentWV != nil {
			currentWV.SetTitle(filename + " - md-viewer")
			currentWV.SetHtml(renderMD(content))
		}
	})
	wv.Bind("saveZoomSensitivity", func(level int) {
		if err := SetZoomSensitivity(level); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config:", err)
		}
	})
	wv.Bind("saveZoomLevel", func(level float64) {
		if err := SetZoomLevel(level); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save zoom level:", err)
		}
	})
	wv.Bind("saveTheme", func(theme string) {
		if err := SetTheme(theme); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save theme:", err)
		}
	})
	wv.Bind("saveFont", func(family string, size int) {
		if err := SetFont(family, size); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save font:", err)
		}
	})
	wv.Bind("saveLanguage", func(lang string) {		if err := SetLanguage(lang); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save language:", err)
		} else {
			UpdateMenuLanguage(lang)
			wv.Eval(getConfigJS())
		}
	})

	// config is injected into htmlTemplate via _initConfig script
	if currentFile == "" {
		wv.SetHtml(renderEmpty())
	} else {
		wv.SetHtml(renderMD("Loading..."))
		wv.Dispatch(func() { loadFile(currentFile) })
	}

	wv.Run()
}
