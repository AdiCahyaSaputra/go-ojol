package dispatch

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	dispatchController := do.MustInvoke[controller.DispatchController](injector)
	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(verifier)

	dispatchRoutes := server.Group(constants.ROUTE_GROUP + "/dispatch")
	{
		dispatchRoutes.POST("/calculate-argo", authenticate, middlewares.Authorize(
			enforcer,
			constants.ENUM_RESOURCE_TRIP,
			constants.ENUM_ACTION_CREATE,
		), dispatchController.CalculateArgo)

		dispatchRoutes.GET("/find-driver", authenticate, middlewares.Authorize(
			enforcer,
			constants.ENUM_RESOURCE_TRIP,
			constants.ENUM_ACTION_READ,
		), dispatchController.FindDriver)
	}
}
