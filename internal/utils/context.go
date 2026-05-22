package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CurrentUserID(c *gin.Context) (int, error) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("current user not found in context")
	}

	userID, ok := value.(int)
	if !ok {
		return 0, errors.New("invalid current user id in context")
	}

	return userID, nil
}
