package handler

import (
	"errors"
	"net/http"
	"strconv"
	requestDTO "task_api/internal/dto/request/user"
	"task_api/internal/mapper"
	"task_api/internal/response"
	"task_api/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid user ID", "invalid user ID"))
		return
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to get user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("User found", mapper.UserToResponse(*user)))
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to get users", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("Get all users successfully", mapper.UsersToResponses(users)))
}

func (h *UserHandler) GetUserByFullName(c *gin.Context) {
	fullName := c.Param("fullname")

	user, err := h.service.GetUserByFullName(fullName)
	if err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to get user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("User found", mapper.UserToResponse(*user)))
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.service.GetUserByEmail(email)
	if err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to get user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("User found", mapper.UserToResponse(*user)))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req requestDTO.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
		return
	}
    
	userEntity := mapper.UserToEntity(req)

	createdUser, err := h.service.CreateUser(userEntity)
	if err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to create user", err.Error()))
		return
	}
     
	c.JSON(http.StatusCreated, response.SuccessResponse("User created successfully", mapper.UserToResponse(createdUser)))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to update user", "invalid user ID"))
		return
	}

	var req requestDTO.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to update user", err.Error()))
		return
	}

	updateUser := mapper.ToUpdateUserEntity(req)
	updatedUser, err := h.service.UpdateUser(id, updateUser)
	if err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to update user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("User updated successfully", mapper.UserToResponse(*updatedUser)))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to delete user", "invalid user ID"))
		return
	}

	if err := h.service.DeleteUser(id); err != nil {
		c.JSON(mapUserErrorToStatus(err), response.ErrorResponse("Failed to delete user", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse("User deleted successfully", nil))
}

func mapUserErrorToStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrInvalidUserID),
		errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrInvalidPassword):
		return http.StatusBadRequest
	case errors.Is(err, services.ErrEmailAlreadyExist):
		return http.StatusConflict
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}