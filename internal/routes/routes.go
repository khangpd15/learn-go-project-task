package routes

import (
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/repositories"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	taskHandler *handler.TaskHandler,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	userRepo repositories.UserRepositoryInterface,
	projectHandler *handler.ProjectHandler,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server is running",
		})
	})

	v1 := r.Group("/api/v1")

	// Public routes: không cần token
	NewAuthRoutes(authHandler).SetupAuthRoutes(v1)

	// Protected routes: cần token
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(userRepo))

	NewTaskRoutes(taskHandler).SetupTaskRoutes(protected)
	NewUserRoutes(userHandler).SetupUserRoutes(protected)
	NewProjectRoutes(projectHandler).SetupProjectRoutes(protected)
}