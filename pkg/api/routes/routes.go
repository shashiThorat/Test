package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/user/go-git-test/pkg/api/handlers"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, h *handlers.Handlers) {
	// API v1 group
	v1 := e.Group("/api/v1")
	
	// Health check
	v1.GET("/health", h.HealthCheck)
	
	// Git operations
	v1.POST("/clone", h.Clone)
	v1.GET("/prdiff", h.GetPRDiff)
	v1.POST("/prcomment", h.AddPRComment)
	
	// PR Comments operations
	v1.GET("/prcomments/parent", h.GetPRCommentsByParentID)
}