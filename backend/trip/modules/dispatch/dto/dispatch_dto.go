package dto

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_SUCCESS_GET_DATA          = "success get data"

	MESSAGE_SUCCESS_CALCULATE_ARGO = "success calculate argo"
	MESSAGE_FAILED_CALCULATE_ARGO  = "failed calculate argo"

	MESSAGE_DRIVER_ID_NOT_SELECTED = "driver_id is required"
)

type (
	CalculateArgoRequest struct {
		PickupLoc   [2]string            `json:"pickup_loc" binding:"required,len=2" validate:"required,latlong"`
		Destination [2]string            `json:"destination" binding:"required,len=2" validate:"required,latlong"`
		VehicleType entities.VehicleType `json:"vehicle_type" binding:"required" validate:"required,vehicle_type"`
	}

	CalculateArgoResponse struct {
		Distance           int                  `json:"distance"`
		Duration           int                  `json:"duration"`
		Path               [][2]float64         `json:"path"`
		FarePerDistance    int                  `json:"fare_per_distance"`
		PlatformPercentage int                  `json:"platform_percentage"`
		TotalFare          int                  `json:"total_fare"`
		VehicleType        entities.VehicleType `json:"vehicle_type"`
	}

	FindDriverRequest struct {
		CurrentLocation [2]string `json:"current_location" binding:"required,len=2" validate:"required,latlong"`
	}
)
