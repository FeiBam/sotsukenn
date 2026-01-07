package models

// FrigateEvent represents an event from Frigate MQTT
type FrigateEvent struct {
	Type   string    `json:"type"`   // new, update, end
	Before EventData `json:"before"`
	After  EventData `json:"after"`
}

// EventData contains the event details
type EventData struct {
	ID                       string              `json:"id"`
	Camera                   string              `json:"camera"`
	FrameTime                float64             `json:"frame_time,omitempty"`
	Snapshot                 *Snapshot           `json:"snapshot,omitempty"`
	Label                    string              `json:"label"`
	SubLabel                 interface{}         `json:"sub_label,omitempty"` // Can be null, string, or array
	TopScore                 float64             `json:"top_score,omitempty"`
	FalsePositive            bool                `json:"false_positive"`
	StartTime                float64             `json:"start_time"`
	EndTime                  *float64            `json:"end_time,omitempty"`
	Score                    float64             `json:"score"`
	Box                      []int               `json:"box,omitempty"`
	Area                     int                 `json:"area,omitempty"`
	Ratio                    float64             `json:"ratio,omitempty"`
	Region                   []int               `json:"region,omitempty"`
	CurrentZones             []string            `json:"current_zones,omitempty"`
	EnteredZones             []string            `json:"entered_zones,omitempty"`
	Thumbnail                string              `json:"thumbnail,omitempty"`
	HasSnapshot              bool                `json:"has_snapshot"`
	HasClip                  bool                `json:"has_clip"`
	Active                   bool                `json:"active"`
	Stationary               bool                `json:"stationary"`
	MotionlessCount          int                 `json:"motionless_count,omitempty"`
	PositionChanges          int                 `json:"position_changes,omitempty"`
	Attributes               *AttributeSummary   `json:"attributes,omitempty"`
	CurrentAttributes        []AttributeDetail   `json:"current_attributes,omitempty"`
	CurrentEstimatedSpeed    float64             `json:"current_estimated_speed,omitempty"`
	VelocityAngle            int                 `json:"velocity_angle,omitempty"`
	RecognizedLicensePlate   string              `json:"recognized_license_plate,omitempty"`
	RecognizedLicensePlateScore float64          `json:"recognized_license_plate_score,omitempty"`
}

// Snapshot represents snapshot data
type Snapshot struct {
	FrameTime float64  `json:"frame_time,omitempty"`
	Box       []int    `json:"box,omitempty"`
	Area      int      `json:"area,omitempty"`
	Region    []int    `json:"region,omitempty"`
	Score     float64  `json:"score,omitempty"`
	Attributes []AttributeDetail `json:"attributes,omitempty"`
}

// AttributeSummary contains summary of attributes
type AttributeSummary struct {
	// Can be string->float or string->AttributeDetail depending on context
}

// AttributeDetail contains detailed attribute information
type AttributeDetail struct {
	Label string  `json:"label,omitempty"`
	Box   []int   `json:"box,omitempty"`
	Score float64 `json:"score,omitempty"`
}

// EventType constants
const (
	EventTypeNew    = "new"
	EventTypeUpdate = "update"
	EventTypeEnd    = "end"
)
