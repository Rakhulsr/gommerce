package calc

import "github.com/shopspring/decimal"

func CalculateDiscount(baseTotal, discountPercent decimal.Decimal) decimal.Decimal {
	return baseTotal.Mul(discountPercent).Div(decimal.NewFromInt(100))
}
