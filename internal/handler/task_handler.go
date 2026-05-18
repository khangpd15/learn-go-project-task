package handler

import (
	"errors"
	"strconv"

	"task_api/internal/entities"
	"task_api/internal/response"
	"task_api/internal/services"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.service.GetAllTasks()
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to get tasks", err.Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("get all tasks successfully", tasks))
}
func (h *TaskHandler) GetTaskById(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Invalid task ID", errors.New("invalid task ID").Error()))
		return
	}
	task, err := h.service.GetTaskById(id)
	if err != nil {
		c.JSON(404, response.ErrorResponse("Failed to get task", err.Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("Task found", task))

}
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task entities.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}

	createTask, err := h.service.CreateTask(task)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to create task", err.Error()))
		return
	}

	c.JSON(201, response.SuccessResponse("Task created successfully", createTask))
}
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idString := c.Param("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to update task", "invalid task ID"))
		return
	}

	var updateTask entities.Task
	if err := c.ShouldBindJSON(&updateTask); err != nil {
		c.JSON(400, response.ErrorResponse("Failed to update task", err.Error()))
		return
	}

	updatedTask, err := h.service.UpdateTask(id, updateTask)
	if err != nil {
		c.JSON(404, response.ErrorResponse("Failed to update task", err.Error()))
		return
	}

	c.JSON(200, response.SuccessResponse("Task updated successfully", updatedTask))
}
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idString := c.Param("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to delete task", "invalid task ID"))
		return
	}

	err = h.service.DeleteTask(id)
	if err != nil {
		c.JSON(404, response.ErrorResponse("Failed to delete task", err.Error()))
		return
	}

	c.JSON(200, response.SuccessResponse("Task deleted successfully", nil))
}
