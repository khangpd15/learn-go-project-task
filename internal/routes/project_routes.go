package routes

import (
	"task_api/internal/handler"

	"github.com/gin-gonic/gin"
)

type ProjectRoutes struct {
	projectHandler *handler.ProjectHandler
}

func NewProjectRoutes(projectHandler *handler.ProjectHandler) *ProjectRoutes {
	return &ProjectRoutes{
		projectHandler: projectHandler,
	}
}

func (r *ProjectRoutes) SetupProjectRoutes(router *gin.RouterGroup) {
	projectGroup := router.Group("/projects")
	{
		projectGroup.GET("/me", r.projectHandler.ListMyProjects)
		projectGroup.GET("", r.projectHandler.ListAllProjects)
		projectGroup.GET("/:id", r.projectHandler.GetProjectByID)
		projectGroup.PUT("/:id", r.projectHandler.UpdateProject)
		projectGroup.DELETE("/:id", r.projectHandler.DeleteProject)
	}
}
