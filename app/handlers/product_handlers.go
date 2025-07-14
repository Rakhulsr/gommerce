package handlers

import (
	"net/http"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type ProductHandler struct {
	repo         repositories.ProductRepositoryImpl
	categoryRepo repositories.CategoryRepositoryImpl
	render       *render.Render
}

func NewProductHandler(p repositories.ProductRepositoryImpl, c repositories.CategoryRepositoryImpl, r *render.Render) *ProductHandler {
	return &ProductHandler{p, c, r}
}

type ProductDetailPageData struct {
	BaseData      other.BasePageData
	Product       models.Product
	Price         float64
	Breadcrumbs   []breadcrumb.Breadcrumb
	MessageStatus string
	Message       string
}

type ProductListPageData struct {
	BaseData        other.BasePageData
	Title           string
	Products        []models.Product
	Categories      []models.Category
	CurrentPage     int
	TotalPages      int
	CategorySlug    string
	SearchQuery     string
	Breadcrumbs     []breadcrumb.Breadcrumb
	CurrentCategory *models.Category
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
	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Produk", URL: "/products"},
	}

	var currentCategory models.Category

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

	dataMap := helpers.GetBaseData(r, map[string]interface{}{
		"title":       "Produk Kami",
		"products":    products,
		"categories":  categories,
		"current":     page,
		"totalPages":  int((total + int64(limit) - 1) / int64(limit)),
		"category":    slug,
		"searchQuery": query,
		"Breadcrumbs": breadcrumbs,
		"IsAuthPage":  false,
	})

	if currentCategory.ID != "" {
		dataMap["currentCategory"] = currentCategory
	}

	datas := helpers.GetBaseData(r, dataMap)

	_ = h.render.HTML(w, http.StatusOK, "products", datas)

}

func (h *ProductHandler) ProductDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productSlug := vars["slug"]

	if productSlug == "" {
		http.NotFound(w, r)
		return
	}

	product, err := h.repo.GetBySlug(r.Context(), productSlug)
	if err != nil {
		http.Error(w, "Gagal mengambil data produk", http.StatusInternalServerError)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Produk", URL: "/products"},
	}

	if len(product.Categories) > 0 {
		mainCategory := product.Categories[0]
		breadcrumbs = append(breadcrumbs, breadcrumb.Breadcrumb{
			Name: mainCategory.Name,
			URL:  "/products?category=" + mainCategory.Slug,
		})
	}

	breadcrumbs = append(breadcrumbs, breadcrumb.Breadcrumb{Name: product.Name, URL: "/products/" + product.Slug})

	priceFloat, _ := product.Price.Float64()

	dataMap := map[string]interface{}{
		"title":       product.Name,
		"product":     *product,
		"price":       priceFloat,
		"Breadcrumbs": breadcrumbs,
	}

	data := helpers.GetBaseData(r, dataMap)

	_ = h.render.HTML(w, http.StatusOK, "product", data)
}
