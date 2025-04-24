package main

import (
	"log"

	"go.uber.org/dig"

	"scripts-management/internal/config"
	"scripts-management/internal/core"
	"scripts-management/internal/handlers"
	"scripts-management/internal/repository"
	"scripts-management/internal/services"
	"scripts-management/pkg/database"
	"scripts-management/pkg/logger"
	"scripts-management/pkg/utils"
)

func main() {
	container := dig.New()

	// Register components
	container.Provide(config.NewConfig)
	container.Provide(logger.NewLogger)
	container.Provide(database.NewMongoClient)
	container.Provide(database.ProvideMongoDB)
	container.Provide(utils.NewJWTManager)

	// Register repositories
	container.Provide(repository.NewUserRepository)
	container.Provide(repository.NewScriptRepository)
	container.Provide(repository.NewScriptShareRepository)

	// Register services (order matters)
	container.Provide(services.NewAuthService)
	container.Provide(services.NewUserService)
	container.Provide(services.NewScriptService)

	// Register handlers
	container.Provide(handlers.NewAuthHandler)
	container.Provide(handlers.NewUserHandler)
	container.Provide(handlers.NewScriptHandler)

	// Register app
	container.Provide(core.NewApp)

	// Run the application
	err := container.Invoke(func(app *core.App) {
		if err := app.Start(); err != nil {
			log.Fatal(err)
		}
	})

	if err != nil {
		log.Fatal(err)
	}
}
