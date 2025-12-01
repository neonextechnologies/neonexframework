package frontend

import (
	"neonexframework/pkg/app"

	"github.com/gofiber/fiber/v2"
)

// FrontendModule handles template rendering and asset management
type FrontendModule struct {
	themeManager *ThemeManager
}

// New creates a new frontend module
func New() *FrontendModule {
	return &FrontendModule{
		themeManager: NewThemeManager(),
	}
}

// Name returns the module name
func (m *FrontendModule) Name() string {
	return "frontend"
}

// RegisterServices registers frontend services
func (m *FrontendModule) RegisterServices(c *app.Container) error {
	c.Provide(func() *ThemeManager { return m.themeManager })
	c.Provide(NewAssetManager)
	c.Provide(NewTemplateService)
	return nil
}

// RegisterRoutes registers frontend routes
func (m *FrontendModule) RegisterRoutes(router fiber.Router) error {
	// Serve static files
	router.Static("/css", "./public/css")
	router.Static("/js", "./public/js")
	router.Static("/images", "./public/images")
	router.Static("/uploads", "./public/uploads")

	// Homepage
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Render("frontend/home", fiber.Map{
			"Title": "Welcome to NeonEx Framework",
		})
	})

	return nil
}

// Boot initializes the frontend module
func (m *FrontendModule) Boot() error {
	// Load default theme
	return m.themeManager.LoadTheme("default")
}
