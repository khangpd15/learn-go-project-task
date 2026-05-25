package middleware

import (
	"net/http"
    "task_api/internal/response"
	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse("Internal server error", err.(string)))
				c.Abort()
			}
		}()

		c.Next()
	}
}