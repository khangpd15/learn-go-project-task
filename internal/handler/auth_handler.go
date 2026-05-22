package handler

import (
	requestDTO "task_api/internal/dto/request/auth"
	"task_api/internal/response"
	"task_api/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req requestDTO.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.ErrorResponse("Invalid request", err.Error()))
		return
	}

	token, err := h.authService.Login(req)
	if err != nil {
		c.JSON(401, response.ErrorResponse("Login failed", err.Error()))
		return
	}
  
	c.JSON(200, response.SuccessResponse("Login successfully", gin.H{
		"access_token": token,
	}))
}
func (h *AuthHandler) LogOut(c *gin.Context) {
	c.JSON(200, response.SuccessResponse(
		"Logout successfully",
		nil,
	))
}
func (h *AuthHandler) Register(c *gin.Context){
	var req requestDTO.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.ErrorResponse("Invalid request", err.Error()))
		return
	}
	err := h.authService.Register(req)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Register failed", err.Error()))
		return
	}
	c.JSON(200, response.SuccessResponse(
		"Register successfully",
		nil,
	))
}