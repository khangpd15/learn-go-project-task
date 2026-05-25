package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"task_api/internal/repositories"
	"task_api/internal/response"
	"task_api/internal/utils"
)

func AuthMiddleware(userRepo repositories.UserRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse(
				"missing Authorization header",
				"authorization header is required",
			),
			)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse(
				"invalid Authorization format",
				"Authorization header must be in the format 'Bearer <token>'",
			),
			)
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse(
				"invalid or expired token",
				"token validation failed: "),
			)
			c.Abort()
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized,
				response.ErrorResponse(
					"Missing Authorization header",
					"authorization header is required",
				),
			)
			c.Abort()
			return
		}

		userID := int(userIDFloat)

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse(
				"user not found",
				"the requested user was not found",
			))
			c.Abort()
			return
		}

		c.Set("current_user", user)
		c.Set("user_id", user.ID)

		c.Next()
	}
}
