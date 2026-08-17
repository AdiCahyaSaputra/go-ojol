package service

import (
	"context"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	casbinrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/casbin/repository"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	db               *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	casbinRepo casbinrepo.CasbinRepository,
	jwtService JWTService,
	enforcer pkgcasbin.Enforcer,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:   userRepo,
		casbinRepository: casbinRepo,
		jwtService:       jwtService,
		enforcer:         enforcer,
		db:               db,
	}
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	if req.Role != constants.ENUM_ROLE_CUSTOMER && req.Role != constants.ENUM_ROLE_DRIVER {
		return userDto.UserResponse{}, userDto.ErrInvalidRole
	}

	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
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

	var createdUser entities.User
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createdUser, err = s.userRepository.Register(ctx, tx, user)
		if err != nil {
			return err
		}

		return s.casbinRepository.AddGroupingPolicy(ctx, tx, createdUser.Email, req.Role)
	})
	if err != nil {
		return userDto.UserResponse{}, err
	}

	if err := s.enforcer.LoadPolicy(); err != nil {
		return userDto.UserResponse{}, err
	}

	return userDto.UserResponse{
		ID:        createdUser.ID.String(),
		Email:     createdUser.Email,
		Role:      req.Role,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}, nil
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
