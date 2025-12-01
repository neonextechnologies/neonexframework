package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type Module interface {
	Name() string
	Init()
	Routes(app *fiber.App, c *Container)
	RegisterServices(c *Container)
}

type ModuleRegistry struct {
	Modules []Module
}

func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		Modules: make([]Module, 0),
	}
}

func (r *ModuleRegistry) Register(m Module) {
	fmt.Println("Register module:", m.Name())
	r.Modules = append(r.Modules, m)
}

func (r *ModuleRegistry) Load() {
	for _, m := range r.Modules {
		fmt.Println("Init module:", m.Name())
		m.Init()
	}
}

func (r *ModuleRegistry) LoadRoutes(app fiber.Router, c *Container) {
	for _, m := range r.Modules {
		m.Routes(app, c)
	}
}

func (r *ModuleRegistry) AutoDiscover() {
	entries, err := os.ReadDir("./modules")
	if err != nil {
		fmt.Println("Cannot read modules folder:", err)
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		moduleFolder := e.Name()
		metaFile := filepath.Join("modules", moduleFolder, "module.json")

		raw, err := os.ReadFile(metaFile)
		if err != nil {
			fmt.Printf("No metadata for module '%s', skipping...\n", moduleFolder)
			continue
		}

		var meta struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}

		json.Unmarshal(raw, &meta)

		if !meta.Enabled {
			fmt.Println("⏹️  Module disabled:", meta.Name)
			continue
		}

		factory, ok := ModuleMap[meta.Name]
		if !ok {
			fmt.Println("No factory found for:", meta.Name)
			continue
		}

		r.Register(factory())
	}
}

func (r *ModuleRegistry) RegisterModuleServices(container *Container) {
	for _, m := range r.Modules {
		m.RegisterServices(container)
	}
}
