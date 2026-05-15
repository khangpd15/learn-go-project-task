package middleware

import (
	"net/http"
	"strconv"

	"task/api/internal/repositories"

	"github.com/gin-gonic/gin"
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

		user := userRepo.GetUserByID(userID)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "user not found",
			})
			c.Abort()
			return
		}

		c.Set("current_user", user)

		c.Next()
	}
}