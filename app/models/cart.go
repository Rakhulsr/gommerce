package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Cart struct {
	ID              string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	UserID          string `gorm:"size:36;index"`
	User            User   `gorm:"foreignKey:UserID"`
	CartItems       []CartItem
	BaseTotalPrice  decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxAmount       decimal.Decimal `gorm:"type:decimal(16,2);"`
	TaxPercent      decimal.Decimal `gorm:"type:decimal(10,2);"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2);"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2);"`
	GrandTotal      decimal.Decimal `gorm:"type:decimal(16,2);"`
	TotalWeight     int             `gorm:"-"`
	ShippingCost    decimal.Decimal `gorm:"type:decimal(16,2);"`
	TotalItems      int             `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (c *Cart) CalculateTotals() {
	c.BaseTotalPrice = decimal.Zero
	totalWeightDecimal := decimal.Zero
	c.TotalItems = 0

	for _, item := range c.CartItems {
		if item.Product != nil {
			itemTotalPrice := item.Product.Price.Mul(decimal.NewFromInt(int64(item.Qty)))
			c.BaseTotalPrice = c.BaseTotalPrice.Add(itemTotalPrice)

			totalWeightDecimal = totalWeightDecimal.Add(item.Product.Weight.Mul(decimal.NewFromInt(int64(item.Qty))))

			c.TotalItems += item.Qty
		}
	}

	c.TotalWeight = int(totalWeightDecimal.IntPart())

	if c.DiscountPercent.GreaterThan(decimal.Zero) {
		discountFactor := c.DiscountPercent.Div(decimal.NewFromInt(100))
		c.DiscountAmount = c.BaseTotalPrice.Mul(discountFactor)
	} else {
		c.DiscountAmount = decimal.Zero
	}

	priceAfterDiscount := c.BaseTotalPrice.Sub(c.DiscountAmount)

	if c.TaxPercent.GreaterThan(decimal.Zero) {
		taxFactor := c.TaxPercent.Div(decimal.NewFromInt(100))
		c.TaxAmount = priceAfterDiscount.Mul(taxFactor)
	} else {
		c.TaxAmount = decimal.Zero
	}

	c.GrandTotal = priceAfterDiscount.Add(c.TaxAmount)
}
