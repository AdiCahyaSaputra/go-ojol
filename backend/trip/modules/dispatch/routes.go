package dispatch

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	dispatchController := do.MustInvoke[controller.DispatchController](injector)

	dispatchRoutes := server.Group(constants.ROUTE_GROUP + "/dispatch")
	{
		dispatchRoutes.GET("/calculate-argo", dispatchController.CalculateArgo)
		dispatchRoutes.GET("/find-driver", dispatchController.FindDriver)
	}
}
