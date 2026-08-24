package dto

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/google/uuid"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_SUCCESS_GET_DATA          = "success get data"

	MESSAGE_SUCCESS_CALCULATE_ARGO = "success calculate argo"
	MESSAGE_FAILED_CALCULATE_ARGO  = "failed calculate argo"
	MESSAGE_CUSTOMER_NOT_FOUND_CTX = "profile customer is not resolved in the context"

	MESSAGE_DRIVER_ID_NOT_SELECTED = "driver_id is required"
	MESSAGE_VEHICLE_NOT_FOUND      = "vehicle doesn't exist"

	MESSAGE_SUCCESS_FIND_DRIVER = "success find driver"
	MESSAGE_FAILED_FIND_DRIVER  = "failed find driver"

	MESSAGE_DRIVE_MODE_INVALID              = "driver mode is invalid, avaliable mode is online or offline"
	MESSAGE_DRIVE_USER_ID_CONTEXT_NOT_FOUND = "can't find user_id in driver context"
	MESSAGE_SET_DRIVER_MODE_FAILED          = "set mode for driver failed"
	MESSAGE_SET_DRIVER_MODE_SUCCESS         = "set mode for driver success"
)

type DriverMode string

const (
	DriverModeOnline  = "online"
	DriverModeOffline = "offline"
)

type (
	CalculateArgoRequest struct {
		PickupLoc   [2]string `form:"pickup_loc" binding:"required,len=2" validate:"required,latlong"`
		Destination [2]string `form:"destination" binding:"required,len=2" validate:"required,latlong"`
	}

	VehicleCategory struct {
		VehicleType entities.VehicleType `json:"vehicle_type" gorm:"column:type"`
		MaxSize     int                  `json:"max_size" gorm:"column:max_size"`
	}

	VehicleOption struct {
		VehicleType     entities.VehicleType `json:"vehicle_type"`
		MaxSize         int                  `json:"max_size"`
		TotalFare       int                  `json:"total_fare"`
	}

	CalculateArgoResponse struct {
		Distance           int             `json:"distance"`
		Duration           int             `json:"duration"`
		Path               [][2]float64    `json:"path"`
		PlatformPercentage int             `json:"platform_percentage"`
		VehicleOptions     []VehicleOption `json:"vehicle_options"`
	}

	FindDriverRequest struct {
		CurrentLatLong [2]string            `json:"current_lat_long" form:"current_location" binding:"required,len=2" validate:"required,latlong"`
		VehicleType    entities.VehicleType `json:"vehicle_type" binding:"required" validate:"required,vehicle_type"`
	}

	FindDriverResponse struct {
		Drivers []NearbyDriver `json:"drivers"`
	}

	NearbyDriverProfile struct {
		UserID            uuid.UUID            `json:"user_id"`
		DriverID          uuid.UUID            `json:"driver_id"`
		Name              string               `json:"name"`
		PhoneNumber       string               `json:"phone_number"`
		ProfilePictureUrl *string              `json:"profile_picture_url"`
		VehicleID         uuid.UUID            `json:"vehicle_id"`
		VehicleName       string               `json:"vehicle_name"`
		LicenseNumber     string               `json:"license_number"`
		MaxSize           int                  `json:"max_size"`
		Type              entities.VehicleType `json:"type"`
	}

	NearbyDriver struct {
		UserID    string              `json:"user_id"`
		DistanceM int                 `json:"distance_m"`
		Location  [2]float64          `json:"location"`
		Profile   NearbyDriverProfile `json:"profile"`
	}

	PendingArgoTransaction struct {
		CustomerID uuid.UUID `json:"customer_id"`
		VehicleID  uuid.UUID `json:"vehicle_id"`

		PickupLatLong      [2]string `json:"pickup_lat_long"`
		DestinationLatLong [2]string `json:"destination_lat_long"`
		LastLatLong        [2]string `json:"last_lat_long"` // Using pickup long by default

		Distance           int `json:"distance"`
		FarePerDistance    int `json:"fare_per_distance"`
		PlatformPercentage int `json:"platform_percentage"`
		TotalFare          int `json:"total_fare"`
	}

	SetDriverModeRequest struct {
		Mode           DriverMode `json:"mode" validate:"required,driver_mode"`
		CurrentLatLong [2]string  `json:"current_lat_long" validate:"required,latlong"`
	}
)
