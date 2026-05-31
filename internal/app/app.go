package app

import (
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"task_api/internal/cache"
	"task_api/internal/config"
	"task_api/internal/database"
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/queue"
	"task_api/internal/realtime"
	"task_api/internal/repositories"
	"task_api/internal/routes"
	"task_api/internal/services"
)

func Run() {
	// Connect PostgreSQL
	db := database.ConnectPostgres()

	// Connect Redis
	redisCfg := config.NewRedisConfig()
	redisClient := database.ConnectRedis(redisCfg)
	defer redisClient.Close()
	redisQueue := queue.NewRedisQueue(redisClient)
	// Gin engine
	r := gin.New()

	// Global middleware
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	hub := realtime.NewHub()
	go hub.Run()
	eventPublisher := realtime.NewPublisher(hub)
	// Services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo)
	projectService := services.NewProjectService(projectRepo)
    notificationService := services.NewNotificationService(notificationRepo)
	// Create cache abstraction and pass to TaskService
	cacheClient := cache.NewRedisCache(redisClient)
	
	taskService := services.NewTaskService(taskRepo, projectRepo, userRepo, cacheClient, redisQueue, eventPublisher, notificationService)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	realtimeHandler := realtime.NewHandler(hub)

	// Routes
	routes.SetupRoutes(
	r,
	taskHandler,
	userHandler,
	authHandler,
	userRepo,
	projectHandler,
	realtimeHandler,
)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}

}
