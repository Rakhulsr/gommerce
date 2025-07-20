package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OrderItem struct {
	ID              string          `gorm:"primaryKey;type:varchar(255);not null;uniqueIndex" json:"id"`
	OrderID         string          `gorm:"type:varchar(255);not null;index" json:"order_id"`
	Order           Order           `gorm:"foreignKey:OrderID;references:ID"`
	ProductID       string          `gorm:"type:varchar(255);not null;index" json:"product_id"`
	Product         Product         `gorm:"foreignKey:ProductID;references:ID"`
	ProductName     string          `gorm:"type:varchar(255);not null" json:"product_name"`
	ProductSku      string          `gorm:"type:varchar(100)" json:"product_sku"`
	Qty             int             `gorm:"not null" json:"qty"`
	Price           decimal.Decimal `gorm:"type:decimal(16,2);not null" json:"price"`
	BaseTotal       decimal.Decimal `gorm:"type:decimal(16,2);not null" json:"base_total"`
	TaxAmount       decimal.Decimal `gorm:"type:decimal(16,2);not null" json:"tax_amount"`
	TaxPercent      decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"tax_percent"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2);not null" json:"discount_amount"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"discount_percent"`
	GrandTotal      decimal.Decimal `gorm:"type:decimal(16,2);not null" json:"grand_total"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if oi.ID == "" {
		oi.ID = uuid.New().String()
	}
	return
}
