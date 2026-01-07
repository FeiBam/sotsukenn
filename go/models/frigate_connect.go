package models

import (
	"time"

	"gorm.io/gorm"
)

// FrigateConnect stores the mapping between local users and Frigate API tokens
type FrigateConnect struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Foreign key to local user
	UserID          uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Frigate configuration
	FrigateURL      string         `gorm:"size:255;not null" json:"frigate_url"`
	FrigateUsername string         `gorm:"size:100" json:"frigate_username,omitempty"`

	// Token management (cookie-based authentication)
	TokenCookie     string         `gorm:"type:text" json:"-"`
	LastVerifiedAt  *time.Time     `json:"last_verified_at,omitempty"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
}

func (FrigateConnect) TableName() string {
	return "frigate_connects"
}

// ShouldVerifyToken checks if token needs verification (not verified in last hour)
func (fc *FrigateConnect) ShouldVerifyToken() bool {
	if fc.LastVerifiedAt == nil {
		return true
	}
	return time.Now().Sub(*fc.LastVerifiedAt) > time.Hour
}
