package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartItem struct {
	ID              string   `gorm:"size:36;not null;uniqueIndex;primary_key" json:"id"`
	Cart            *Cart    `gorm:"foreignKey:CartID"`
	CartID          string   `gorm:"size:36;index"`
	Product         *Product `gorm:"foreignKey:ProductID"`
	ProductID       string   `gorm:"size:36;index"`
	Qty             int
	Price           decimal.Decimal `gorm:"type:decimal(16,2);"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2);"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2);"`
	FinalPriceUnit  decimal.Decimal `gorm:"type:decimal(16,2);"`
	Subtotal        decimal.Decimal `gorm:"type:decimal(16,2);"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (ci *CartItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ci.ID == "" {
		ci.ID = uuid.New().String()
	}
	return
}
