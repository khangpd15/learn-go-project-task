package app

import (
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"task_api/internal/config"
	"task_api/internal/database"
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/repositories"
	"task_api/internal/cache"
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

	// Services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo)
	projectService := services.NewProjectService(projectRepo)

	// Create cache abstraction and pass to TaskService
	cacheClient := cache.NewRedisCache(redisClient)
	taskService := services.NewTaskService(taskRepo, projectRepo, userRepo, cacheClient)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Routes
	routes.SetupRoutes(
		r,
		taskHandler,
		userHandler,
		authHandler,
		userRepo,
		projectHandler,
	)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
	
}