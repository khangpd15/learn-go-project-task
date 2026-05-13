package routes

import (
"github.com/gin-gonic/gin"
"task/api/internal/handler"
)
func SetupRoutes(
	r *gin.Engine,
	taskHandler *handler.TaskHandler,
) {

	v1 := r.Group("/api/v1")

	{
		v1.GET("/tasks", taskHandler.GetAllTasks)

		v1.GET("/tasks/:id", taskHandler.GetTaskById)

		// v1.POST("/tasks", taskHandler.CreateTask)

		// v1.PUT("/tasks/:id", taskHandler.UpdateTask)

		// v1.DELETE("/tasks/:id", taskHandler.DeleteTask)
	}
}