package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/database"
	arts_module "github.com/muga20/artsMarket/modules/artwork-management/routes"
	"github.com/muga20/artsMarket/modules/notifications/services"
	user_module "github.com/muga20/artsMarket/modules/users/routes"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	logs_module "github.com/muga20/artsMarket/pkg/logs/routes"
	"github.com/muga20/artsMarket/pkg/middleware"
	"github.com/muga20/artsMarket/pkg/worker"

	//"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"

	_ "github.com/muga20/artsMarket/docs"
	swagger "github.com/swaggo/fiber-swagger"
)

// @title ArtsMarket API
// @version 1.0
// @description This is the API for the ArtsMarket system
// @host localhost:8080
// @BasePath /api/v1

func main() {
	initializeRedis()
	db := initializeDatabase()
	responseHandler := handlers.NewResponseHandler(db)

	// Initialize the notification service
	notificationService := services.NewNotificationService(responseHandler)

	// Create and start the worker
	worker := worker.NewNotificationWorker(notificationService, responseHandler, db)
	go worker.Start()

	// Start the Fiber app
	app := fiber.New()
	setupRoutes(app, db, responseHandler)
	startServer(app)
}

func initializeRedis() {
	// Initialize Redis connection
	config.InitializeRedis()

	// Initialize Redis client for background tasks (used in email sending, etc.)
	redisClient := asynq.NewClient(*config.RedisConfig)
	defer redisClient.Close()
}

func initializeDatabase() *gorm.DB {
	// Connect to the database
	db, err := database.ConnectToDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migration
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	return db
}

func configureMiddleware(app *fiber.App) {
	// Configure CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080, https://yourfrontenddomain.com",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Apply rate-limiting middleware
	app.Use(middleware.RateLimitMiddleware())
}

func setupRoutes(app *fiber.App, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	// Swagger Route for API documentation
	app.Get("/swagger/*", swagger.WrapHandler)

	// Initialize S3 client
	cld, err := config.NewCloudinaryClient()
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary client: %v", err)
	}
	// Initialize the API routes
	apiV1 := app.Group("/api/v1")
	user_module.UserModuleSetupRoutes(apiV1, db, responseHandler, cld)
	logs_module.LogsModuleSetupRoutes(apiV1, db, responseHandler)
	arts_module.ArtsManagementSetupRoutes(apiV1, db, cld, responseHandler)
}

func startServer(app *fiber.App) {
	// Set the server port from the config or default to 8080
	port := config.Envs.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	// Start the Fiber app
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
