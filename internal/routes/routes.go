package routes

import (
	"task_api/internal/entities"
	"task_api/internal/handler"
	"task_api/internal/middleware"
	"task_api/internal/repositories"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	taskHandler *handler.TaskHandler,
	userRepo repositories.UserRepositoryInterface,
	userHandler *handler.UserHandler,
) {
	// Route test server sống chưa, không cần login
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server is running",
		})
	})

	v1 := r.Group("/api/v1")

	// Tất cả route bên dưới đều phải có User-ID
	v1.Use(middleware.AuthMiddleware(userRepo))

	{
		// GUEST, CUSTOMER, ADMIN đều được xem danh sách
		v1.GET(
			"/tasks",
			middleware.RoleMiddleware(
				entities.RoleGuest,
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			taskHandler.GetAllTasks,
		)

		// CUSTOMER, ADMIN được xem chi tiết
		v1.GET(
			"/tasks/:id",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			taskHandler.GetTaskById,
		)

		// CUSTOMER, ADMIN được tạo task
		v1.POST(
			"/tasks",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			taskHandler.CreateTask,
		)

		// CUSTOMER, ADMIN được cập nhật task
		v1.PUT(
			"/tasks/:id",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			taskHandler.UpdateTask,
		)

		// Chỉ ADMIN được xóa task
		v1.DELETE(
			"/tasks/:id",
			middleware.RoleMiddleware(
				entities.RoleAdmin,
			),
			taskHandler.DeleteTask,
		)
	}

	v2 := r.Group("/api/v2")
	v2.Use(middleware.AuthMiddleware(userRepo))
	{
		// GUEST, CUSTOMER, ADMIN đều được xem danh sách
		v2.GET(
			"/users",
			middleware.RoleMiddleware(
				entities.RoleGuest,
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			userHandler.GetAllUsers,
		)

		// CUSTOMER, ADMIN được xem chi tiết
		v2.GET(
			"/users/:id",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			userHandler.GetUserByID,
		)

		// CUSTOMER, ADMIN được tạo user
		v2.POST(
			"/users",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			userHandler.CreateUser,
		)

		// CUSTOMER, ADMIN được cập nhật user
		v2.PUT(
			"/users/:id",
			middleware.RoleMiddleware(
				entities.RoleCustomer,
				entities.RoleAdmin,
			),
			userHandler.UpdateUser,
		)

		// Chỉ ADMIN được xóa user
		v2.DELETE(
			"/users/:id",
			middleware.RoleMiddleware(
				entities.RoleAdmin,
			),
			userHandler.DeleteUser,
		)
	}
}
