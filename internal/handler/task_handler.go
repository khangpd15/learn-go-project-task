package handler

import (
	"errors"
	"net/http"
	"strconv"

	TaskRequestDTO "task_api/internal/dto/request/task"
	"task_api/internal/mapper"
	"task_api/internal/response"
	"task_api/internal/services"
	"task_api/internal/utils"

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
	ctx := c.Request.Context()

	tasks, err := h.service.GetAllTasks(currentUserID, ctx)
	if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to get tasks", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Get all tasks successfully", mapper.TasksToResponses(tasks)))
}

func (h *TaskHandler) GetTaskById(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", "invalid task ID"))
		return
	}
	ctx := c.Request.Context()

	task, err := h.service.GetTaskById(currentUserID, ctx, id)
	if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to get task", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Task found", mapper.ToTaskResponse(*task)))
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	var req TaskRequestDTO.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
    ctx := c.Request.Context()
	task := mapper.CreateTaskRequestToTaskEntity(req)

	createdTask, err := h.service.CreateTask(currentUserID, ctx, task)
	if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to create task", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse("Task created successfully", mapper.ToTaskResponse(createdTask)))
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", "invalid task ID"))
		return
	}

	var req TaskRequestDTO.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}

	ctx := c.Request.Context()
	updatedTask, err := h.service.UpdateTask(id, currentUserID, ctx, req)
	if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to update task", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Task updated successfully", mapper.ToTaskResponse(*updatedTask)))
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", "invalid task ID"))
		return
	}

	ctx := c.Request.Context()
	if err := h.service.DeleteTask(id, currentUserID, ctx); err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to delete task", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Task deleted successfully", nil))
}
func (h *TaskHandler) AssignedTask(c *gin.Context){
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", "invalid task ID"))
		return
	}
	var req TaskRequestDTO.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
	ctx := c.Request.Context()
	assignTask, err := h.service.AssignTask(id, currentUserID, req.AssigneeID, ctx)
		if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to assign task", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Task assign successfully", mapper.ToTaskResponse(*assignTask)))
}
func (h *TaskHandler) UnassignTask(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", "invalid task ID"))
		return
	}

	ctx := c.Request.Context()

	updatedTask, err := h.service.UnassignTask(ctx, id, currentUserID)
	if err != nil {
		c.JSON(mapTaskErrorToStatus(err), response.ErrorResponse("Failed to unassign task", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Task unassigned successfully", updatedTask))
}

func mapTaskErrorToStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrInvalidTaskID),
		errors.Is(err, services.ErrInvalidProjectID),
		errors.Is(err, services.ErrInvalidStatus),
		errors.Is(err, services.ErrInvalidAssigneeID):
		return http.StatusBadRequest

	case errors.Is(err, services.ErrUnauthorized):
		return http.StatusUnauthorized

	case errors.Is(err, services.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, services.ErrTaskNotFound),
		errors.Is(err, services.ErrProjectNotFound):
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}
