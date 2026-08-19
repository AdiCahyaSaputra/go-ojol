package auth

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)

	server.GET("/.well-known/jwks.json", authController.JWKS)

	authRoutes := server.Group(constants.ROUTE_GROUP)
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/logout", authController.Logout)
	}
}
