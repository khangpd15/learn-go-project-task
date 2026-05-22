package routes

import (
	"task_api/internal/handler"

	"github.com/gin-gonic/gin"
)

type TaskRoutes struct {
	taskHandler *handler.TaskHandler
}

func NewTaskRoutes(taskHandler *handler.TaskHandler) *TaskRoutes {
	return &TaskRoutes{
		taskHandler: taskHandler,
	}
}

func (r *TaskRoutes) SetupTaskRoutes(router *gin.RouterGroup) {
	taskGroup := router.Group("/tasks")
	{
		taskGroup.GET("", r.taskHandler.GetAllTasks)
		taskGroup.GET("/:id", r.taskHandler.GetTaskById)
		taskGroup.POST("", r.taskHandler.CreateTask)
		taskGroup.PUT("/:id", r.taskHandler.UpdateTask)
		taskGroup.DELETE("/:id", r.taskHandler.DeleteTask)
	}
}