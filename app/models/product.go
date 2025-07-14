package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	ID              string          `gorm:"size:36;not null;uniqueIndex;primary_key"`
	UserID          string          `gorm:"size:36;index"`
	User            User            `gorm:"foreignKey:UserID"`
	Name            string          `gorm:"size:255;not null"`
	Slug            string          `gorm:"size:255;not null;uniqueIndex"`
	Description     string          `gorm:"type:text"`
	Sku             string          `gorm:"size:100;uniqueIndex"`
	Price           decimal.Decimal `gorm:"type:decimal(16,2);not null"`
	Stock           int             `gorm:"not null"`
	Weight          decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2);default:0.00"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2);default:0.00"`
	Categories      []Category      `gorm:"many2many:product_categories;"`
	ProductImages   []ProductImage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type ProductCategory struct {
	ProductID  string `gorm:"size:36;primaryKey"`
	CategoryID string `gorm:"size:36;primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
