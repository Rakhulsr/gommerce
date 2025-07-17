package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderCustomer struct {
	gorm.Model
	ID        string `gorm:"type:char(36);primaryKey"`
	FirstName string `gorm:"type:varchar(255);not null"`
	LastName  string `gorm:"type:varchar(255);null"`
	Email     string `gorm:"type:varchar(255);not null"`
	Phone     string `gorm:"type:varchar(20);not null"`
	Address1  string `gorm:"type:varchar(255);not null"`
	Address2  string `gorm:"type:varchar(255);null"`
	// OLD: CityID       string `gorm:"type:varchar(10);not null"`     // DIHAPUS
	// OLD: ProvinceID   string `gorm:"type:varchar(10);not null"` // DIHAPUS
	LocationID   string `gorm:"type:varchar(20);not null"`  // NEW: Untuk menyimpan Subdistrict ID dari Komerce API
	LocationName string `gorm:"type:varchar(255);not null"` // NEW: Untuk menyimpan nama lengkap lokasi (Kecamatan, Kota, Provinsi)
	PostCode     string `gorm:"type:varchar(10);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
