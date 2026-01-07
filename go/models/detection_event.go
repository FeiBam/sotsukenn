package models

import (
	"time"

	"gorm.io/gorm"
)

// DetectionEvent 存储所有Frigate检测事件
type DetectionEvent struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 事件基本信息
	EventID string `gorm:"type:varchar(100);uniqueIndex;not null" json:"event_id"` // Frigate事件ID
	Camera  string `gorm:"type:varchar(100);index;not null" json:"camera"`         // 摄像头名称
	Label   string `gorm:"type:varchar(50);index;not null" json:"label"`           // 检测类型 (person, car等)
	SubLabel string `gorm:"type:varchar(100);index" json:"sub_label,omitempty"`    // Re-ID识别结果

	// 时间信息
	StartTime float64 `gorm:"index;not null" json:"start_time"`    // 事件开始时间(Unix时间戳)
	EndTime   *float64 `json:"end_time,omitempty"`                 // 事件结束时间
	IsCurrent bool    `gorm:"index;default:false" json:"is_current"` // 是否为最后事件

	// 其他信息
	TopScore   float64 `json:"top_score,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Active     bool    `json:"active"`
	Stationary bool    `json:"stationary"`
}

func (DetectionEvent) TableName() string {
	return "detection_events"
}
