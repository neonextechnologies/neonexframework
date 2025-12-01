package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

// TemplateService handles template rendering
type TemplateService struct {
	engine       *html.Engine
	themeManager *ThemeManager
	assetManager *AssetManager
	globalData   fiber.Map
}

// NewTemplateService creates a new template service
func NewTemplateService(tm *ThemeManager, am *AssetManager) *TemplateService {
	engine := html.New("./templates", ".html")

	// Add global functions
	engine.AddFunc("asset", am.GetAssetURL)

	return &TemplateService{
		engine:       engine,
		themeManager: tm,
		assetManager: am,
		globalData:   make(fiber.Map),
	}
}

// Render renders a template with data
func (ts *TemplateService) Render(c *fiber.Ctx, template string, data fiber.Map) error {
	// Merge global data
	for key, value := range ts.globalData {
		if _, exists := data[key]; !exists {
			data[key] = value
		}
	}

	// Add theme data
	if theme := ts.themeManager.GetActiveTheme(); theme != nil {
		data["Theme"] = theme
	}

	return c.Render(template, data)
}

// SetGlobal sets a global template variable
func (ts *TemplateService) SetGlobal(key string, value interface{}) {
	ts.globalData[key] = value
}

// GetEngine returns the template engine
func (ts *TemplateService) GetEngine() *html.Engine {
	return ts.engine
}
