package auth

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/controller"
	authrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/repository"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)
	sessionRepo := do.MustInvoke[authrepo.SessionRepository](injector)
	authenticate := middlewares.Authenticate(jwtService, sessionRepo)

	server.GET("/.well-known/jwks.json", authController.JWKS)

	authRoutes := server.Group(constants.ROUTE_GROUP)
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/refresh", authController.Refresh)
		authRoutes.POST("/logout", authenticate, authController.Logout)
		authRoutes.POST("/logout-all", authenticate, authController.LogoutAll)
		authRoutes.GET("/sessions", authenticate, authController.ListSessions)
		authRoutes.DELETE("/sessions/:id", authenticate, authController.RevokeSession)
	}
}
