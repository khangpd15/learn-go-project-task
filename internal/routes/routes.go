package routes

import (
	"github.com/gin-gonic/gin"
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/realtime"
	"task_api/internal/repositories"
)

func SetupRoutes(
	r *gin.Engine,
	taskHandler *handler.TaskHandler,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	userRepo repositories.UserRepositoryInterface,
	projectHandler *handler.ProjectHandler,
	realtimeHandler *realtime.Handler,
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

	protected.GET("/ws", realtimeHandler.Connect)

	NewTaskRoutes(taskHandler).SetupTaskRoutes(protected)
	NewUserRoutes(userHandler).SetupUserRoutes(protected)
	NewProjectRoutes(projectHandler).SetupProjectRoutes(protected)
}
