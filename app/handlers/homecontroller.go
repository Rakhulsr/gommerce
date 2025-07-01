package handlers

import (
	"net/http"

	"github.com/unrolled/render"
)

func Home(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html"},
	})

	_ = render.HTML(w, http.StatusOK, "home", map[string]interface{}{
		"title": "Home title",
		"body":  "Home Description",
	})

}
