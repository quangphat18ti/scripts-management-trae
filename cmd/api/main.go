package main

import (
	"log"

	"go.uber.org/dig"

	"scripts-management/internal/config"
	"scripts-management/internal/core"
	"scripts-management/pkg/database"
	"scripts-management/pkg/logger"
)

func main() {
	container := dig.New()

	// Register components
	container.Provide(config.NewConfig)
	container.Provide(logger.NewLogger)
	container.Provide(database.NewMongoClient)
	container.Provide(database.ProvideMongoDB)
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
