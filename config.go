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

// GetConfig returns a copy of the current config.
func GetConfig() Config {
	return currentConfig
}

// ConfigToJS returns a JS snippet that sets window.mdConfig.
func ConfigToJS() string {
	return fmt.Sprintf(
		`window.mdConfig = {zoomSensitivity: %d, theme: %q, zoomLevel: %f, fontFamily: %q, fontSize: %d, language: %q};`,
		currentConfig.ZoomSensitivity,
		currentConfig.Theme,
		currentConfig.ZoomLevel,
		currentConfig.FontFamily,
		currentConfig.FontSize,
		currentConfig.Language,
	)
}
