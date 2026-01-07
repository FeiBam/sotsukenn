package types

import "time"

// FrigateStatus represents Frigate status information
type FrigateStatus struct {
	IsOnline      bool      `json:"is_online"`
	ResponseTime  int64     `json:"response_time_ms"`
	LastError     string    `json:"last_error,omitempty"`
	LastCheckTime time.Time `json:"last_check_time"`
}

// CameraStatus represents camera status information
type CameraStatus struct {
	OnlineCount  int      `json:"online_count"`
	OfflineCount int      `json:"offline_count"`
	TotalCount   int      `json:"total_count"`
	OnlineList   []string `json:"online_list,omitempty"`
	OfflineList  []string `json:"offline_list,omitempty"`
}
