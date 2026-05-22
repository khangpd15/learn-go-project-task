package middleware

import (
	"net/http"
	"strings"

	"task_api/internal/repositories"
	"task_api/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(userRepo repositories.UserRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "missing Authorization header",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid Authorization format",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token payload",
			})
			c.Abort()
			return
		}

		userID := int(userIDFloat)

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "user not found",
			})
			c.Abort()
			return
		}

		c.Set("current_user", user)
		c.Set("user_id", user.ID)

		c.Next()
	}
}