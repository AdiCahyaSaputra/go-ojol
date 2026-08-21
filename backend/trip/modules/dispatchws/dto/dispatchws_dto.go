package dto

const (
	TypeStandby   = "standby"
	TypeLocation  = "location"
	TypeStandbyOK = "standby_ok"
	TypeError     = "error"
)

const (
	MESSAGE_INVALID_LAT_LONG = "invalid lat lng"
	MESSAGE_LOCATION_STORE_UNAVAILABLE = "location store unavailable"
	MESSAGE_FAILED_SAVE_LOC = "failed to save location"
	MESSAGE_UNKNOWN_MESSAGE_TYPE = "unknown message type"
)

type ClientMessage struct {
	Type string  `json:"type"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type ServerMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
