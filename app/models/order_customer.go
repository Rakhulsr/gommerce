package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderCustomer struct {
	ID string `gorm:"type:char(36);primaryKey"`

	OrderID string `gorm:"type:varchar(36);not null;uniqueIndex"`
	Order   *Order `gorm:"foreignKey:OrderID;references:ID"`

	FirstName    string `gorm:"type:varchar(255);not null"`
	LastName     string `gorm:"type:varchar(255);null"`
	Email        string `gorm:"type:varchar(255);not null"`
	Phone        string `gorm:"type:varchar(20);not null"`
	Address1     string `gorm:"type:varchar(255);not null"`
	Address2     string `gorm:"type:varchar(255);null"`
	LocationID   string `gorm:"type:varchar(20);not null"`
	LocationName string `gorm:"type:varchar(255);not null"`
	PostCode     string `gorm:"type:varchar(10);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (oc *OrderCustomer) BeforeCreate(tx *gorm.DB) (err error) {
	if oc.ID == "" {
		oc.ID = uuid.New().String()
	}
	return
}
