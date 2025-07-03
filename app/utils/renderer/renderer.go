// renderer/renderer.go
package renderer

import (
	"html/template"

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
			},
		},
	})
}
