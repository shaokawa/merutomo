package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shaokawa/merutomo/backend/internal/auth"
)

func Setup(r *gin.Engine, authHandler *auth.Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	authGroup := r.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.GET("/me", authHandler.RequireAuth(), authHandler.Me)
}
