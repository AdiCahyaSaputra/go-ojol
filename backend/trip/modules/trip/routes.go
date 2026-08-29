package trip

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	tripController := do.MustInvoke[controller.TripController](injector)
	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)
	sessions := do.MustInvokeNamed[session.Checker](injector, constants.SessionChecker)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(verifier, sessions)
	resolveProfile := middlewares.ResolveProfileId(db)
	tripRead := middlewares.Authorize(enforcer, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_READ)
	tripUpdate := middlewares.Authorize(enforcer, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE)
	driverUpdate := middlewares.Authorize(enforcer, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE)

	tripRoutes := server.Group(constants.ROUTE_GROUP)
	{
		tripRoutes.GET("/protected", authenticate, tripRead, tripController.Protected)

		transactionRoutes := tripRoutes.Group("/transactions")
		{
			transactionRoutes.GET("/active", authenticate, resolveProfile, tripRead, tripController.GetActive)
			transactionRoutes.GET("/:id", authenticate, resolveProfile, tripRead, tripController.GetByID)
			transactionRoutes.POST("/:id/start", authenticate, resolveProfile, driverUpdate, tripController.StartTrip)
			transactionRoutes.POST("/:id/complete", authenticate, resolveProfile, driverUpdate, tripController.CompleteTrip)
			transactionRoutes.POST("/:id/cancel", authenticate, resolveProfile, tripUpdate, tripController.CancelTrip)
		}
	}
}
