package dto

import (
	"errors"
	"mime/multipart"
	"time"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_REGISTER_USER      = "failed create user"
	MESSAGE_FAILED_GET_LIST_USER      = "failed get list user"
	MESSAGE_FAILED_TOKEN_NOT_VALID    = "token not valid"
	MESSAGE_FAILED_TOKEN_NOT_FOUND    = "token not found"
	MESSAGE_FAILED_GET_USER           = "failed get user"
	MESSAGE_FAILED_LOGIN              = "failed login"
	MESSAGE_FAILED_UPDATE_USER        = "failed update user"
	MESSAGE_FAILED_DELETE_USER        = "failed delete user"
	MESSAGE_FAILED_PROSES_REQUEST     = "failed proses request"
	MESSAGE_FAILED_DENIED_ACCESS      = "denied access"

	// Success
	MESSAGE_SUCCESS_REGISTER_USER = "success create user"
	MESSAGE_SUCCESS_GET_LIST_USER = "success get list user"
	MESSAGE_SUCCESS_GET_USER      = "success get user"
	MESSAGE_SUCCESS_LOGIN         = "success login"
	MESSAGE_SUCCESS_UPDATE_USER   = "success update user"
	MESSAGE_SUCCESS_DELETE_USER   = "success delete user"
)

var (
	ErrCreateUser         = errors.New("failed to create user")
	ErrGetUserById        = errors.New("failed to get user by id")
	ErrGetUserByEmail     = errors.New("failed to get user by email")
	ErrEmailAlreadyExists = errors.New("email already exist")
	ErrUpdateUser         = errors.New("failed to update user")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotFound      = errors.New("email not found")
	ErrDeleteUser         = errors.New("failed to delete user")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidRole           = errors.New("invalid role")
	ErrRoleNotAssigned       = errors.New("role not assigned")
	ErrVehicleRequired       = errors.New("vehicle details are required for driver registration")
	ErrLicenseNumberExists   = errors.New("vehicle license number already exists")
	ErrInvalidProfilePicture = errors.New("invalid profile picture")
	ErrUploadProfilePicture  = errors.New("failed to upload profile picture")
)

type (
	UserCreateRequest struct {
		Email                string                `json:"email" form:"email" binding:"required,email"`
		Password             string                `json:"password" form:"password" binding:"required,min=8"`
		Role                 string                `json:"role" form:"role" binding:"required,oneof=customer driver" validate:"required,oneof=customer driver"`
		Name                 string                `json:"name" form:"name" binding:"omitempty"`
		PhoneNumber          string                `json:"phone_number" form:"phone_number" binding:"omitempty"`
		VehicleName          string                `json:"vehicle_name" form:"vehicle_name" binding:"omitempty" validate:"required_if=Role driver"`
		VehicleLicenseNumber string                `json:"vehicle_license_number" form:"vehicle_license_number" binding:"omitempty" validate:"required_if=Role driver"`
		VehicleMaxSize       int                   `json:"vehicle_max_size" form:"vehicle_max_size" binding:"omitempty" validate:"required_if=Role driver,omitempty,gt=0"`
		VehicleType          string                `json:"vehicle_type" form:"vehicle_type" binding:"omitempty" validate:"required_if=Role driver,omitempty,oneof=car motorcycle"`
		ProfilePicture       *multipart.FileHeader `form:"profile_picture" binding:"omitempty"`
	}

	UserResponse struct {
		ID        string                   `json:"id"`
		Email     string                   `json:"email"`
		Role      string                   `json:"role,omitempty"`
		Customer  *CustomerProfileResponse `json:"customer,omitempty"`
		Driver    *DriverProfileResponse   `json:"driver,omitempty"`
		CreatedAt time.Time                `json:"created_at"`
		UpdatedAt time.Time                `json:"updated_at"`
	}

	CustomerProfileResponse struct {
		ID                string  `json:"id"`
		Name              string  `json:"name"`
		PhoneNumber       string  `json:"phone_number"`
		ProfilePictureUrl *string `json:"profile_picture_url,omitempty"`
	}

	VehicleProfileResponse struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		LicenseNumber string `json:"license_number"`
		MaxSize       int    `json:"max_size"`
		Type          string `json:"type"`
	}

	DriverProfileResponse struct {
		ID                string                  `json:"id"`
		Name              string                  `json:"name"`
		PhoneNumber       string                  `json:"phone_number"`
		Address           string                  `json:"address"`
		ProfilePictureUrl *string                 `json:"profile_picture_url,omitempty"`
		Vehicle           *VehicleProfileResponse `json:"vehicle,omitempty"`
	}

	UserUpdateRequest struct {
		Email string `json:"email" form:"email" binding:"omitempty,email"`
	}

	UserUpdateResponse struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	UserLoginRequest struct {
		Email    string `json:"email" form:"email" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}
)
