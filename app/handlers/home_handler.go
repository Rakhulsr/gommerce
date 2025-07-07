package handlers

import (
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/unrolled/render"
)

type HomeHandler struct {
	render       *render.Render
	categoryRepo repositories.CategoryRepository
	productRepo  repositories.ProductRepository
}

func NewHomeHandler(r *render.Render, c repositories.CategoryRepository, p repositories.ProductRepository) *HomeHandler {
	return &HomeHandler{
		render:       r,
		categoryRepo: c,
		productRepo:  p,
	}
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Gagal mengambil kategori", http.StatusInternalServerError)
		return
	}

	products, err := h.productRepo.GetFeaturedProducts(r.Context(), 8)
	if err != nil {
		http.Error(w, "Gagal mengambil Featured Product", http.StatusInternalServerError)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
	}

	datas := helpers.GetBaseData(r, map[string]interface{}{
		"title":       "Beranda",
		"categories":  categories,
		"category":    "",
		"current":     1,
		"totalPages":  1,
		"featured":    products,
		"breadcrumbs": breadcrumbs,
	})

	_ = h.render.HTML(w, http.StatusOK, "home", datas)
}
