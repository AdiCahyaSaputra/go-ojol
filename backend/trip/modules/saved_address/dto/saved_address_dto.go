package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_CUSTOMER_NOT_FOUND_CTX    = "profile customer is not resolved in the context"

	MESSAGE_SUCCESS_LIST_SAVED_ADDRESS   = "success list saved addresses"
	MESSAGE_FAILED_LIST_SAVED_ADDRESS    = "failed list saved addresses"
	MESSAGE_SUCCESS_GET_SAVED_ADDRESS    = "success get saved address"
	MESSAGE_FAILED_GET_SAVED_ADDRESS     = "failed get saved address"
	MESSAGE_SUCCESS_CREATE_SAVED_ADDRESS = "success create saved address"
	MESSAGE_FAILED_CREATE_SAVED_ADDRESS  = "failed create saved address"
	MESSAGE_SUCCESS_UPDATE_SAVED_ADDRESS = "success update saved address"
	MESSAGE_FAILED_UPDATE_SAVED_ADDRESS  = "failed update saved address"
	MESSAGE_SUCCESS_DELETE_SAVED_ADDRESS = "success delete saved address"
	MESSAGE_FAILED_DELETE_SAVED_ADDRESS  = "failed delete saved address"
)

var (
	ErrSavedAddressNotFound = errors.New("saved address not found")
	ErrInvalidLatLong       = errors.New("invalid lat long")
	ErrCustomerNotInCtx     = errors.New("customer not in context")
)

type (
	SavedAddressCreateRequest struct {
		Name            string    `json:"name" binding:"required" validate:"required,min=1,max=255"`
		LatLong         [2]string `json:"lat_long" binding:"required,len=2" validate:"required,latlong"`
		IsDefaultPickup bool      `json:"is_default_pickup"`
	}

	SavedAddressUpdateRequest struct {
		Name            string    `json:"name" binding:"required" validate:"required,min=1,max=255"`
		LatLong         [2]string `json:"lat_long" binding:"required,len=2" validate:"required,latlong"`
		IsDefaultPickup bool      `json:"is_default_pickup"`
	}

	SavedAddressResponse struct {
		ID              uuid.UUID `json:"id"`
		Name            string    `json:"name"`
		LatLong         [2]string `json:"lat_long"`
		IsDefaultPickup bool      `json:"is_default_pickup"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}
)
