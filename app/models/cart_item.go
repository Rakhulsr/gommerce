package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CartItem struct {
	ID         string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	Cart       *Cart
	CartID     string `gorm:"size:36;index"`
	Product    *Product
	ProductID  string `gorm:"size:36;index"`
	Qty        int
	BasePrice  decimal.Decimal `gorm:"type:decimal(16,2);"`
	BaseTotal  decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxAmount  decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxPercent decimal.Decimal `gorm:"type:decimal(10,2);"`
	GrandTotal decimal.Decimal `gorm:"type:decimal(16,2);"`
	SubTotal   decimal.Decimal `gorm:"type:decimal(16,2);"`
	TotalPrice decimal.Decimal `gorm:"type:decimal(16,2);"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
