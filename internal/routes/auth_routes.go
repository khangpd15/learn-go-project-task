package routes

import (
	"task_api/internal/handler"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authHandler *handler.AuthHandler
}

func NewAuthRoutes(authHandler *handler.AuthHandler) *AuthRoutes {
	return &AuthRoutes{
		authHandler: authHandler,
	}
}

func (r *AuthRoutes) SetupAuthRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", r.authHandler.Login)
		authGroup.POST("/register", r.authHandler.Register)
	}
}