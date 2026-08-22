package dispatch

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	dispatchController := do.MustInvoke[controller.DispatchController](injector)

	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(verifier)

	dispatchCustomerRoutes := server.Group(constants.ROUTE_GROUP + "/dispatch/customer")
	{
		dispatchCustomerRoutes.POST("/calculate-argo",
			authenticate,
			middlewares.ResolveProfileId(db), // customer_id lookup
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_DISPATCH,
				constants.ENUM_ACTION_CREATE,
			),
			dispatchController.CalculateArgo,
		)

		dispatchCustomerRoutes.POST("/find-driver", authenticate, middlewares.Authorize(
			enforcer,
			constants.ENUM_ROLE_CUSTOMER,
			constants.ENUM_RESOURCE_DISPATCH,
			constants.ENUM_ACTION_READ,
		), dispatchController.FindDriver)
	}

	dispatchDriverRoutes := server.Group(constants.ROUTE_GROUP + "/dispatch/driver")
	{
		dispatchDriverRoutes.POST("/mode",
			authenticate,
			middlewares.ResolveProfileId(db), // customer_id lookup
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_DRIVER,
				constants.ENUM_RESOURCE_DISPATCH,
				constants.ENUM_ACTION_UPDATE,
			),
			dispatchController.SetDriverMode,
		)
	}
}
