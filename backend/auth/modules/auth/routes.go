package auth

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)

	authRoutes := server.Group("/api/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/logout", authController.Logout)
	}
}
