package handlers

import (
	"net/http"

	"github.com/unrolled/render"
)

func Products(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html"},
	})

	_ = render.HTML(w, http.StatusOK, "products", map[string]interface{}{
		"title": "Daftar Produk",
	})
}
