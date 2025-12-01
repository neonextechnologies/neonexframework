package ecommerce

import (
	"neonexframework/pkg/app"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// EcommerceModule handles e-commerce functionality
type EcommerceModule struct {
	db *gorm.DB
}

// New creates a new e-commerce module
func New() *EcommerceModule {
	return &EcommerceModule{}
}

// Name returns the module name
func (m *EcommerceModule) Name() string {
	return "ecommerce"
}

// RegisterServices registers e-commerce services
func (m *EcommerceModule) RegisterServices(c *app.Container) error {
	// Register repositories
	c.Provide(NewProductRepository)
	c.Provide(NewCartRepository)
	c.Provide(NewOrderRepository)

	// Register services
	c.Provide(NewProductService)
	c.Provide(NewCartService)
	c.Provide(NewOrderService)
	c.Provide(NewPaymentService)

	// Register controllers
	c.Provide(NewProductController)
	c.Provide(NewCartController)
	c.Provide(NewOrderController)

	return nil
}

// RegisterRoutes registers e-commerce routes
func (m *EcommerceModule) RegisterRoutes(router fiber.Router) error {
	// API routes
	api := router.Group("/api/v1/ecommerce")

	// Products
	products := api.Group("/products")
	productCtrl := app.Resolve[*ProductController]()
	products.Get("/", productCtrl.List)
	products.Get("/:id", productCtrl.Get)
	products.Post("/", productCtrl.Create)
	products.Put("/:id", productCtrl.Update)
	products.Delete("/:id", productCtrl.Delete)

	// Cart
	cart := api.Group("/cart")
	cartCtrl := app.Resolve[*CartController]()
	cart.Get("/", cartCtrl.Get)
	cart.Post("/items", cartCtrl.AddItem)
	cart.Put("/items/:id", cartCtrl.UpdateItem)
	cart.Delete("/items/:id", cartCtrl.RemoveItem)
	cart.Delete("/", cartCtrl.Clear)

	// Orders
	orders := api.Group("/orders")
	orderCtrl := app.Resolve[*OrderController]()
	orders.Get("/", orderCtrl.List)
	orders.Get("/:id", orderCtrl.Get)
	orders.Post("/", orderCtrl.Create)
	orders.Put("/:id/status", orderCtrl.UpdateStatus)

	return nil
}

// Boot initializes the e-commerce module
func (m *EcommerceModule) Boot() error {
	return nil
}

// RegisterModels registers e-commerce models for migration
func (m *EcommerceModule) RegisterModels() []interface{} {
	return []interface{}{
		&Product{},
		&ProductCategory{},
		&Cart{},
		&Order{},
		&Payment{},
		&Coupon{},
		&Review{},
	}
}
