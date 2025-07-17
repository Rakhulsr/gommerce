package models

import (
	"time"

	"gorm.io/gorm"
)

type Address struct {
	gorm.Model
	ID           string `gorm:"size:36;not null;uniqueIndex;primary_key" json:"id"`
	UserID       string `gorm:"type:uuid;not null"`
	User         User   `gorm:"foreignKey:UserID"` // Relasi ke model User
	Name         string // Nama penerima (opsional, bisa sama dengan nama user)
	IsPrimary    bool   `gorm:"default:false"`
	Address1     string `gorm:"type:text;not null"` // Alamat lengkap (Jl, No. Rumah, RT/RW)
	Address2     string `gorm:"type:text"`          // Detail lain (Blok, Unit, Patokan)
	LocationID   string `gorm:"type:varchar(255)"`  // ID lokasi dari RajaOngkir (jika masih diperlukan untuk shipping, tapi tidak untuk form alamat)
	LocationName string `gorm:"type:text;not null"` // Format: "Kel. Mekar Jaya, Kec. Sukmajaya, Kota Depok, Prov. Jawa Barat"
	PostCode     string `gorm:"type:varchar(10);not null"`
	Phone        string `gorm:"type:varchar(20);not null"`
	Email        string `gorm:"type:varchar(100)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// type Address struct {
// 	ID           string `gorm:"size:36;not null;uniqueIndex;primary_key"`
// 	User         User
// 	UserID       string `gorm:"size:36;index"`
// 	Name         string `gorm:"size:255;not null"`
// 	IsPrimary    bool
// 	CityID       string `gorm:"size:100"`
// 	ProvinceID   string `gorm:"size:100"`
// 	Address1     string `gorm:"size:255"`
// 	Address2     string `gorm:"size:255"`
// 	Phone        string `gorm:"size:100"`
// 	Email        string `gorm:"size:100"`
// 	PostCode     string `gorm:"size:100"`
// 	ProvinceName string `gorm:"-" json:"province_name,omitempty"`
// 	CityName     string `gorm:"-" json:"city_name,omitempty"`
// 	CreatedAt    time.Time
// 	UpdatedAt    time.Time
// }
