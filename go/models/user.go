package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email     string         `gorm:"size:100" json:"email,omitempty"`
	Password  string         `gorm:"not null" json:"-"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
}

func (User) TableName() string {
	return "users"
}
