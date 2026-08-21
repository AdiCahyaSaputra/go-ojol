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
)

type (
	CalculateArgoRequest struct {
		PickupLoc   [2]string `json:"pickup_loc" binding:"required,len=2" validate:"required,latlong"`
		Destination [2]string `json:"destination" binding:"required,len=2" validate:"required,latlong"`
		VehicleId   uuid.UUID `json:"vehicle_id" binding:"required" validate:"required"`
	}

	CalculateArgoResponse struct {
		Distance           int          `json:"distance"`
		Duration           int          `json:"duration"`
		Path               [][2]float64 `json:"path"`
		FarePerDistance    int          `json:"fare_per_distance"`
		PlatformPercentage int          `json:"platform_percentage"`
		TotalFare          int          `json:"total_fare"`
	}

	FindDriverRequest struct {
		CurrentLocation [2]string            `json:"current_location" form:"current_location" binding:"required,len=2" validate:"required,latlong"`
		VehicleType     entities.VehicleType `json:"vehicle_type" binding:"required" validate:"required,vehicle_type"`
	}

	FindDriverResponse struct {
		Drivers []NearbyDriver `json:"drivers"`
	}

	NearbyDriver struct {
		UserID    string     `json:"user_id"`
		DistanceM int        `json:"distance_m"`
		Location  [2]float64 `json:"location"`
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
)
