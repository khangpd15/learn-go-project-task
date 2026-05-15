package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"task/api/internal/handler"
	"task/api/internal/middleware"
	"task/api/internal/repositories"
	"task/api/internal/routes"
	"task/api/internal/services"
)

func main() {

	// Không dùng gin.Default()
	// vì mình tự custom middleware
	r := gin.New()

	// Middleware global
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())

	// Task
	taskRepo := repositories.NewTaskRepository()

	taskService := services.NewTaskService(taskRepo)

	taskHandler := handler.NewTaskHandler(taskService)

	// User Repository
	userRepo := repositories.NewUserRepository()

	// Routes
	routes.SetupRoutes(
		r,
		taskHandler,
		userRepo,
	)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}