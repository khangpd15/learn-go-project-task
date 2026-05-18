package handler

import (
	"errors"
	"strconv"
	"task_api/internal/entities"
	"task_api/internal/response"
	"task_api/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Invalid user ID", errors.New("invalid user ID").Error()))
		return
	}
	user := h.service.GetUserByID(id)
	if user == nil {
		c.JSON(404, response.ErrorResponse("Failed to get user", errors.New("user not found").Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("User found", user))
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users := h.service.GetAllUsers()
	c.JSON(200, response.SuccessResponse("Get all users successfully", users))
}
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user := h.service.GetUserByUsername(username)
	if user == nil {
		c.JSON(404, response.ErrorResponse("Failed to get user", errors.New("user not found").Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("User found", user))
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user := h.service.GetUserByEmail(email)
	if user == nil {
		c.JSON(404, response.ErrorResponse("Failed to get user", errors.New("user not found").Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("User found", user))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user entities.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, response.ErrorResponse("Invalid request body", errors.New("invalid request body").Error()))
		return
	}
	createUser := h.service.CreateUser(user)
	c.JSON(201, response.SuccessResponse("User created successfully", createUser))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to update user", errors.New("invalid user ID").Error()))
		return
	}
	var updateUser entities.User
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(400, response.ErrorResponse("Failed to update user", errors.New("invalid request body").Error()))
		return
	}
	updatedUser := h.service.UpdateUser(id, updateUser)
	if updatedUser == nil {
		c.JSON(404, response.ErrorResponse("Failed to update user", errors.New("id not found").Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("User updated successfully", updatedUser))
}
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to delete user", errors.New("invalid user ID").Error()))
		return
	}
	err = h.service.DeleteUser(id)

	if err != nil {
		c.JSON(400, response.ErrorResponse("Failed to delete user", errors.New("id not found").Error()))
		return
	}
	c.JSON(200, response.SuccessResponse("User deleted successfully", nil))
}
