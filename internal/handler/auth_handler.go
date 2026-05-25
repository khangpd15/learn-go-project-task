package handler

import (
	"errors"
	"net/http"
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
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request", err.Error()))
		return
	}

	token, err := h.authService.Login(req)
	if err != nil {
		c.JSON(mapAuthErrorToStatus(err), response.ErrorResponse("Login failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Login successfully", gin.H{
		"access_token": token,
	}))
}
func (h *AuthHandler) LogOut(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessResponse(
		"Logout successfully",
		nil,
	))
}
func (h *AuthHandler) Register(c *gin.Context) {
	var req requestDTO.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request", err.Error()))
		return
	}
	err := h.authService.Register(req)
	if err != nil {
		c.JSON(mapAuthErrorToStatus(err), response.ErrorResponse("Register failed", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse(
		"Register successfully",
		nil,
	))
}

func mapAuthErrorToStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrInvalidInput),
		errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrInvalidPassword):
		return http.StatusBadRequest
	case errors.Is(err, services.ErrEmailAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, services.ErrInvalidCredentials),
		errors.Is(err, services.ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
