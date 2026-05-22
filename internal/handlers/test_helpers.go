package handlers

import (
	"fmt"
	"os"
	"sync"
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

var (
	testDBOnce   sync.Once
	testDBShared *gorm.DB
)

func openMySQLTestDB(host, port, user, password, name string) *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=10s&writeTimeout=10s",
		user, password, host, port, name,
	)
	var db *gorm.DB
	var err error
	var lastErr error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			return db
		}
		lastErr = err
		if i < 4 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	panic(fmt.Sprintf("Failed to connect to MySQL test database after 5 attempts: %v", lastErr))
}

func migrateTestSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.TaskSharedWith{},
		&models.Tag{},
		&models.Comment{},
		&models.Notification{},
		&models.TokenDenylist{},
		&models.Group{},
		&models.GroupMember{},
		&models.GroupInvitation{},
		&models.UserNotification{},
	)
}

func truncateTestData(db *gorm.DB, useMySQL bool) {
	if useMySQL {
		db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		db.Exec("TRUNCATE TABLE user_notifications")
		db.Exec("TRUNCATE TABLE group_invitations")
		db.Exec("TRUNCATE TABLE group_members")
		db.Exec("TRUNCATE TABLE `groups`")
		db.Exec("TRUNCATE TABLE task_shared_with")
		db.Exec("TRUNCATE TABLE notifications")
		db.Exec("TRUNCATE TABLE comments")
		db.Exec("TRUNCATE TABLE task_tags")
		db.Exec("TRUNCATE TABLE tasks")
		db.Exec("TRUNCATE TABLE tags")
		db.Exec("TRUNCATE TABLE users")
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		return
	}
	db.Exec("DELETE FROM user_notifications")
	db.Exec("DELETE FROM group_invitations")
	db.Exec("DELETE FROM group_members")
	db.Exec("DELETE FROM groups")
	db.Exec("DELETE FROM task_shared_with")
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM task_tags")
	db.Exec("DELETE FROM tasks")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM users")
}

// setupTestDB cria um banco de dados para testes
// Tenta usar MySQL se as variáveis de ambiente estiverem configuradas (CI),
// caso contrário tenta usar SQLite (requer CGO habilitado)
func setupTestDB() *gorm.DB {
	// Verificar se MySQL está disponível (como na pipeline CI)
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")

	useMySQL := dbHost != "" && dbPort != "" && dbUser != "" && dbPassword != "" && dbName != ""

	if useMySQL {
		testDBOnce.Do(func() {
			testDBShared = openMySQLTestDB(dbHost, dbPort, dbUser, dbPassword, dbName)
			if err := migrateTestSchema(testDBShared); err != nil {
				panic("Failed to migrate test database: " + err.Error())
			}
		})
		truncateTestData(testDBShared, true)
		database.DB = testDBShared
		return testDBShared
	}

	tmpFile, tmpErr := os.CreateTemp("", "test_*.db")
	if tmpErr != nil {
		panic("Failed to create temp file for test database: " + tmpErr.Error())
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	db, err := gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to SQLite test database. SQLite requires CGO to be enabled. " +
			"Either enable CGO (set CGO_ENABLED=1) or configure MySQL environment variables " +
			"(DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME). " +
			"Error: " + err.Error())
	}
	if err := migrateTestSchema(db); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}
	truncateTestData(db, false)
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
	groupRepo := repositories.NewGroupRepository()
	groupInvitationRepo := repositories.NewGroupInvitationRepository()
	userNotificationRepo := repositories.NewUserNotificationRepository()

	groupService := services.NewGroupService(groupRepo, groupInvitationRepo, userNotificationRepo, userRepo)

	// Initialize services
	authService := services.NewAuthService(userRepo, jwtSecret, groupService)
	tagRepo := repositories.NewTagRepository()
	taskService := services.NewTaskService(taskRepo, userRepo, tagRepo, groupService)

	userService := services.NewUserService(userRepo, groupRepo, groupInvitationRepo, userNotificationRepo)
	userNotificationService := services.NewUserNotificationService(userNotificationRepo)

	// Initialize handlers
	cfg := &config.Config{AppEnv: "development", JWTSecret: jwtSecret}
	authHandler := NewAuthHandler(authService, cfg)
	taskHandler := NewTaskHandler(taskService)
	userHandler := NewUserHandler(nil, userRepo, userService, groupService)
	groupHandler := NewGroupHandler(groupService)
	groupInvitationHandler := NewGroupInvitationHandler(groupService)
	userNotificationHandler := NewUserNotificationHandler(userNotificationService)

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
		protected.POST("/tasks/:id/share", taskHandler.ShareTask)

		protected.GET("/users", userHandler.GetUsers)

		protected.GET("/groups", groupHandler.ListGroups)
		protected.POST("/groups", groupHandler.CreateGroup)
		protected.GET("/groups/:id", groupHandler.GetGroup)
		protected.PUT("/groups/:id", groupHandler.UpdateGroup)
		protected.DELETE("/groups/:id", groupHandler.DeleteGroup)
		protected.POST("/groups/:id/invitations", groupHandler.InviteUser)
		protected.DELETE("/groups/:id/invitations/:invitation_id", groupHandler.CancelInvitation)
		protected.DELETE("/groups/:id/members/:user_id", groupHandler.RemoveMember)

		protected.GET("/group-invitations", groupInvitationHandler.ListReceived)
		protected.POST("/group-invitations/:id/accept", groupInvitationHandler.Accept)
		protected.POST("/group-invitations/:id/decline", groupInvitationHandler.Decline)

		protected.GET("/notifications/in-app", userNotificationHandler.List)
		protected.GET("/notifications/in-app/unread-count", userNotificationHandler.UnreadCount)
	}

	return router
}
