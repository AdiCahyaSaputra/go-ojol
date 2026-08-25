package dto

const (
	TypeStandby      = "standby"
	TypeLocation     = "location"
	TypeStandbyOK    = "standby_ok"
	TypeError        = "error"
	TypeTripOffer    = "trip_offer"
	TypeOfferTaken   = "offer_taken"
	TypeOfferExpired = "offer_expired"
)

const (
	MESSAGE_INVALID_LAT_LONG           = "invalid lat lng"
	MESSAGE_LOCATION_STORE_UNAVAILABLE = "location store unavailable"
	MESSAGE_FAILED_SAVE_LOC            = "failed to save location"
	MESSAGE_UNKNOWN_MESSAGE_TYPE       = "unknown message type"
)

type ClientMessage struct {
	Type string  `json:"type"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type TripOfferPayload struct {
	TransactionID  string     `json:"transaction_id"`
	CustomerName   string     `json:"customer_name"`
	Pickup         [2]float64 `json:"pickup"`
	Destination    [2]float64 `json:"destination"`
	DistanceM      int        `json:"distance_m"`
	TotalFare      int        `json:"total_fare"`
	ExpiresInSec   int        `json:"expires_in_sec"`
}

type ServerMessage struct {
	Type          string            `json:"type"`
	Message       string            `json:"message,omitempty"`
	TransactionID string            `json:"transaction_id,omitempty"`
	Offer         *TripOfferPayload `json:"offer,omitempty"`
}
