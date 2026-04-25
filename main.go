package main

import (
	"C"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"md-viewer-app/core"
	"github.com/webview/webview_go"
)

const cssContent = `
:root {
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
@media (prefers-color-scheme: dark) {
  :root {
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
.markdown-body { max-width: 900px; margin: 0 auto; }
.markdown-body h1 { font-size: 2em; border-bottom: 1px solid var(--color-border-default); padding-bottom: 0.3em; margin-bottom: 1em; margin-top: 1.5em; }
.markdown-body h2 { font-size: 1.5em; border-bottom: 1px solid var(--color-border-default); padding-bottom: 0.3em; margin-bottom: 0.8em; margin-top: 1.5em; }
.markdown-body h3 { font-size: 1.25em; margin-bottom: 0.6em; margin-top: 1.2em; }
.markdown-body h4 { font-size: 1em; margin-bottom: 0.5em; margin-top: 1em; }
.markdown-body h1:first-child, .markdown-body h2:first-child, .markdown-body h3:first-child { margin-top: 0; }
.markdown-body p { margin-bottom: 1em; }
.markdown-body a { color: var(--color-accent-fg); text-decoration: none; }
.markdown-body a:hover { text-decoration: underline; }
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
  padding: 4px 8px; border-radius: 4px; opacity: 0.7;
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
<style>%s</style>
</head>
<body>
<div class="drop-zone" id="dropZone"><div class="drop-zone-msg">Drop .md file here</div></div>
<div class="markdown-body" id="mdContent">%s</div>
<div class="keyboard-hint">⌘O Open | ⌘R Reload | ⌘+/⌘- Zoom | ⌘0 Reset | ⌘, Settings | ⌘Q Quit</div>
<div class="settings-overlay" id="settingsOverlay" style="display:none">
  <div class="settings-panel">
    <div class="settings-title">
      設定
      <button class="settings-close" id="settingsClose" title="關閉 (Esc)">×</button>
    </div>
    <div class="setting-row">
      <div class="setting-label">縮放靈敏度</div>
      <div class="setting-desc">控制 ⌘+ 和 ⌘- 每次調整的幅度</div>
      <input type="range" id="zoomSensitivity" min="1" max="3" value="2" step="1">
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
  </div>
</div>
<script>
(function() {
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

  // Zoom state
  var zoom = 1.0;
  var minZoom = 0.5;
  var maxZoom = 2.0;
  var step = 0.1;

  // Expose zoom state on window for external access (settings panel)
  window.zoomState = { level: 1.0, step: 0.1 };

  function applyZoom(level) {
    level = Math.max(minZoom, Math.min(maxZoom, level));
    window.zoomState.level = level;
    document.body.style.zoom = level;
    if (window._setZoomLevel) window._setZoomLevel(level);
    showZoomIndicator(Math.round(level * 100));
  }
  function zoomIn()  { applyZoom(window.zoomState.level + window.zoomState.step); }
  function zoomOut() { applyZoom(window.zoomState.level - window.zoomState.step); }
  function zoomReset(){ applyZoom(1.0); }

  // Trackpad pinch-to-zoom (WebKit gesture events)
  var lastDist = 0;
  document.addEventListener('gesturestart', function(e){ e.preventDefault(); lastDist = 0; }, {passive:false});
  document.addEventListener('gesturechange', function(e){
    e.preventDefault();
    if (lastDist > 0) applyZoom(window.zoomState.level * (e.scale / lastDist));
    lastDist = e.scale;
  }, {passive:false});
  document.addEventListener('gestureend', function(){ lastDist = 0; });

  // Ctrl+Wheel zoom
  document.addEventListener('wheel', function(e){
    if (e.ctrlKey || e.metaKey) { e.preventDefault(); applyZoom(window.zoomState.level + (e.deltaY > 0 ? -window.zoomState.step : window.zoomState.step)); }
  }, {passive:false});

  // Keyboard shortcuts
  document.addEventListener('keydown', function(e){
    if (!e.metaKey) return;
    switch(e.key) {
      case '=': case '+': case '§': e.preventDefault(); zoomIn();  break;
      case '-': case '_':            e.preventDefault(); zoomOut(); break;
      case '0':                      e.preventDefault(); zoomReset(); break;
      case 'o': case 'O':            e.preventDefault(); window.openFile();   break;
      case 'r': case 'R':            e.preventDefault(); window.reloadFile(); break;
    }
  });

  window.zoomIn  = zoomIn;
  window.zoomOut = zoomOut;
  window.zoomReset = zoomReset;

  // Zoom sensitivity: apply from config
  window.applyZoomSensitivity = function(level) {
    window.zoomState.step = [0, 0.05, 0.10, 0.20][level] || 0.10;
  };
  // Apply on init (called after window.mdConfig is injected)
  window.applyZoomSensitivity(window.mdConfig && window.mdConfig.zoomSensitivity || 2);
  if (window.mdConfig && window.mdConfig.zoomLevel) {
    window.applyZoomLevel(window.mdConfig.zoomLevel);
  }

  // Zoom indicator
  var timer;
  function showZoomIndicator(pct) {
    var el = document.getElementById('zoomIndicator');
    if (!el) {
      el = document.createElement('div');
      el.id = 'zoomIndicator';
      el.style.cssText = 'position:fixed;top:10px;right:10px;background:rgba(0,0,0,0.8);color:white;padding:5px 12px;border-radius:6px;font-size:13px;font-weight:bold;z-index:99999;pointer-events:none;transition:opacity 0.3s;';
      document.body.appendChild(el);
    }
    el.textContent = pct + '%';
    el.style.opacity = '1';
    clearTimeout(timer);
    timer = setTimeout(function(){ el.style.opacity = '0'; }, 1500);
  }
})();
  // Settings panel
  window.showSettingsPanel = function() {
    var el = document.getElementById('settingsOverlay');
    if (el) {
      el.style.display = 'flex';
      var slider = document.getElementById('zoomSensitivity');
      if (slider && window.mdConfig) slider.value = window.mdConfig.zoomSensitivity || 2;
      var ff = document.getElementById('fontFamily');
      if (ff && window.mdConfig) {
        var curr = (window.mdConfig.fontFamily || '').toLowerCase();
        var opts = ff.options;
        for (var i = 0; i < opts.length; i++) {
          if (curr.indexOf(opts[i].value.split(',')[0].trim().replace(/'/g,'')) !== -1) {
            ff.selectedIndex = i; break;
          }
        }
      }
      var fs = document.getElementById('fontSize');
      if (fs && window.mdConfig) fs.value = window.mdConfig.fontSize || 16;
      var lang = document.getElementById('language');
      if (lang && window.mdConfig) {
        lang.value = window.mdConfig.language || 'zhTW';
      }
    }
  };
  var zoomSlider = document.getElementById('zoomSensitivity');
  if (zoomSlider) {
    zoomSlider.addEventListener('change', function() {
      var level = parseInt(this.value, 10);
      if (window.saveZoomSensitivity) {
        window.saveZoomSensitivity(level);
      }
      window.applyZoomSensitivity && window.applyZoomSensitivity(level);
    });
  }
  var fontFamilySelect = document.getElementById('fontFamily');
  if (fontFamilySelect) {
    fontFamilySelect.addEventListener('change', function() {
      var val = this.value;
      document.body.style.fontFamily = val;
      if (window.saveFont) window.saveFont(val, (window.mdConfig && window.mdConfig.fontSize) || 16);
    });
  }
  var fontSizeSelect = document.getElementById('fontSize');
  if (fontSizeSelect) {
    fontSizeSelect.addEventListener('change', function() {
      var val = parseInt(this.value, 10);
      document.body.style.fontSize = val + 'px';
      if (window.saveFont) window.saveFont(
        (window.mdConfig && window.mdConfig.fontFamily) || '-apple-system, BlinkMacSystemFont, Segoe UI, Helvetica, Arial, sans-serif', val);
    });
  }
  var langSelect = document.getElementById('language');
  if (langSelect) {
    langSelect.addEventListener('change', function() {
      var val = this.value;
      if (window.saveLanguage) window.saveLanguage(val);
      if (window.applyLanguage) window.applyLanguage(val);
    });
  }
  window.hideSettingsPanel = function() {
    var el = document.getElementById('settingsOverlay');
    if (el) { el.style.display = 'none'; }
  };
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') window.hideSettingsPanel();
  });
  var closeBtn = document.getElementById('settingsClose');
  if (closeBtn) closeBtn.addEventListener('click', window.hideSettingsPanel);
</script>
</body>
</html>`

var (
	currentWV   webview.WebView
	renderer     *core.MarkdownRenderer
	currentFile  string
)

func renderMD(md string) string {
	html, err := renderer.Render(md)
	if err != nil {
		return fmt.Sprintf(htmlTemplate, cssContent, fmt.Sprintf(`<div class="error"><strong>Error:</strong> %s</div>`, err.Error()))
	}
	return fmt.Sprintf(htmlTemplate, cssContent, html)
}

func renderEmpty() string {
	return fmt.Sprintf(htmlTemplate, cssContent, `<div class="empty-state"><div class="empty-state-icon">📄</div><div class="empty-state-title">No file loaded</div><div>Press ⌘O to open a Markdown file</div></div>`)
}

func renderError(msg string) string {
	return fmt.Sprintf(htmlTemplate, cssContent, fmt.Sprintf(`<div class="error"><strong>Error:</strong> %s</div>`, msg))
}

// getConfigJS returns the config JS snippet (call after LoadConfig)
func getConfigJS() string {
	return ConfigToJS()
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
	data, err := os.ReadFile(path)
	if err != nil {
		if currentWV != nil {
			currentWV.Eval("window._savedZoom = {level: window.zoomState && window.zoomState.level || 1.0, step: window.zoomState && window.zoomState.step || 0.1};")
		currentWV.SetHtml(renderError("Cannot read: " + err.Error()))
		currentWV.Eval(getConfigJS())
		currentWV.Eval("if(window.mdConfig && window.mdConfig.zoomLevel){window.applyZoomLevel && window.applyZoomLevel(window.mdConfig.zoomLevel)}")
		currentWV.Eval("if(window._savedZoom){window.zoomState=window._savedZoom;document.body.style.zoom=window.zoomState.level;}")
		}
		return
	}
	currentFile = path
	if currentWV != nil {
		currentWV.Eval("window._savedZoom = {level: window.zoomState && window.zoomState.level || 1.0, step: window.zoomState && window.zoomState.step || 0.1};")
		currentWV.SetTitle(filepath.Base(path) + " - md-viewer")
		currentWV.SetHtml(renderMD(string(data)))
		currentWV.Eval(getConfigJS())
		currentWV.Eval("if(window.mdConfig && window.mdConfig.zoomLevel){window.applyZoomLevel && window.applyZoomLevel(window.mdConfig.zoomLevel)}")
	}
}

func reloadFile() {
	if currentFile != "" {
		loadFile(currentFile)
	}
}

func main() {
	renderer = core.NewMarkdownRenderer()

	if err := LoadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
	}

	if len(os.Args) > 1 {
		currentFile = os.Args[1]
	}

	title := "md-viewer"
	if currentFile != "" {
		title = filepath.Base(currentFile) + " - md-viewer"
	}

	wv := webview.New(true)
	if wv == nil {
		fmt.Fprintln(os.Stderr, "Failed to create webview")
		os.Exit(1)
	}
	currentWV = wv
	defer wv.Destroy()

	SetupMenu(func(menuID int) {
		switch menuID {
		case MenuPreferences:
			wv.Eval("showSettingsPanel && showSettingsPanel()")
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
		}
	})

	wv.SetTitle(title)
	wv.SetSize(1200, 800, webview.HintNone)

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
	wv.Bind("loadFileContent", func(filename, content string) {
		currentFile = "(dragged): " + filename
		if currentWV != nil {
			currentWV.Eval("window._savedZoom = {level: window.zoomState && window.zoomState.level || 1.0, step: window.zoomState && window.zoomState.step || 0.1};")
			currentWV.SetTitle(filename + " - md-viewer")
			currentWV.SetHtml(renderMD(content))
			currentWV.Eval(getConfigJS())
			currentWV.Eval("if(window.mdConfig && window.mdConfig.zoomLevel){window.applyZoomLevel && window.applyZoomLevel(window.mdConfig.zoomLevel)}")
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
	wv.Bind("saveFont", func(family string, size int) {
		if err := SetFont(family, size); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save font:", err)
		}
	})
	wv.Bind("saveLanguage", func(lang string) {
		if err := SetLanguage(lang); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save language:", err)
		} else {
			wv.Eval(getConfigJS())
		}
	})

	// config is injected into htmlTemplate via fmt.Sprintf below
	if currentFile == "" {
		wv.SetHtml(renderEmpty())
	wv.Eval(getConfigJS())
	} else {
		// Load file after webview is ready
			wv.SetHtml(renderMD("Loading..."))
		wv.Eval(getConfigJS())
			wv.Dispatch(func() { loadFile(currentFile) })
	}

	wv.Run()
}
