package frontend

import (
	"fmt"
	"os"
	"path/filepath"
)

// Theme represents a frontend theme
type Theme struct {
	Name        string
	Path        string
	Version     string
	Author      string
	Description string
	Assets      *ThemeAssets
}

// ThemeAssets contains theme asset paths
type ThemeAssets struct {
	CSS []string
	JS  []string
}

// ThemeManager manages frontend themes
type ThemeManager struct {
	activeTheme *Theme
	themes      map[string]*Theme
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	return &ThemeManager{
		themes: make(map[string]*Theme),
	}
}

// LoadTheme loads a theme by name
func (tm *ThemeManager) LoadTheme(name string) error {
	themePath := filepath.Join("templates", "frontend", name)

	// Check if theme exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return fmt.Errorf("theme '%s' not found", name)
	}

	theme := &Theme{
		Name:        name,
		Path:        themePath,
		Version:     "1.0.0",
		Author:      "NeoNex Technologies",
		Description: "Default NeonEx Framework theme",
		Assets: &ThemeAssets{
			CSS: []string{"/css/theme.css"},
			JS:  []string{"/js/theme.js"},
		},
	}

	tm.themes[name] = theme
	tm.activeTheme = theme

	return nil
}

// GetActiveTheme returns the currently active theme
func (tm *ThemeManager) GetActiveTheme() *Theme {
	return tm.activeTheme
}

// SetActiveTheme sets the active theme
func (tm *ThemeManager) SetActiveTheme(name string) error {
	if theme, exists := tm.themes[name]; exists {
		tm.activeTheme = theme
		return nil
	}
	return fmt.Errorf("theme '%s' not loaded", name)
}

// ListThemes returns all available themes
func (tm *ThemeManager) ListThemes() []*Theme {
	themes := make([]*Theme, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, theme)
	}
	return themes
}
