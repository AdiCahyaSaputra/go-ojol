package dto

const (
	TypeStandby      = "standby"
	TypeLocation     = "location"
	TypeStandbyOK    = "standby_ok"
	TypeError        = "error"
	TypeTripOffer    = "trip_offer"
	TypeOfferTaken   = "offer_taken"
	TypeOfferExpired = "offer_expired"

	TypeWaiting        = "waiting"
	TypeDriverMatched  = "driver_matched"
	TypeOfferRejected  = "offer_rejected"
	TypeNoDrivers      = "no_drivers"
	TypeRetry          = "retry"
)

const (
	MESSAGE_INVALID_LAT_LONG           = "invalid lat lng"
	MESSAGE_LOCATION_STORE_UNAVAILABLE = "location store unavailable"
	MESSAGE_FAILED_SAVE_LOC            = "failed to save location"
	MESSAGE_UNKNOWN_MESSAGE_TYPE       = "unknown message type"
	MESSAGE_RETRY_UNAVAILABLE          = "retry unavailable"
)

type ClientMessage struct {
	Type string  `json:"type"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type TripOfferPayload struct {
	TransactionID string     `json:"transaction_id"`
	CustomerName  string     `json:"customer_name"`
	Pickup        [2]float64 `json:"pickup"`
	Destination   [2]float64 `json:"destination"`
	DistanceM     int        `json:"distance_m"`
	TotalFare     int        `json:"total_fare"`
	ExpiresInSec  int        `json:"expires_in_sec"`
}

type MatchedDriverPayload struct {
	UserID        string `json:"user_id"`
	DriverID      string `json:"driver_id"`
	Name          string `json:"name"`
	PhoneNumber   string `json:"phone_number"`
	VehicleID     string `json:"vehicle_id"`
	VehicleName   string `json:"vehicle_name"`
	LicenseNumber string `json:"license_number"`
	VehicleType   string `json:"vehicle_type"`
}

type ServerMessage struct {
	Type          string                `json:"type"`
	Message       string                `json:"message,omitempty"`
	TransactionID string                `json:"transaction_id,omitempty"`
	ExpiresInSec  int                   `json:"expires_in_sec,omitempty"`
	Offer         *TripOfferPayload     `json:"offer,omitempty"`
	MatchedDriver *MatchedDriverPayload `json:"matched_driver,omitempty"`
}
