package helpers

import (
	"strings"

	"github.com/shopspring/decimal"
)

func FormatRupiah(amount interface{}) string {
	var decAmount decimal.Decimal
	switch v := amount.(type) {
	case decimal.Decimal:
		decAmount = v
	case float64:
		decAmount = decimal.NewFromFloat(v)
	case int:
		decAmount = decimal.NewFromInt(int64(v))
	case int64:
		decAmount = decimal.NewFromInt(v)
	case string:

		parsed, err := decimal.NewFromString(v)
		if err != nil {
			return "Rp 0"
		}
		decAmount = parsed
	default:
		return "Rp 0"
	}

	str := decAmount.StringFixed(0)

	n := len(str)
	if n <= 3 {
		return "Rp " + str
	}
	var b strings.Builder
	for i, char := range str {
		b.WriteRune(char)
		if (n-1-i)%3 == 0 && i != n-1 {
			b.WriteRune('.')
		}
	}
	return "Rp " + b.String()
}
