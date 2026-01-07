package types

// MQTTConfig holds MQTT configuration
type MQTTConfig struct {
	BrokerURL  string
	BrokerPort string
	ClientID   string
	Username   string
	Password   string
	Topic      string
}

// FrigateLoginRequest represents Frigate /api/login request
type FrigateLoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}
