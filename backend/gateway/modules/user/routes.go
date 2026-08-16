package user

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/modules/auth/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/modules/user/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	userRoutes := server.Group("/api/user")
	{
		userRoutes.GET("", userController.GetAllUser)
		userRoutes.GET("/me", middlewares.Authenticate(jwtService), userController.Me)
		userRoutes.PUT("/:id", middlewares.Authenticate(jwtService), userController.Update)
		userRoutes.DELETE("/:id", middlewares.Authenticate(jwtService), userController.Delete)
	}
}
