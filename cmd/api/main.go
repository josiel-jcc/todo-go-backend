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
	_ "todo-go-backend/docs"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	handlers.SetResponseConfig(cfg)

	userRepo := repositories.NewUserRepository()
	taskRepo := repositories.NewTaskRepository()
	tagRepo := repositories.NewTagRepository()
	commentRepo := repositories.NewCommentRepository()

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo)
	taskService := services.NewTaskService(taskRepo, userRepo, tagRepo)
	tagService := services.NewTagService(tagRepo)
	commentService := services.NewCommentService(commentRepo, taskRepo)

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

	authHandler := handlers.NewAuthHandler(authService, cfg)
	taskHandler := handlers.NewTaskHandler(taskService)
	tagHandler := handlers.NewTagHandler(tagService)
	commentHandler := handlers.NewCommentHandler(commentService)
	userHandler := handlers.NewUserHandler(notificationService, userRepo, userService)

	go notifications.StartScheduler(cfg, notificationService)

	router := gin.Default()
	router.Use(middleware.SecurityHeadersMiddleware(cfg))
	router.Use(middleware.CORSMiddleware(cfg))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if !cfg.IsProduction() {
		registerSwaggerRoutes(router)
	}

	api := router.Group("/api/v1")
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.AuthRateLimitMiddleware())
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/auth/logout", authHandler.Logout)

		protected.GET("/tasks", taskHandler.GetTasks)
		protected.GET("/tasks/assigned", taskHandler.GetAssignedTasks)
		protected.POST("/tasks", taskHandler.CreateTask)
		protected.GET("/tasks/:id/comments", commentHandler.GetComments)
		protected.GET("/tasks/:id", taskHandler.GetTask)
		protected.PUT("/tasks/:id", taskHandler.UpdateTask)
		protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
		protected.POST("/tasks/:id/share", taskHandler.ShareTask)
		protected.DELETE("/tasks/:id/share/:user_id", taskHandler.UnshareTask)

		protected.GET("/tags", tagHandler.GetTags)
		protected.GET("/tags/:id", tagHandler.GetTag)
		protected.POST("/tags", tagHandler.CreateTag)
		protected.PUT("/tags/:id", tagHandler.UpdateTag)
		protected.DELETE("/tags/:id", tagHandler.DeleteTag)

		protected.POST("/comments", commentHandler.CreateComment)
		protected.GET("/comments/:id", commentHandler.GetComment)
		protected.PUT("/comments/:id", commentHandler.UpdateComment)
		protected.DELETE("/comments/:id", commentHandler.DeleteComment)

		protected.GET("/users/me", userHandler.GetCurrentUser)
		protected.GET("/users/me/export", userHandler.ExportCurrentUser)
		protected.DELETE("/users/me", userHandler.DeleteCurrentUser)
		protected.GET("/users", userHandler.GetUsers)
		protected.PUT("/users/telegram-chat-id", userHandler.UpdateTelegramChatID)
		protected.PUT("/users/notifications-enabled", userHandler.UpdateNotificationsEnabled)

		if !cfg.IsProduction() {
			protected.POST("/notifications/test", userHandler.TestNotifications)
			protected.GET("/notifications/debug", userHandler.GetNotificationDebugInfo)
		}
	}

	log.Printf("Server starting on port %s (APP_ENV=%s)", cfg.Port, cfg.AppEnv)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func registerSwaggerRoutes(router *gin.Engine) {
	url := ginSwagger.URL("/openapi.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	router.GET("/openapi.json", func(c *gin.Context) {
		possiblePaths := []string{
			filepath.Join("docs", "openapi.json"),
			filepath.Join(".", "docs", "openapi.json"),
			filepath.Join("/root", "docs", "openapi.json"),
		}
		var data []byte
		var err error
		for _, path := range possiblePaths {
			data, err = os.ReadFile(path)
			if err == nil {
				break
			}
		}
		if err != nil {
			c.Redirect(302, "/swagger/doc.json")
			return
		}
		c.Data(200, "application/json", data)
	})

	router.GET("/swagger.json", func(c *gin.Context) {
		c.Redirect(302, "/openapi.json")
	})
}
