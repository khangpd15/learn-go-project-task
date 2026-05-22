package handler

import (
	"errors"
	"net/http"
	"strconv"
	"task_api/internal/response"
	"task_api/internal/services"
	"task_api/internal/utils"
   TaskRequestDTO "task_api/internal/dto/request/task"
   "task_api/internal/mapper"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}
	tasks, err := h.service.GetAllTasks(currentUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to get tasks", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("get all tasks successfully", mapper.TasksToResponses(tasks)))
}
func (h *TaskHandler) GetTaskById(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", errors.New("invalid task ID").Error()))
		return
	}

	mappedTask, err := h.service.GetTaskById(currentUserID, id)
	if err != nil {
		if errors.Is(err, errors.New("forbidden")) {
			c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
			return
		}
		c.JSON(http.StatusNotFound, response.ErrorResponse("Task not found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Task found", mapper.ToTaskResponse(*mappedTask)))

}
func (h *TaskHandler) CreateTask(c *gin.Context) {

currentID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}
var req TaskRequestDTO.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
task := mapper.CreateTaskRequestToTaskEntity(req)
	createdTask, err := h.service.CreateTask(currentID, task)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to create task", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.SuccessResponse("Task created successfully", createdTask))

}
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    currentID,err := utils.CurrentUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
			return
		}
	taskID := c.Param("id")
	id, err := strconv.Atoi(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", err.Error()))
		return
	}
	var req TaskRequestDTO.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
updateTask := mapper.UpdateTaskRequestToTaskEntity(req)

	updatedTask, err := h.service.UpdateTask(id,currentID, updateTask)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to update task", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Task updated successfully", updatedTask))
}
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	id, err := strconv.Atoi(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", err.Error()))
		return
	}
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(401, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}
	err = h.service.DeleteTask(id, currentUserID)
	if err != nil {
		c.JSON(404, response.ErrorResponse("Failed to delete task", err.Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("Task deleted successfully", nil))
}
