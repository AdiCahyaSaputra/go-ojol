package providers

import (
	"net/http"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/config"
	dispatchController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	dispatchRepository "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/repository"
	dispatchService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	tripController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(injector, constants.JWKSVerifier, func(i *do.Injector) (jwks.Verifier, error) {
		return jwks.NewVerifierFromEnv()
	})

	do.ProvideNamed(injector, constants.CasbinEnforcer, func(i *do.Injector) (pkgcasbin.Enforcer, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return pkgcasbin.NewEnforcer(db)
	})

	do.Provide(injector, func(i *do.Injector) (tripController.TripController, error) {
		return tripController.NewTripController(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	dispatchRepo := dispatchRepository.NewDispatchRepository(db)
	dispatchSvc := dispatchService.NewDispatchService(
		dispatchRepo,
		db,
		&http.Client{Timeout: 10 * time.Second},
		"",
	)

	do.Provide(injector, func(i *do.Injector) (dispatchController.DispatchController, error) {
		return dispatchController.NewDispatchController(i, dispatchSvc), nil
	})
}
