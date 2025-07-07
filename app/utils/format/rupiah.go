package format

import (
	"fmt"
	"strings"
)

func Rupiah(amount float64) string {
	str := fmt.Sprintf("%.0f", amount)
	n := len(str)
	if n <= 3 {
		return "Rp" + str
	}

	var result []string
	for i := n; i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		result = append([]string{str[start:i]}, result...)
	}

	return "Rp" + strings.Join(result, ".")
}
