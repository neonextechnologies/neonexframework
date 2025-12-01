package cms

import (
	"neonexcore/internal/core"
	"neonexframework/pkg/app"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CMSModule handles content management functionality
type CMSModule struct {
	db *gorm.DB
}

// New creates a new CMS module
func New() *CMSModule {
	return &CMSModule{}
}

// Name returns the module name
func (m *CMSModule) Name() string {
	return "cms"
}

// RegisterServices registers CMS services
func (m *CMSModule) RegisterServices(c *app.Container) error {
	// Register repositories
	c.Provide(NewPageRepository)
	c.Provide(NewPostRepository)
	c.Provide(NewCategoryRepository)
	c.Provide(NewMediaRepository)

	// Register services
	c.Provide(NewPageService)
	c.Provide(NewPostService)
	c.Provide(NewCategoryService)
	c.Provide(NewMediaService)

	// Register controllers
	c.Provide(NewPageController)
	c.Provide(NewPostController)
	c.Provide(NewCategoryController)
	c.Provide(NewMediaController)

	return nil
}

// RegisterRoutes registers CMS routes
func (m *CMSModule) RegisterRoutes(router fiber.Router) error {
	// API routes
	api := router.Group("/api/v1/cms")

	// Pages
	pages := api.Group("/pages")
	pageCtrl := core.Resolve[*PageController]()
	pages.Get("/", pageCtrl.List)
	pages.Get("/:id", pageCtrl.Get)
	pages.Post("/", pageCtrl.Create)
	pages.Put("/:id", pageCtrl.Update)
	pages.Delete("/:id", pageCtrl.Delete)

	// Posts
	posts := api.Group("/posts")
	postCtrl := app.Resolve[*PostController]()
	posts.Get("/", postCtrl.List)
	posts.Get("/:id", postCtrl.Get)
	posts.Post("/", postCtrl.Create)
	posts.Put("/:id", postCtrl.Update)
	posts.Delete("/:id", postCtrl.Delete)

	// Categories
	categories := api.Group("/categories")
	categoryCtrl := app.Resolve[*CategoryController]()
	categories.Get("/", categoryCtrl.List)
	categories.Get("/:id", categoryCtrl.Get)
	categories.Post("/", categoryCtrl.Create)
	categories.Put("/:id", categoryCtrl.Update)
	categories.Delete("/:id", categoryCtrl.Delete)

	// Media
	media := api.Group("/media")
	mediaCtrl := app.Resolve[*MediaController]()
	media.Get("/", mediaCtrl.List)
	media.Get("/:id", mediaCtrl.Get)
	media.Post("/upload", mediaCtrl.Upload)
	media.Delete("/:id", mediaCtrl.Delete)

	return nil
}

// Boot initializes the CMS module
func (m *CMSModule) Boot() error {
	return nil
}

// RegisterModels registers CMS models for migration
func (m *CMSModule) RegisterModels() []interface{} {
	return []interface{}{
		&Page{},
		&Post{},
		&Category{},
		&Tag{},
		&Media{},
		&Menu{},
		&MenuItem{},
	}
}
