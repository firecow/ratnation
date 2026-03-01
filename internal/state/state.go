package state

type State struct {
	Revision int            `json:"revision"`
	Services []StateService `json:"services"`
	Kings    []StateKing    `json:"kings"`
	Lings    []StateLing    `json:"lings"`
}

type StateKing struct {
	BindPort     int    `json:"bind_port"`
	Host         string `json:"host"`
	Ports        string `json:"ports"`
	ShuttingDown bool   `json:"shutting_down"`
	Beat         int64  `json:"beat"`
	Location     string `json:"location"`
	CertPEM      string `json:"cert_pem,omitempty"`
}

type StateLing struct {
	LingID       string `json:"ling_id"`
	ShuttingDown bool   `json:"shutting_down"`
	Beat         int64  `json:"beat"`
}

type StateService struct {
	Name              string  `json:"name"`
	Token             string  `json:"token"`
	ServiceID         string  `json:"service_id"`
	LingID            string  `json:"ling_id"`
	PreferredLocation string  `json:"preferred_location"`
	LingReady         bool    `json:"ling_ready"`
	KingReady         bool    `json:"king_ready"`
	Host              *string `json:"host"`
	BindPort          *int    `json:"bind_port"`
	RemotePort        *int    `json:"remote_port"`
}
