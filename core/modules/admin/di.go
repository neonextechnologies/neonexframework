package admin

import (
	"neonexcore/internal/core"

	"gorm.io/gorm"
)

func RegisterDependencies(container *core.Container, db *gorm.DB) {
	// Register Repository as Singleton
	container.RegisterSingleton("admin.repository", func() interface{} {
		return NewRepository(db)
	})

	// Register Service as Singleton
	container.RegisterSingleton("admin.service", func() interface{} {
		repo := container.GetSingleton("admin.repository").(*Repository)
		return NewService(repo)
	})

	// Register Controller as Singleton
	container.RegisterSingleton("admin.controller", func() interface{} {
		service := container.GetSingleton("admin.service").(*Service)
		return NewController(service)
	})
}
