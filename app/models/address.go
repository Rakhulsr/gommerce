package models

import (
	"time"

	"gorm.io/gorm"
)

type Address struct {
	ID           string `gorm:"size:36;not null;uniqueIndex;primary_key" json:"id"`
	UserID       string `gorm:"type:uuid;not null" json:"user_id"`
	User         User   `gorm:"foreignKey:UserID"`
	Name         string
	IsPrimary    bool   `gorm:"default:false"`
	Address1     string `gorm:"type:varchar(255);not null" json:"address1"`
	Address2     string `gorm:"type:varchar(255)" json:"address2"`
	LocationID   string `gorm:"type:varchar(255);not null" json:"location_id"`
	LocationName string `gorm:"type:varchar(255);not null" json:"location_name"`
	PostCode     string `gorm:"type:varchar(10);not null" json:"post_code"`
	Phone        string `gorm:"type:varchar(20);not null" json:"phone"`
	Email        string `gorm:"type:varchar(100)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
