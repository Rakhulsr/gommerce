package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                    string     `gorm:"size:36;not null;uniqueIndex;primary_key" json:"id"`
	FirstName             string     `gorm:"size:100;not null"`
	LastName              string     `gorm:"size:100;not null"`
	Address               []Address  `gorm:"foreignKey:UserID"`
	Email                 string     `gorm:"size:100;not null;uniqueIndex"`
	Phone                 string     `gorm:"size:20"`
	Password              string     `gorm:"size:255;not null"`
	Role                  string     `gorm:"size:20;default:'customer';not null"`
	RememberTokenSelector *string    `gorm:"size:64;uniqueIndex;null"`
	RememberTokenHash     string     `gorm:"size:255;null"`
	PasswordResetToken    *string    `gorm:"size:255;uniqueIndex;null"`
	PasswordResetExpires  *time.Time `gorm:"null"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt
}

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
)
