package middleware

import (
	"net/http"

	"task_api/internal/entities"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		user, ok := value.(*entities.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "invalid user context",
			})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"message": "forbidden",
		})
		c.Abort()
	}
}