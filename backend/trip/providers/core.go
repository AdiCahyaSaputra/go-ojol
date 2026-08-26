package providers

import (
	"net/http"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/config"
	dispatchController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	dispatchRepository "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/repository"
	dispatchService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	dispatchwsController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/controller"
	dispatchwsService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/service"
	savedAddressController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/controller"
	savedAddressRepository "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/repository"
	savedAddressService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/service"
	tripController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func InitRedis(injector *do.Injector) {
	do.ProvideNamed(injector, constants.Redis, func(i *do.Injector) (*redis.Client, error) {
		return config.SetUpRedis(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)
	InitRedis(injector)

	do.ProvideNamed(injector, constants.JWKSVerifier, func(i *do.Injector) (jwks.Verifier, error) {
		return jwks.NewVerifierFromEnv()
	})

	do.ProvideNamed(injector, constants.SessionChecker, func(i *do.Injector) (session.Checker, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return session.NewRepository(db), nil
	})

	do.ProvideNamed(injector, constants.CasbinEnforcer, func(i *do.Injector) (pkgcasbin.Enforcer, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return pkgcasbin.NewEnforcer(db)
	})

	do.Provide(injector, func(i *do.Injector) (tripController.TripController, error) {
		return tripController.NewTripController(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	rdb := do.MustInvokeNamed[*redis.Client](injector, constants.Redis)
	locations := drivergeo.NewStore(rdb)

	wsSvc := dispatchwsService.NewDispatchWSService(locations)
	do.Provide(injector, func(i *do.Injector) (dispatchwsController.DispatchWSController, error) {
		return dispatchwsController.NewDispatchWSController(wsSvc), nil
	})

	dispatchRepo := dispatchRepository.NewDispatchRepository(db)
	dispatchSvc := dispatchService.NewDispatchService(
		dispatchRepo,
		db,
		&http.Client{Timeout: 10 * time.Second},
		"",
		locations,
		wsSvc,
	)
	wsSvc.SetOfferRetrier(dispatchSvc)

	do.Provide(injector, func(i *do.Injector) (dispatchController.DispatchController, error) {
		return dispatchController.NewDispatchController(i, dispatchSvc), nil
	})

	savedAddressRepo := savedAddressRepository.NewSavedAddressRepository(db)
	savedAddressSvc := savedAddressService.NewSavedAddressService(savedAddressRepo, db)
	do.Provide(injector, func(i *do.Injector) (savedAddressController.SavedAddressController, error) {
		return savedAddressController.NewSavedAddressController(i, savedAddressSvc), nil
	})
}
