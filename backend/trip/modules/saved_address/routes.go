package saved_address

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	savedAddressController := do.MustInvoke[controller.SavedAddressController](injector)

	verifier := do.MustInvokeNamed[jwks.Verifier](injector, constants.JWKSVerifier)
	sessions := do.MustInvokeNamed[session.Checker](injector, constants.SessionChecker)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	authenticate := middlewares.Authenticate(verifier, sessions)

	savedAddressRoutes := server.Group(constants.ROUTE_GROUP + "/saved-addresses")
	{
		savedAddressRoutes.GET("",
			authenticate,
			middlewares.ResolveProfileId(db),
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_SAVED_ADDRESS,
				constants.ENUM_ACTION_READ,
			),
			savedAddressController.List,
		)

		savedAddressRoutes.GET("/:id",
			authenticate,
			middlewares.ResolveProfileId(db),
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_SAVED_ADDRESS,
				constants.ENUM_ACTION_READ,
			),
			savedAddressController.GetByID,
		)

		savedAddressRoutes.POST("",
			authenticate,
			middlewares.ResolveProfileId(db),
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_SAVED_ADDRESS,
				constants.ENUM_ACTION_CREATE,
			),
			savedAddressController.Create,
		)

		savedAddressRoutes.PUT("/:id",
			authenticate,
			middlewares.ResolveProfileId(db),
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_SAVED_ADDRESS,
				constants.ENUM_ACTION_UPDATE,
			),
			savedAddressController.Update,
		)

		savedAddressRoutes.DELETE("/:id",
			authenticate,
			middlewares.ResolveProfileId(db),
			middlewares.Authorize(
				enforcer,
				constants.ENUM_ROLE_CUSTOMER,
				constants.ENUM_RESOURCE_SAVED_ADDRESS,
				constants.ENUM_ACTION_DELETE,
			),
			savedAddressController.Delete,
		)
	}
}
