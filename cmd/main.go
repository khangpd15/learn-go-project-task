package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"task/api/internal/handler"
	"task/api/internal/repositories"
	"task/api/internal/routes"
	"task/api/internal/services"
)

func main() {

	r := gin.Default()

	taskRepo := repositories.NewTaskRepository()

	taskService := services.NewTaskService(taskRepo)

	taskHandler := handler.NewTaskHandler(taskService)

	routes.SetupRoutes(r, taskHandler)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
