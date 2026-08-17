package providers

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/config"
	authController "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/controller"
	authService "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	userController "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	userService "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
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
