package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	OrderStatusPending    = 1
	OrderStatusProcessing = 2
	OrderStatusShipped    = 3
	OrderStatusCompleted  = 4
	OrderStatusCancelled  = 5
	OrderStatusRefunded   = 6
	OrderStatusFailed     = 7
)

type Order struct {
	ID        string    `gorm:"size:36;not null;uniqueIndex;primary_key"`
	UserID    string    `gorm:"size:36;index"`
	User      User      `gorm:"foreignKey:UserID"`
	OrderCode string    `gorm:"type:varchar(255);unique;not null" json:"order_code"`
	OrderDate time.Time `gorm:"not null" json:"order_date"`

	OrderItems           []OrderItem
	BaseTotalPrice       decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxAmount            decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxPercent           decimal.Decimal `gorm:"type:decimal(10,2);"`
	DiscountAmount       decimal.Decimal `gorm:"type:decimal(16,2);"`
	DiscountPercent      decimal.Decimal `gorm:"type:decimal(10,2);"`
	ShippingCost         decimal.Decimal `gorm:"type:decimal(16,2);"`
	GrandTotal           decimal.Decimal `gorm:"type:decimal(16,2);"`
	ShippingAddress      string          `gorm:"type:text"`
	ShippingService      string          `gorm:"size:255"`
	ShippingServiceCode  string          `gorm:"type:varchar(50);not null" json:"shipping_service_code"`
	ShippingServiceName  string          `gorm:"type:varchar(255);not null" json:"shipping_service_name"`
	ShippingTrackingCode string          `gorm:"size:255"`

	AddressID string  `gorm:"type:varchar(255);not null" json:"address_id"`
	Address   Address `gorm:"foreignKey:AddressID;references:ID"`

	MidtransTransactionID string `gorm:"size:255;index"`
	MidtransPaymentURL    string `gorm:"type:text"`
	PaymentStatus         string `gorm:"size:100"`

	Status int `gorm:"default:1"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return
}
