package calc

import "github.com/shopspring/decimal"

func GetTaxPercent() decimal.Decimal {
	var taxPercent = decimal.NewFromInt(12)

	return taxPercent
}

func CalculateTax(baseTotal decimal.Decimal) decimal.Decimal {

	taxPercent := GetTaxPercent()

	return baseTotal.Mul(taxPercent).Div(decimal.NewFromInt(100))

}

func CalculateGrandTotal(baseTotal, taxAmount, discountAmount decimal.Decimal) decimal.Decimal {
	return baseTotal.Add(taxAmount).Sub(discountAmount)
}
