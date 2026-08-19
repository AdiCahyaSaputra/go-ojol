package dto

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_SUCCESS_GET_DATA          = "success get data"

	MESSAGE_DRIVER_ID_NOT_SELECTED = "driver_id is required"
)

type (
	CalculateArgoRequest struct {
		PickupLoc   [2]string            `json:"pickup_loc" binding:"required,len=2" validate:"required,latlong"`
		Destination [2]string            `json:"destination" binding:"required,len=2" validate:"required,latlong"`
		VehicleType entities.VehicleType `json:"vehicle_type" binding:"required" validate:"required,vehicle_type"`
	}

	FindDriverRequest struct {
		CurrentLocation [2]string `json:"current_location" binding:"required,len=2" validate:"required,latlong"`
	}
)
