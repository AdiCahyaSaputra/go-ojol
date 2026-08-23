package service

import (
	"context"
	"strings"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	casbinrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/casbin/repository"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/helpers"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/uploadthing"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxProfilePictureBytes = 2 * 1024 * 1024

var allowedProfilePictureMIME = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error)
	Logout(ctx context.Context, userId string) error
}

type authService struct {
	userRepository   repository.UserRepository
	casbinRepository casbinrepo.CasbinRepository
	jwtService       JWTService
	enforcer         pkgcasbin.Enforcer
	uploadClient     uploadthing.Client
	db               *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	casbinRepo casbinrepo.CasbinRepository,
	jwtService JWTService,
	enforcer pkgcasbin.Enforcer,
	uploadClient uploadthing.Client,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:   userRepo,
		casbinRepository: casbinRepo,
		jwtService:       jwtService,
		enforcer:         enforcer,
		uploadClient:     uploadClient,
		db:               db,
	}
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	if req.Role != constants.ENUM_ROLE_CUSTOMER && req.Role != constants.ENUM_ROLE_DRIVER {
		return userDto.UserResponse{}, userDto.ErrInvalidRole
	}

	if req.Role == constants.ENUM_ROLE_DRIVER {
		if strings.TrimSpace(req.VehicleName) == "" ||
			strings.TrimSpace(req.VehicleLicenseNumber) == "" ||
			req.VehicleMaxSize <= 0 ||
			(req.VehicleType != string(entities.VehicleTypeCar) && req.VehicleType != string(entities.VehicleTypeMotorcycle)) {
			return userDto.UserResponse{}, userDto.ErrVehicleRequired
		}
	}

	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}

	var profilePictureURL *string
	if req.ProfilePicture != nil {
		url, err := s.uploadProfilePicture(ctx, req)
		if err != nil {
			return userDto.UserResponse{}, err
		}
		profilePictureURL = &url
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	user := entities.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: hashedPassword,
	}

	var (
		createdUser     entities.User
		createdCustomer *entities.Customer
		createdDriver   *entities.Driver
	)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createdUser, err = s.userRepository.Register(ctx, tx, user)
		if err != nil {
			return err
		}

		if err := s.casbinRepository.AddGroupingPolicy(ctx, tx, createdUser.Email, req.Role); err != nil {
			return err
		}

		switch req.Role {
		case constants.ENUM_ROLE_CUSTOMER:
			customer, createErr := s.userRepository.CreateCustomer(ctx, tx, entities.Customer{
				UserID:            createdUser.ID,
				Name:              req.Name,
				PhoneNumber:       req.PhoneNumber,
				ProfilePictureUrl: profilePictureURL,
			})
			if createErr != nil {
				return createErr
			}
			createdCustomer = &customer
		case constants.ENUM_ROLE_DRIVER:
			licenseNumber := strings.TrimSpace(req.VehicleLicenseNumber)
			existing, findErr := s.userRepository.FindVehicleByLicenseNumber(ctx, tx, licenseNumber)
			if findErr != nil && findErr != gorm.ErrRecordNotFound {
				return findErr
			}
			if existing != nil {
				return userDto.ErrLicenseNumberExists
			}

			vehicle, createVehicleErr := s.userRepository.CreateVehicle(ctx, tx, entities.Vehicle{
				Name:          strings.TrimSpace(req.VehicleName),
				LicenseNumber: licenseNumber,
				MaxSize:       req.VehicleMaxSize,
				Type:          entities.VehicleType(req.VehicleType),
			})
			if createVehicleErr != nil {
				return createVehicleErr
			}

			driver, createErr := s.userRepository.CreateDriver(ctx, tx, entities.Driver{
				UserID:            createdUser.ID,
				VehicleID:         vehicle.ID,
				Name:              req.Name,
				PhoneNumber:       req.PhoneNumber,
				Address:           "",
				ProfilePictureUrl: profilePictureURL,
			})
			if createErr != nil {
				return createErr
			}
			createdDriver = &driver
		}

		return nil
	})
	if err != nil {
		return userDto.UserResponse{}, err
	}

	if err := s.enforcer.LoadPolicy(); err != nil {
		return userDto.UserResponse{}, err
	}

	resp := userDto.UserResponse{
		ID:        createdUser.ID.String(),
		Email:     createdUser.Email,
		Role:      req.Role,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}

	if createdCustomer != nil {
		resp.Customer = &userDto.CustomerProfileResponse{
			ID:                createdCustomer.ID.String(),
			Name:              createdCustomer.Name,
			PhoneNumber:       createdCustomer.PhoneNumber,
			ProfilePictureUrl: createdCustomer.ProfilePictureUrl,
		}
	}

	if createdDriver != nil {
		// Intentionally omit vehicle from register response.
		resp.Driver = &userDto.DriverProfileResponse{
			ID:                createdDriver.ID.String(),
			Name:              createdDriver.Name,
			PhoneNumber:       createdDriver.PhoneNumber,
			Address:           createdDriver.Address,
			ProfilePictureUrl: createdDriver.ProfilePictureUrl,
		}
	}

	return resp, nil
}

func (s *authService) uploadProfilePicture(ctx context.Context, req userDto.UserCreateRequest) (string, error) {
	fileHeader := req.ProfilePicture
	if fileHeader.Size <= 0 || fileHeader.Size > maxProfilePictureBytes {
		return "", userDto.ErrInvalidProfilePicture
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return "", userDto.ErrInvalidProfilePicture
	}
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if _, ok := allowedProfilePictureMIME[contentType]; !ok {
		return "", userDto.ErrInvalidProfilePicture
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", userDto.ErrInvalidProfilePicture
	}
	defer file.Close()

	url, err := s.uploadClient.Upload(ctx, fileHeader.Filename, contentType, fileHeader.Size, file)
	if err != nil {
		return "", userDto.ErrUploadProfilePicture
	}
	return url, nil
}

func (s *authService) Login(ctx context.Context, req userDto.UserLoginRequest) (dto.TokenResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return dto.TokenResponse{}, userDto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return dto.TokenResponse{}, dto.ErrInvalidCredentials
	}

	roles, err := s.casbinRepository.GetRolesForUser(ctx, s.db, user.Email)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if len(roles) == 0 {
		return dto.TokenResponse{}, userDto.ErrRoleNotAssigned
	}

	role := roles[0]
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email, role)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken: accessToken,
		Role:        role,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userId string) error {
	return nil
}
