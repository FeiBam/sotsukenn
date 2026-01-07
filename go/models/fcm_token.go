package models

import (
	"time"

	"gorm.io/gorm"
)

// FCMToken represents a Firebase Cloud Messaging token for push notifications
type FCMToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Foreign key to local user
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	User      User   `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// FCM token information
	Token      string `gorm:"size:500;not null;uniqueIndex" json:"token"`
	DeviceName string `gorm:"size:100" json:"device_name"` // e.g., "iPhone 13", "Pixel 6"
	IsActive   bool   `gorm:"default:true" json:"is_active"`
}

func (FCMToken) TableName() string {
	return "fcm_tokens"
}
