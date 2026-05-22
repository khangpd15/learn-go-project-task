package routes

import (
	"task_api/internal/handler"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userHandler *handler.UserHandler
}

func NewUserRoutes(userHandler *handler.UserHandler) *UserRoutes {
	return &UserRoutes{
		userHandler: userHandler,
	}
}

func (r *UserRoutes) SetupUserRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("", r.userHandler.GetAllUsers)
		userGroup.GET("/:id", r.userHandler.GetUserByID)
		userGroup.GET("/email/:email", r.userHandler.GetUserByEmail)
		userGroup.GET("/fullname/:fullname", r.userHandler.GetUserByFullName)
		userGroup.POST("", r.userHandler.CreateUser)
		userGroup.PUT("/:id", r.userHandler.UpdateUser)
		userGroup.DELETE("/:id", r.userHandler.DeleteUser)
	}
}