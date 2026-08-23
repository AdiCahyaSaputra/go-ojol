package providers

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/config"
	authController "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/controller"
	authrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/repository"
	authService "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	casbinrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/casbin/repository"
	userController "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	userService "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/service"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/uploadthing"
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
		return authService.NewJWTService()
	})

	do.ProvideNamed(injector, constants.CasbinEnforcer, func(i *do.Injector) (pkgcasbin.Enforcer, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return pkgcasbin.NewEnforcer(db)
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)
	enforcer := do.MustInvokeNamed[pkgcasbin.Enforcer](injector, constants.CasbinEnforcer)

	userRepository := repository.NewUserRepository(db)
	casbinRepository := casbinrepo.NewCasbinRepository(db)
	sessionRepository := authrepo.NewSessionRepository(db)

	do.Provide(injector, func(i *do.Injector) (authrepo.SessionRepository, error) {
		return sessionRepository, nil
	})

	uploadClient, err := uploadthing.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	userService := userService.NewUserService(userRepository, db)
	authSvc := authService.NewAuthService(userRepository, casbinRepository, sessionRepository, jwtService, enforcer, uploadClient, db)

	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authSvc), nil
		},
	)
}
