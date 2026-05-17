// @title           Todo API
// @version         1.0
// @description     A RESTful API for managing tasks with JWT authentication and user task assignment
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"os"
	"path/filepath"
	_ "todo-go-backend/docs" // Swagger documentation
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/handlers"
	"todo-go-backend/internal/middleware"
	"todo-go-backend/internal/notifications"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository()
	taskRepo := repositories.NewTaskRepository()
	tagRepo := repositories.NewTagRepository()
	commentRepo := repositories.NewCommentRepository()

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	taskService := services.NewTaskService(taskRepo, userRepo, tagRepo)
	tagService := services.NewTagService(tagRepo)
	commentService := services.NewCommentService(commentRepo, taskRepo)

	// Initialize notification services
	emailService := notifications.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)
	telegramService := notifications.NewTelegramService(cfg.TelegramBotToken)
	notificationRepo := repositories.NewNotificationRepository()
	notificationService := notifications.NewNotificationService(
		emailService,
		telegramService,
		notificationRepo,
		taskRepo,
		userRepo,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskService)
	tagHandler := handlers.NewTagHandler(tagService)
	commentHandler := handlers.NewCommentHandler(commentService)
	userHandler := handlers.NewUserHandler(notificationService, userRepo)

	// Start notification scheduler
	go notifications.StartScheduler(cfg, notificationService)

	// Setup router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Health check endpoint
	// @Summary     Health check endpoint
	// @Description Returns the health status of the API
	// @Tags        health
	// @Accept      json
	// @Produce     json
	// @Success     200  {object}  map[string]string
	// @Router      /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation - configure to use openapi.json
	url := ginSwagger.URL("/openapi.json") // Point to OpenAPI 3.0 JSON
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Serve openapi.json (OpenAPI 3.0) and swagger.json (fallback) directly for openapi-typescript
	// Using different paths to avoid conflicts with /swagger/*any
	router.GET("/openapi.json", func(c *gin.Context) {
		// Try different possible paths (local dev and Docker)
		possiblePaths := []string{
			filepath.Join("docs", "openapi.json"),
			filepath.Join(".", "docs", "openapi.json"),
			filepath.Join("/root", "docs", "openapi.json"),
		}

		var data []byte
		var err error
		var usedPath string
		for _, path := range possiblePaths {
			data, err = os.ReadFile(path)
			if err == nil {
				usedPath = path
				break
			}
		}

		if err != nil {
			log.Printf("Failed to read openapi.json from all paths, redirecting to /swagger/doc.json")
			c.Redirect(302, "/swagger/doc.json")
			return
		}

		log.Printf("Serving openapi.json from: %s", usedPath)
		c.Data(200, "application/json", data)
	})

	// Keep swagger.json endpoint for backward compatibility (redirects to openapi.json)
	router.GET("/swagger.json", func(c *gin.Context) {
		c.Redirect(302, "/openapi.json")
	})

	router.GET("/swagger.yaml", func(c *gin.Context) {
		// Try different possible paths (local dev and Docker)
		possiblePaths := []string{
			filepath.Join("docs", "swagger.yaml"),
			filepath.Join(".", "docs", "swagger.yaml"),
			filepath.Join("/root", "docs", "swagger.yaml"),
		}

		var data []byte
		var err error
		var usedPath string
		for _, path := range possiblePaths {
			data, err = os.ReadFile(path)
			if err == nil {
				usedPath = path
				break
			}
		}

		if err != nil {
			log.Printf("Failed to read swagger.yaml from all paths. Tried: %v", possiblePaths)
			c.JSON(404, gin.H{"error": "swagger.yaml not found. The file may not be available in this environment. Try using /swagger.json or /swagger/doc.json instead."})
			return
		}

		log.Printf("Serving swagger.yaml from: %s", usedPath)
		c.Data(200, "application/x-yaml", data)
	})

	// Public routes
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Tasks routes
		protected.GET("/tasks", taskHandler.GetTasks)
		protected.GET("/tasks/assigned", taskHandler.GetAssignedTasks)
		protected.POST("/tasks", taskHandler.CreateTask)

		// Comments routes for tasks (must be before /tasks/:id to avoid route conflict)
		// Using /tasks/:id/comments with same parameter name to avoid Gin route conflict
		protected.GET("/tasks/:id/comments", commentHandler.GetComments)

		// Tasks routes with ID (must be after /tasks/:id/comments)
		protected.GET("/tasks/:id", taskHandler.GetTask)
		protected.PUT("/tasks/:id", taskHandler.UpdateTask)
		protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
		protected.POST("/tasks/:id/share", taskHandler.ShareTask)
		protected.DELETE("/tasks/:id/share/:user_id", taskHandler.UnshareTask)

		// Tags routes
		protected.GET("/tags", tagHandler.GetTags)
		protected.GET("/tags/:id", tagHandler.GetTag)
		protected.POST("/tags", tagHandler.CreateTag)
		protected.PUT("/tags/:id", tagHandler.UpdateTag)
		protected.DELETE("/tags/:id", tagHandler.DeleteTag)

		// Comments routes
		protected.POST("/comments", commentHandler.CreateComment)
		protected.GET("/comments/:id", commentHandler.GetComment)
		protected.PUT("/comments/:id", commentHandler.UpdateComment)
		protected.DELETE("/comments/:id", commentHandler.DeleteComment)

		// User routes
		protected.GET("/users/me", userHandler.GetCurrentUser)
		protected.GET("/users", userHandler.GetUsers)
		protected.PUT("/users/telegram-chat-id", userHandler.UpdateTelegramChatID)
		protected.PUT("/users/notifications-enabled", userHandler.UpdateNotificationsEnabled)

		// Notification test routes (for testing)
		protected.POST("/notifications/test", userHandler.TestNotifications)
		protected.GET("/notifications/debug", userHandler.GetNotificationDebugInfo)
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
