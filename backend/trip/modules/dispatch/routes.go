package dispatch

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	dispatchController := do.MustInvoke[controller.DispatchController](injector)
	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)

	dispatchRoutes := server.Group(constants.ROUTE_GROUP + "/dispatch")
	{
		dispatchRoutes.POST("/calculate-argo", middlewares.Authenticate(verifier), dispatchController.CalculateArgo)
		dispatchRoutes.POST("/find-driver", middlewares.Authenticate(verifier), dispatchController.FindDriver)
	}
}
