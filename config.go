package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Config holds user preferences
type Config struct {
	ZoomSensitivity int     `json:"zoomSensitivity"`
	Theme           string  `json:"theme"`
	ZoomLevel       float64 `json:"zoomLevel"`
	FontFamily      string  `json:"fontFamily"`
	FontSize        int     `json:"fontSize"`
	Language        string  `json:"language"`
	ShowLineNumbers bool    `json:"showLineNumbers"`
	WindowWidth     int     `json:"windowWidth"`
	WindowHeight    int     `json:"windowHeight"`
	WindowX         int     `json:"windowX"`
	WindowY         int     `json:"windowY"`
	LastOpenedFile  string  `json:"lastOpenedFile"`
	RecentFiles     []string `json:"recentFiles"`
}

var currentConfig Config

func configDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(usr.HomeDir, ".md-viewer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func defaultConfig() Config {
	return Config{
		ZoomSensitivity: 5,
		Theme:           "auto",
		ZoomLevel:       1.0,
		FontFamily:      "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
		FontSize:        16,
		Language:        "zhTW",
	}
}

// LoadConfig loads config from ~/.md-viewer/config.json.
// Creates with defaults if file does not exist.
func LoadConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			currentConfig = defaultConfig()
			return saveConfig() // create default file
		}
		return err
	}
	return json.Unmarshal(data, &currentConfig)
}

// saveConfig writes currentConfig to the config file.
func saveConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SetZoomSensitivity updates the zoom sensitivity and persists it.
func SetZoomSensitivity(level int) error {
	currentConfig.ZoomSensitivity = level
	return saveConfig()
}

// SetTheme updates the theme and persists it.
func SetTheme(theme string) error {
	currentConfig.Theme = theme
	return saveConfig()
}

// SetZoomLevel updates the zoom level and persists it.
func SetZoomLevel(level float64) error {
	currentConfig.ZoomLevel = level
	return saveConfig()
}

// SetFont updates the font and persists it.
func SetFont(family string, size int) error {
	currentConfig.FontFamily = family
	currentConfig.FontSize = size
	return saveConfig()
}

// SetLanguage updates the language and persists it.
func SetLanguage(lang string) error {
	currentConfig.Language = lang
	return saveConfig()
}

// SetLineNumbers updates the show line numbers preference and persists it.
func SetLineNumbers(show bool) error {
	currentConfig.ShowLineNumbers = show
	return saveConfig()
}

// SetWindowSize updates the window size and persists it.
func SetWindowSize(width, height int) error {
	currentConfig.WindowWidth = width
	currentConfig.WindowHeight = height
	return saveConfig()
}

// SetWindowPosition updates the window position and persists it.
func SetWindowPosition(x, y int) error {
	currentConfig.WindowX = x
	currentConfig.WindowY = y
	return saveConfig()
}

// SetLastOpenedFile updates the last opened file path and persists it.
func SetLastOpenedFile(path string) error {
	currentConfig.LastOpenedFile = path
	// Also add to recent files
	AddRecentFile(path)
	return saveConfig()
}

// AddRecentFile adds a file to the recent files list (max 10)
func AddRecentFile(path string) {
	if path == "" {
		return
	}
	// Remove if already exists
	for i, f := range currentConfig.RecentFiles {
		if f == path {
			currentConfig.RecentFiles = append(currentConfig.RecentFiles[:i], currentConfig.RecentFiles[i+1:]...)
			break
		}
	}
	// Add to front
	currentConfig.RecentFiles = append([]string{path}, currentConfig.RecentFiles...)
	// Keep only 10
	if len(currentConfig.RecentFiles) > 10 {
		currentConfig.RecentFiles = currentConfig.RecentFiles[:10]
	}
}

// GetRecentFiles returns the list of recent files
func GetRecentFiles() []string {
	return currentConfig.RecentFiles
}

// GetConfig returns a copy of the current config.
func GetConfig() Config {
	return currentConfig
}

// ConfigToJS returns a JS snippet that sets window.mdConfig.
func ConfigToJS() string {
	return fmt.Sprintf(
		`window.mdConfig = {zoomSensitivity: %d, theme: %q, zoomLevel: %f, fontFamily: %q, fontSize: %d, language: %q, showLineNumbers: %t};`,
		currentConfig.ZoomSensitivity,
		currentConfig.Theme,
		currentConfig.ZoomLevel,
		currentConfig.FontFamily,
		currentConfig.FontSize,
		currentConfig.Language,
		currentConfig.ShowLineNumbers,
	)
}
