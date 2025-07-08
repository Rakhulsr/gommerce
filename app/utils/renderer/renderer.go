package renderer

import (
	"html/template"

	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

func New() *render.Render {
	return render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html"},
		Funcs: []template.FuncMap{
			{
				"until": func(count int) []int {
					items := make([]int, count)
					for i := 0; i < count; i++ {
						items[i] = i
					}
					return items
				},
				"add": func(a, b int) int { return a + b },
				"sub": func(a, b int) int { return a - b },
				"min": func(a, b int) int {
					if a < b {
						return a
					}
					return b
				},
				"max": func(a, b int) int {
					if a > b {
						return a
					}
					return b
				},
				"rupiah": func(d decimal.Decimal) string {
					ac := accounting.Accounting{
						Symbol:    "Rp",
						Precision: 0,
						Thousand:  ".",
						Decimal:   ",",
					}
					f, _ := d.Float64()
					return ac.FormatMoney(f)
				},
			},
		},
	})
}
