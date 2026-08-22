package trip

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	tripController := do.MustInvoke[controller.TripController](injector)
	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(verifier)
	tripRead := middlewares.Authorize(enforcer, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_READ)

	tripRoutes := server.Group(constants.ROUTE_GROUP)
	{
		tripRoutes.GET("/protected", authenticate, tripRead, tripController.Protected)
	}
}
