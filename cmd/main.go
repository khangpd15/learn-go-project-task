package main

import (
	"log"

	"github.com/gin-gonic/gin"
    "task_api/internal/database"
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/repositories"
	"task_api/internal/routes"
	"task_api/internal/services"
)

func main() {
    // Connect PostgreSQL
	db := database.ConnectPostgres()
	// Không dùng gin.Default()
	// vì mình tự custom middleware
	r := gin.New()

	// Middleware global
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())

	// Task
	taskRepo := repositories.NewTaskRepository(db)

	taskService := services.NewTaskService(taskRepo)

	taskHandler := handler.NewTaskHandler(taskService)

	// User
	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Routes
	routes.SetupRoutes(
		r,
		taskHandler,
		userRepo,
		userHandler,
	)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
