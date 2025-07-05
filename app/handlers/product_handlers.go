package handlers

import (
	"net/http"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type ProductHandler struct {
	repo         repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	render       *render.Render
}

func NewProductHandler(p repositories.ProductRepository, c repositories.CategoryRepository, r *render.Render) *ProductHandler {
	return &ProductHandler{p, c, r}
}

func (h *ProductHandler) Products(w http.ResponseWriter, r *http.Request) {

	slug := r.URL.Query().Get("category")
	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 9
	offset := (page - 1) * limit

	var (
		products []models.Product
		err      error
		total    int64
	)

	switch {
	case query != "":
		products, total, err = h.repo.SearchProductsPaginated(r.Context(), query, limit, offset)
	case slug != "":
		products, total, err = h.repo.GetByCategorySlugPaginated(r.Context(), slug, limit, offset)
	default:
		products, total, err = h.repo.GetPaginated(r.Context(), limit, offset)
	}

	if err != nil {
		http.Error(w, "Gagal mengambil data produk", http.StatusInternalServerError)
		return
	}

	categories, err := h.categoryRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Gagal mengambil kategori", http.StatusInternalServerError)
		return
	}

	_ = h.render.HTML(w, http.StatusOK, "products", map[string]interface{}{
		"title":       "Produk Kami",
		"products":    products,
		"categories":  categories,
		"current":     page,
		"totalPages":  int((total + int64(limit) - 1) / int64(limit)),
		"category":    slug,
		"searchQuery": query,
	})
}

func (h *ProductHandler) ProductDetail(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	if vars["slug"] == "" {
		return
	}

	product, err := h.repo.GetBySlug(r.Context(), vars["slug"])
	if err != nil {
		http.Error(w, "Gagal mengambil data produk", http.StatusInternalServerError)
		return
	}

	_ = h.render.HTML(w, http.StatusOK, "product", map[string]interface{}{
		"title":   product.Name,
		"product": product,
	})

}
