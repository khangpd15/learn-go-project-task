package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"task_api/internal/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(userRepo repositories.UserRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDHeader := c.GetHeader("User-ID")

		if userIDHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "missing User-ID",
			})
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(userIDHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "User-ID must be a number",
			})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "user not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to authorize user",
				})
			}
			c.Abort()
			return
		}

		c.Set("current_user", user)

		c.Next()
	}
}