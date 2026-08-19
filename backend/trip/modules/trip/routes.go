package trip

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	tripController := do.MustInvoke[controller.TripController](injector)
	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)

	tripRoutes := server.Group(constants.ROUTE_GROUP)
	{
		tripRoutes.GET("/protected", middlewares.Authenticate(verifier), tripController.Protected)
	}
}
