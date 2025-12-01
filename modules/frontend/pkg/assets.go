package frontend

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AssetManager manages CSS and JS assets
type AssetManager struct {
	assetsPath string
	manifest   map[string]string
}

// NewAssetManager creates a new asset manager
func NewAssetManager() *AssetManager {
	return &AssetManager{
		assetsPath: "./public",
		manifest:   make(map[string]string),
	}
}

// GetAssetURL returns the URL for an asset with cache busting
func (am *AssetManager) GetAssetURL(path string) string {
	// Check if file exists
	fullPath := filepath.Join(am.assetsPath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return path
	}

	// Generate hash for cache busting
	hash := am.getFileHash(fullPath)
	if hash != "" {
		if strings.Contains(path, "?") {
			return fmt.Sprintf("%s&v=%s", path, hash[:8])
		}
		return fmt.Sprintf("%s?v=%s", path, hash[:8])
	}

	return path
}

// getFileHash generates MD5 hash of file content
func (am *AssetManager) getFileHash(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Bundle combines multiple assets into one
func (am *AssetManager) Bundle(assetType string, files []string) (string, error) {
	var content strings.Builder

	for _, file := range files {
		fullPath := filepath.Join(am.assetsPath, file)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		content.Write(data)
		content.WriteString("\n")
	}

	// Save bundled file
	bundlePath := filepath.Join(am.assetsPath, assetType, fmt.Sprintf("bundle.%s", assetType))
	if err := os.WriteFile(bundlePath, []byte(content.String()), 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("/%s/bundle.%s", assetType, assetType), nil
}

// Minify minifies CSS or JS content (simplified version)
func (am *AssetManager) Minify(content string, assetType string) string {
	// Remove comments and extra whitespace
	lines := strings.Split(content, "\n")
	var minified strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "//") {
			minified.WriteString(line)
			minified.WriteString(" ")
		}
	}

	return strings.TrimSpace(minified.String())
}
