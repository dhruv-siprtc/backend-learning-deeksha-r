package routes

import (
	"go-backend-learning/backend/handlers"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo) {
	// Routes
	api := e.Group("/api")
	{
		users := api.Group("/users")
		users.POST("", handlers.CreateUser)
		users.GET("", handlers.GetUsers)
		users.GET("/:id", handlers.GetUserByID)
		users.PUT("/:id", handlers.UpdateUser)
		users.DELETE("/:id", handlers.DeleteUser)
	}

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
}
