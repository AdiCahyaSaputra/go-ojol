package dto

const (
	TypeStandby   = "standby"
	TypeLocation  = "location"
	TypeStandbyOK = "standby_ok"
	TypeError     = "error"
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
