package core

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"scripts-management/internal/config"
	"scripts-management/internal/handlers"
	"scripts-management/internal/middleware"
	"scripts-management/internal/models"
	"scripts-management/internal/repository"
	"scripts-management/internal/services"
	"scripts-management/pkg/utils"
)

type App struct {
	config         *config.Config
	logger         *zap.Logger
	fiber          *fiber.App
	authHandler    *handlers.AuthHandler
	userHandler    *handlers.UserHandler
	userService    *services.UserService
	scriptHandler  *handlers.ScriptHandler
	processHandler *handlers.ProcessHandler
	jwtManager     *utils.JWTManager
}

func NewApp(config *config.Config, logger *zap.Logger, db *mongo.Database) *App {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	scriptRepo := repository.NewScriptRepository(db)
	scriptShareRepo := repository.NewScriptShareRepository(db)
	processRepo := repository.NewProcessRepository(db)

	// Initialize JWT manager
	jwtManager, err := utils.NewJWTManager()
	if err != nil {
		logger.Fatal("Failed to initialize JWT manager", zap.Error(err))
	}

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtManager)
	userService, err := services.NewUserService(userRepo, config, authService)
	if err != nil {
		logger.Fatal("Failed to initialize user service", zap.Error(err))
	}
	scriptService := services.NewScriptService(scriptRepo, scriptShareRepo, userRepo)
	processService := services.NewProcessService(processRepo, scriptRepo, scriptService, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	scriptHandler := handlers.NewScriptHandler(scriptService)
	processHandler := handlers.NewProcessHandler(processService)

	app := &App{
		config:         config,
		logger:         logger,
		fiber:          fiber.New(),
		authHandler:    authHandler,
		userHandler:    userHandler,
		userService:    userService,
		scriptHandler:  scriptHandler,
		processHandler: processHandler,
		jwtManager:     jwtManager,
	}

	// setup logger for apps
	app.fiber.Use(func(c *fiber.Ctx) error {
		logger.Info("Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("user-agent", c.Get("User-Agent")),
		)
		return c.Next()
	})

	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	// Health check route
	a.fiber.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Auth routes
	auth := a.fiber.Group("/auth")
	auth.Post("/login", a.authHandler.Login)
	auth.Post("/signup", a.authHandler.Signup)

	// Initialize root account
	if err := a.userService.InitRootAccount(context.Background()); err != nil {
		a.logger.Fatal("Failed to initialize root account", zap.Error(err))
	}

	// User management routes
	api := a.fiber.Group("/api", middleware.AuthMiddleware(a.jwtManager))

	// User management (Root and Admin only)
	users := api.Group("/users")
	users.Post("/", middleware.RoleAuth(models.RoleRoot, models.RoleAdmin), a.userHandler.CreateUser)
	users.Delete("/:id", middleware.RoleAuth(models.RoleRoot, models.RoleAdmin), a.userHandler.DeleteUser)
	users.Put("/:id/password", middleware.RoleAuth(models.RoleRoot, models.RoleAdmin), a.userHandler.ChangePassword)

	// Script management routes
	scripts := api.Group("/scripts")
	scripts.Post("/", a.scriptHandler.CreateScript)
	scripts.Get("/", a.scriptHandler.GetUserScripts)
	scripts.Get("/:id", a.scriptHandler.GetScript)
	scripts.Put("/:id", a.scriptHandler.UpdateScript)
	scripts.Delete("/:id", a.scriptHandler.DeleteScript)
	scripts.Post("/:id/share", a.scriptHandler.ShareScript)
	scripts.Delete("/:id/share/:userId", a.scriptHandler.RevokeShare)

	// Process management routes
	scripts.Post("/:id/run", a.processHandler.RunScript)

	processes := api.Group("/processes")
	processes.Get("/", a.processHandler.GetProcesses)
	processes.Post("/:id/stop", a.processHandler.StopProcess)
}

func (a *App) Start() error {
	return a.fiber.Listen(fmt.Sprintf(":%s", a.config.AppPort))
}
