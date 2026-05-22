package handler

import (
	"net/http"
	"strconv"
	"errors"
	requestDTO "task_api/internal/dto/request/project"
	"task_api/internal/mapper"
	"task_api/internal/response"
	"task_api/internal/services"
	"task_api/internal/utils"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) ListMyProjects(c *gin.Context) {
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	projects, err := h.service.ListProjectsByOwner(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to list projects", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Projects listed successfully", mapper.ProjectsToResponses(projects)))
}
func (h *ProjectHandler) ListAllProjects(c *gin.Context) {
	projects, err := h.service.ListAllProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to list projects", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Projects listed successfully", mapper.ProjectsToResponses(projects)))
}
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
		return
	}
	project, err := h.service.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to get project", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Project retrieved successfully", mapper.ProjectToResponse(project)))
}
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
		return
	}
	var req requestDTO.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	err = h.service.UpdateProject(currentUserID, projectID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProjectID) {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, response.ErrorResponse("Project not found", err.Error()))
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to update project", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Project updated successfully", nil))
}
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
		return
	}
	currentUserID, err := utils.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	err = h.service.DeleteProject(currentUserID, projectID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidProjectID) {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, response.ErrorResponse("Project not found", err.Error()))
			return
		}
		if errors.Is(err, services.ErrForbidden) {
			c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to delete project", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse("Project deleted successfully", nil))
}
