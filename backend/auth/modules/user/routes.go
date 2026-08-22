package user

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(jwtService)
	userRead := middlewares.Authorize(enforcer, "", constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_READ)
	userUpdate := middlewares.Authorize(enforcer, "", constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_UPDATE)
	userDelete := middlewares.Authorize(enforcer, constants.ENUM_ROLE_ADMIN, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_DELETE)

	userRoutes := server.Group(constants.ROUTE_GROUP + "/user")
	{
		userRoutes.GET("", authenticate, userRead, userController.GetAllUser)
		userRoutes.GET("/me", authenticate, userRead, userController.Me)
		userRoutes.PUT("/:id", authenticate, userUpdate, userController.Update)
		userRoutes.DELETE("/:id", authenticate, userDelete, userController.Delete)
	}
}
