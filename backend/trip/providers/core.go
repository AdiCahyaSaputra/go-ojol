package providers

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/config"
	tripController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
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

	do.Provide(injector, func(i *do.Injector) (tripController.TripController, error) {
		return tripController.NewTripController(), nil
	})
}
