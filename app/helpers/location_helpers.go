package helpers

import (
	"fmt"
	"net/url"
	"strings"
)

func FormatCurrency(amount float64) string {
	return fmt.Sprintf("Rp %.2f", amount)
}

func URLQueryEscape(s string) string {
	return url.QueryEscape(s)
}

func Add(a, b int) int { return a + b }
func Sub(a, b int) int { return a - b }
func Mul(a, b int) int { return a * b }
func Div(a, b int) int { return a / b }
func Mod(a, b int) int { return a % b }

func Eq(a, b interface{}) bool { return a == b }
func Ne(a, b interface{}) bool { return a != b }
func Lt(a, b interface{}) bool {
	switch a := a.(type) {
	case int:
		return a < b.(int)
	case float64:
		return a < b.(float64)
	default:
		return false
	}
}
func Le(a, b interface{}) bool {
	switch a := a.(type) {
	case int:
		return a <= b.(int)
	case float64:
		return a <= b.(float64)
	default:
		return false
	}
}
func Gt(a, b interface{}) bool {
	switch a := a.(type) {
	case int:
		return a > b.(int)
	case float64:
		return a > b.(float64)
	default:
		return false
	}
}
func Ge(a, b interface{}) bool {
	switch a := a.(type) {
	case int:
		return a >= b.(int)
	case float64:
		return a >= b.(float64)
	default:
		return false
	}
}

func ExtractProvinceFromLocationName(locationName string) string {
	parts := strings.Split(locationName, ", Prov. ")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func ExtractCityFromLocationName(locationName string) string {
	parts := strings.Split(locationName, ", Kota ")
	if len(parts) > 1 {
		cityAndProvince := strings.Split(parts[1], ", Prov. ")
		if len(cityAndProvince) > 0 {
			return strings.TrimSpace(cityAndProvince[0])
		}
	}
	return ""
}

func ExtractDistrictFromLocationName(locationName string) string {
	parts := strings.Split(locationName, ", Kec. ")
	if len(parts) > 1 {
		districtAndCity := strings.Split(parts[1], ", Kota ")
		if len(districtAndCity) > 0 {
			return strings.TrimSpace(districtAndCity[0])
		}
	}
	return ""
}

func ExtractSubdistrictFromLocationName(locationName string) string {
	parts := strings.Split(locationName, "Kel. ")
	if len(parts) > 1 {
		subdistrictAndRest := strings.Split(parts[1], ", Kec. ")
		if len(subdistrictAndRest) > 0 {
			return strings.TrimSpace(subdistrictAndRest[0])
		}
	}
	return ""
}
