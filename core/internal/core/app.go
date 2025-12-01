package core

import (
	"fmt"
	"time"

	"neonexcore/internal/config"
	"neonexcore/pkg/api"
	"neonexcore/pkg/database"
	"neonexcore/pkg/logger"
	"neonexcore/pkg/metrics"
	"neonexcore/pkg/websocket"

	"github.com/gofiber/fiber/v2"
)

// -----------------------------------------------------------
// 1) App Struct
// -----------------------------------------------------------
type App struct {
	Registry   *ModuleRegistry
	Container  *Container
	Migrator   *database.Migrator
	Logger     logger.Logger
	WSHub      *websocket.Hub // WebSocket hub
	Collector  *metrics.Collector
	Dashboard  *metrics.Dashboard
}

// -----------------------------------------------------------
// 2) NewApp() - สร้าง App + โหลด ModuleRegistry
// -----------------------------------------------------------
func NewApp() *App {
	// Initialize WebSocket hub
	hubConfig := websocket.DefaultHubConfig()
	wsHub := websocket.NewHub(hubConfig)
	
	// Initialize metrics collector
	collectorConfig := metrics.DefaultCollectorConfig()
	collectorConfig.CollectSystemMetrics = true
	collectorConfig.SystemMetricsInterval = 5 * time.Second
	collector := metrics.NewCollector(collectorConfig)
	
	// Initialize dashboard
	dashConfig := metrics.DefaultDashboardConfig()
	dashConfig.BroadcastInterval = 1 * time.Second
	dashboard := metrics.NewDashboard(collector, wsHub, dashConfig)
	
	return &App{
		Registry:  NewModuleRegistry(),
		Container: NewContainer(),
		Logger:    logger.NewLogger(),
		WSHub:     wsHub,
		Collector: collector,
		Dashboard: dashboard,
	}
}

// -----------------------------------------------------------
// 3) InitLogger() - Initialize Logger
// -----------------------------------------------------------
func (a *App) InitLogger(cfg logger.Config) error {
	if err := logger.Setup(cfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.Logger = logger.NewLogger()
	a.Logger.Info("Logger initialized", logger.Fields{
		"level":  cfg.Level,
		"format": cfg.Format,
		"output": cfg.Output,
	})
	return nil
}

// -----------------------------------------------------------
// 4) InitDatabase() - เริ่ม Database + Migrator
// -----------------------------------------------------------
func (a *App) InitDatabase() error {
	dbConfig := config.LoadDatabaseConfig()
	_, err := config.InitDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize migrator
	a.Migrator = database.NewMigrator(config.DB.GetDB())
	a.Logger.Info("Database initialized", logger.Fields{"driver": dbConfig.Driver})

	return nil
}

// -----------------------------------------------------------
// 5) RegisterModels() - Register models for auto-migration
// -----------------------------------------------------------
func (a *App) RegisterModels(models ...interface{}) {
	if a.Migrator != nil {
		a.Migrator.RegisterModels(models...)
		a.Logger.Info("Models registered for migration", logger.Fields{"count": len(models)})
	}
}

// -----------------------------------------------------------
// 6) AutoMigrate() - Run auto-migration
// -----------------------------------------------------------
func (a *App) AutoMigrate() error {
	if a.Migrator != nil {
		a.Logger.Info("Running auto-migration...")
		if err := a.Migrator.AutoMigrate(); err != nil {
			a.Logger.Error("Auto-migration failed", logger.Fields{"error": err.Error()})
			return err
		}
		a.Logger.Info("Auto-migration completed")
	}
	return nil
}

// -----------------------------------------------------------
// 7) Boot() - เริ่มระบบพื้นฐาน
// -----------------------------------------------------------
func (a *App) Boot() {
	fmt.Println("⚙️  Booting Neonex Core...")
	a.Logger.Info("Neonex Core booting...")
}

// -----------------------------------------------------------
// 8) StartHTTP() - HTTP Server Engine
// -----------------------------------------------------------
func (a *App) StartHTTP() {
	// Configure Fiber with custom branding
	app := fiber.New(fiber.Config{
		AppName:               "Neonex Core v0.1-alpha",
		DisableStartupMessage: true, // Disable default Fiber banner
	})

	// Global middleware - CORS
	app.Use(api.CORSMiddleware())

	// Global middleware - Security headers
	app.Use(api.SecurityHeadersMiddleware())

	// Global middleware - Request ID
	app.Use(api.RequestIDMiddleware())

	// Global middleware - Logger
	app.Use(logger.RequestIDMiddleware(a.Logger))
	app.Use(logger.HTTPMiddleware(a.Logger))

	// Global middleware - Metrics
	app.Use(metrics.Middleware(a.Collector))
	app.Use(metrics.MethodMiddleware(a.Collector))
	app.Use(metrics.ErrorMiddleware(a.Collector))

	// Global rate limiting (100 requests per minute per IP)
	app.Use(api.IPRateLimitMiddleware(100, time.Minute))

	// Health check routes
	healthChecker := api.NewHealthChecker("0.1-alpha", config.DB.GetDB())
	api.SetupHealthRoutes(app, healthChecker, config.DB.GetDB())

	// API versioning
	versionManager := api.NewVersionManager()
	versionManager.RegisterVersion("v1", "1.0.0")

	// Setup Swagger documentation
	swagger := api.CreateDefaultSwagger()
	swagger.Info.Title = "Neonex Core API"
	swagger.Info.Description = "Neonex Core - Modular Backend Framework with Authentication, RBAC, and Module System"
	swagger.Info.Version = "0.1-alpha"
	api.SetupSwaggerRoutes(app, swagger)

	// Create versioned API routes
	apiV1 := api.VersionedRouter(app, "v1")
	apiV1.Use(api.VersionMiddleware(versionManager))

	// Load module routes
	a.Logger.Info("Registering modules...")
	a.Registry.RegisterModuleServices(a.Container)
	a.Registry.LoadRoutes(apiV1, a.Container) // Load routes into /api/v1

	// Setup WebSocket routes
	a.Logger.Info("Setting up WebSocket support...")
	websocket.SetupRoutes(app, a.WSHub, nil) // nil = use default message handler

	// Setup metrics dashboard
	a.Logger.Info("Setting up metrics dashboard...")
	a.Dashboard.SetupRoutes(app)

	// Default homepage
	app.Get("/", func(c *fiber.Ctx) error {
		return api.Success(c, fiber.Map{
			"framework": "Neonex Core",
			"version":   "0.1-alpha",
			"status":    "running",
			"engine":    "Fiber (fasthttp)",
			"endpoints": fiber.Map{
				"health":        "/health",
				"documentation": "/api/docs",
				"api":           "/api/v1",
			},
		})
	})

	// Custom Neonex startup banner
	fmt.Println()
	fmt.Println("┌───────────────────────────────────────────────────┐")
	fmt.Println("│              Neonex Core v0.1-alpha               │")
	fmt.Println("│               http://127.0.0.1:8080               │")
	fmt.Println("│       (bound on host 0.0.0.0 and port 8080)       │")
	fmt.Println("│                                                   │")
	fmt.Println("│ Framework .... Neonex  Engine ..... Fiber/fasthttp│")
	fmt.Println("│                                                   │")
	fmt.Println("│ 📚 Documentation: http://127.0.0.1:8080/api/docs  │")
	fmt.Println("│ 💚 Health Check:  http://127.0.0.1:8080/health    │")
	fmt.Println("│ 🚀 API v1:        http://127.0.0.1:8080/api/v1    │")
	fmt.Println("│ 🔴 WebSocket:     ws://127.0.0.1:8080/ws          │")
	fmt.Println("│ 📊 Metrics:       http://127.0.0.1:8080/metrics/dashboard │")
	fmt.Println("└───────────────────────────────────────────────────┘")
	fmt.Println()

	a.Logger.Info("HTTP server starting", logger.Fields{"port": 8080})
	if err := app.Listen(":8080"); err != nil {
		a.Logger.Fatal("Failed to start server", logger.Fields{"error": err.Error()})
	}
}
