package handlers

import (
	"fmt"
	"os"
	"time"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/middleware"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB cria um banco de dados para testes
// Tenta usar MySQL se as variáveis de ambiente estiverem configuradas (CI),
// caso contrário tenta usar SQLite (requer CGO habilitado)
func setupTestDB() *gorm.DB {
	var db *gorm.DB
	var err error

	// Verificar se MySQL está disponível (como na pipeline CI)
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")

	if dbHost != "" && dbPort != "" && dbUser != "" && dbPassword != "" && dbName != "" {
		// Usar MySQL (como na pipeline CI)
		// Adicionar parâmetros para melhorar robustez da conexão
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=10s&writeTimeout=10s",
			dbUser,
			dbPassword,
			dbHost,
			dbPort,
			dbName,
		)
		
		// Tentar conectar com retry
		var lastErr error
		for i := 0; i < 5; i++ {
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err == nil {
				break
			}
			lastErr = err
			if i < 4 {
				// Aguardar antes de tentar novamente (exponencial backoff)
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
		
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to MySQL test database after 5 attempts: %v", lastErr))
		}
	} else {
		// Tentar usar SQLite (requer CGO habilitado)
		// Usar arquivo temporário ao invés de :memory: para compatibilidade
		tmpFile, err := os.CreateTemp("", "test_*.db")
		if err != nil {
			panic("Failed to create temp file for test database: " + err.Error())
		}
		tmpFile.Close()
		
		// Remover o arquivo após os testes (será recriado pelo SQLite)
		os.Remove(tmpFile.Name())

		db, err = gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
		if err != nil {
			panic("Failed to connect to SQLite test database. SQLite requires CGO to be enabled. " +
				"Either enable CGO (set CGO_ENABLED=1) or configure MySQL environment variables " +
				"(DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME). " +
				"Error: " + err.Error())
		}
	}

	err = db.AutoMigrate(&models.User{}, &models.Task{}, &models.Tag{}, &models.Comment{}, &models.Notification{}, &models.TokenDenylist{})
	if err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	// Limpar dados existentes para garantir testes isolados
	// Isso é especialmente importante quando usando MySQL compartilhado na CI
	// Verificar se é MySQL ou SQLite
	if dbHost != "" {
		// MySQL - desabilitar foreign keys temporariamente
		db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		db.Exec("TRUNCATE TABLE notifications")
		db.Exec("TRUNCATE TABLE comments")
		db.Exec("TRUNCATE TABLE task_tags")
		db.Exec("TRUNCATE TABLE tasks")
		db.Exec("TRUNCATE TABLE tags")
		db.Exec("TRUNCATE TABLE users")
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	} else {
		// SQLite - usar DELETE (TRUNCATE não funciona em SQLite)
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM comments")
		db.Exec("DELETE FROM task_tags")
		db.Exec("DELETE FROM tasks")
		db.Exec("DELETE FROM tags")
		db.Exec("DELETE FROM users")
	}

	database.DB = db
	return db
}

// setupTestRouter cria um router de teste com handlers configurados
func setupTestRouter(jwtSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize repositories
	userRepo := repositories.NewUserRepository()
	taskRepo := repositories.NewTaskRepository()

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtSecret)
	tagRepo := repositories.NewTagRepository()
	taskService := services.NewTaskService(taskRepo, userRepo, tagRepo)

	// Initialize handlers
	cfg := &config.Config{AppEnv: "development", JWTSecret: jwtSecret}
	authHandler := NewAuthHandler(authService, cfg)
	taskHandler := NewTaskHandler(taskService)

	// Public routes
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		protected.GET("/tasks", taskHandler.GetTasks)
		protected.GET("/tasks/assigned", taskHandler.GetAssignedTasks)
		protected.GET("/tasks/:id", taskHandler.GetTask)
		protected.POST("/tasks", taskHandler.CreateTask)
		protected.PUT("/tasks/:id", taskHandler.UpdateTask)
		protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
	}

	return router
}
