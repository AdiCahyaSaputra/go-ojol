package service

import (
	"context"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
	Delete(ctx context.Context, userId string) error
}

type userService struct {
	userRepository repository.UserRepository
	db             *gorm.DB
}

func NewUserService(
	userRepo repository.UserRepository,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepository: userRepo,
		db:             db,
	}
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserResponse{}, err
	}

	resp := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.Customer != nil {
		resp.Customer = &dto.CustomerProfileResponse{
			ID:                user.Customer.ID.String(),
			Name:              user.Customer.Name,
			PhoneNumber:       user.Customer.PhoneNumber,
			ProfilePictureUrl: user.Customer.ProfilePictureUrl,
		}
	}

	if user.Driver != nil {
		resp.Driver = &dto.DriverProfileResponse{
			ID:                user.Driver.ID.String(),
			Name:              user.Driver.Name,
			PhoneNumber:       user.Driver.PhoneNumber,
			Address:           user.Driver.Address,
			ProfilePictureUrl: user.Driver.ProfilePictureUrl,
		}
		if user.Driver.Vehicle.ID != uuid.Nil && user.Driver.Vehicle.ID == user.Driver.VehicleID {
			resp.Driver.Vehicle = &dto.VehicleProfileResponse{
				ID:            user.Driver.Vehicle.ID.String(),
				Name:          user.Driver.Vehicle.Name,
				LicenseNumber: user.Driver.Vehicle.LicenseNumber,
				MaxSize:       user.Driver.Vehicle.MaxSize,
				Type:          string(user.Driver.Vehicle.Type),
			}
		}
	}

	return resp, nil
}

func (s *userService) Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	updatedUser, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.UserUpdateResponse{}, err
	}

	return dto.UserUpdateResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *userService) Delete(ctx context.Context, userId string) error {
	return s.userRepository.Delete(ctx, s.db, userId)
}
