package dto

const (
	MESSAGE_SUCCESS_GET_ACTIVE_TRANSACTION = "Active transaction retrieved successfully"
	MESSAGE_SUCCESS_GET_TRANSACTION        = "Transaction retrieved successfully"
	MESSAGE_SUCCESS_START_TRIP           = "Trip started successfully"
	MESSAGE_SUCCESS_COMPLETE_TRIP        = "Trip completed successfully"
	MESSAGE_SUCCESS_CANCEL_TRIP          = "Trip cancelled successfully"

	MESSAGE_FAILED_GET_ACTIVE_TRANSACTION = "Failed to get active transaction"
	MESSAGE_FAILED_GET_TRANSACTION        = "Failed to get transaction"
	MESSAGE_FAILED_START_TRIP             = "Failed to start trip"
	MESSAGE_FAILED_COMPLETE_TRIP          = "Failed to complete trip"
	MESSAGE_FAILED_CANCEL_TRIP            = "Failed to cancel trip"

	MESSAGE_TRANSACTION_NOT_FOUND      = "transaction not found"
	MESSAGE_NO_ACTIVE_TRANSACTION      = "no active transaction"
	MESSAGE_NOT_TRANSACTION_PARTICIPANT = "not a participant of this transaction"
	MESSAGE_INVALID_STATUS_TRANSITION  = "invalid status transition"
)

type LatLongPair [2]float64

type TripDriverProfile struct {
	UserID        string `json:"user_id"`
	DriverID      string `json:"driver_id"`
	Name          string `json:"name"`
	PhoneNumber   string `json:"phone_number"`
	VehicleID     string `json:"vehicle_id"`
	VehicleName   string `json:"vehicle_name"`
	LicenseNumber string `json:"license_number"`
	VehicleType   string `json:"vehicle_type"`
}

type TransactionResponse struct {
	ID                   string             `json:"id"`
	Status               string             `json:"status"`
	PickupLatLong        LatLongPair        `json:"pickup_lat_long"`
	DestinationLatLong   LatLongPair        `json:"destination_lat_long"`
	DriverLastLatLong    LatLongPair        `json:"driver_last_lat_long"`
	CustomerLastLatLong  *LatLongPair       `json:"customer_last_lat_long,omitempty"`
	Distance             int                `json:"distance"`
	TotalFare            int                `json:"total_fare"`
	PaidAt               *string            `json:"paid_at,omitempty"`
	Driver               *TripDriverProfile `json:"driver,omitempty"`
}

type StartTripResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type CompleteTripResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	TotalFare     int    `json:"total_fare"`
	PaidAt        string `json:"paid_at"`
}

type CancelTripResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
