package providers

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/config"
	authController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/auth/controller"
	authService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/auth/service"
	userController "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/user/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/user/repository"
	userService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/user/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
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

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := repository.NewUserRepository(db)

	userService := userService.NewUserService(userRepository, db)
	authService := authService.NewAuthService(userRepository, jwtService, db)

	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authService), nil
		},
	)
}
