package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Cart struct {
	ID              string          `gorm:"size:36;not null;uniqueIndex;primary_key"`
	UserID          string          `gorm:"size:36;index"`
	User            User            `gorm:"foreignKey:UserID"`
	CartItems       []CartItem      `gorm:"foreignKey:CartID"`
	BaseTotalPrice  decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxAmount       decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxPercent      decimal.Decimal `gorm:"type:decimal(10,2);"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2);"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2);"`
	GrandTotal      decimal.Decimal `gorm:"type:decimal(16,2);"`
	TotalWeight     decimal.Decimal `gorm:"type:decimal(16,2);default:0.00"`
	ShippingCost    decimal.Decimal `gorm:"type:decimal(16,2);"`
	TotalItems      int             `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (c *Cart) CalculateTotals(defaultTaxPercent decimal.Decimal) {
	c.BaseTotalPrice = decimal.Zero
	totalWeightDecimal := decimal.Zero
	c.TotalItems = 0

	for _, item := range c.CartItems {

		c.BaseTotalPrice = c.BaseTotalPrice.Add(item.Subtotal)

		if item.Product != nil {
			totalWeightDecimal = totalWeightDecimal.Add(item.Product.Weight.Mul(decimal.NewFromInt(int64(item.Qty))))
		}

		c.TotalItems += item.Qty
	}

	c.TotalWeight = totalWeightDecimal

	cartLevelDiscountAmount := decimal.Zero
	if c.DiscountPercent.GreaterThan(decimal.Zero) {

		discountFactor := c.DiscountPercent.Div(decimal.NewFromInt(100))
		cartLevelDiscountAmount = c.BaseTotalPrice.Mul(discountFactor)
	}

	c.DiscountAmount = cartLevelDiscountAmount

	priceAfterCartDiscount := c.BaseTotalPrice.Sub(c.DiscountAmount)
	if priceAfterCartDiscount.LessThan(decimal.Zero) {
		priceAfterCartDiscount = decimal.Zero
	}

	c.TaxPercent = defaultTaxPercent
	if c.TaxPercent.GreaterThan(decimal.Zero) {
		taxFactor := c.TaxPercent.Div(decimal.NewFromInt(100))
		c.TaxAmount = priceAfterCartDiscount.Mul(taxFactor)
	} else {
		c.TaxAmount = decimal.Zero
	}

	c.GrandTotal = priceAfterCartDiscount.Add(c.TaxAmount).Add(c.ShippingCost)
}
