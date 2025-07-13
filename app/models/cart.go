package models

import "github.com/shopspring/decimal"

type Cart struct {
	ID              string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	UserID          string `gorm:"size:36;index"`
	User            User   `gorm:"foreignKey:UserID"`
	CartItems       []CartItem
	BaseTotalPrice  decimal.Decimal `gorm:"decimal(16,2);"`
	TaxAmount       decimal.Decimal `gorm:"decimal(16,2);"`
	TaxPercent      decimal.Decimal `gorm:"decimal(10,2);"`
	DiscountAmount  decimal.Decimal `gorm:"decimal(16,2);"`
	DiscountPercent decimal.Decimal `gorm:"decimal(10,2);"`
	GrandTotal      decimal.Decimal `gorm:"decimal(16,2);"`
	TotalWeight     int             `gorm:"-"`
}
