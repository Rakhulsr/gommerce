package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

type AdminHandler struct {
	render       *render.Render
	validator    *validator.Validate
	productRepo  repositories.ProductRepositoryImpl
	categoryRepo repositories.CategoryRepositoryImpl
	sectionRepo  repositories.SectionRepositoryImpl
	userRepo     repositories.UserRepositoryImpl
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	cartSvc      services.CartService
}

func NewAdminHandler(
	render *render.Render,
	validator *validator.Validate,
	productRepo repositories.ProductRepositoryImpl,
	categoryRepo repositories.CategoryRepositoryImpl,
	sectionRepo repositories.SectionRepositoryImpl,
	userRepo repositories.UserRepositoryImpl,
	cartRepo repositories.CartRepositoryImpl,
	cartItemRepo repositories.CartItemRepositoryImpl,
	cartSvc services.CartService,
) *AdminHandler {
	return &AdminHandler{
		render:       render,
		validator:    validator,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		sectionRepo:  sectionRepo,
		userRepo:     userRepo,
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		cartSvc:      cartSvc,
	}
}

type AdminPageData struct {
	other.BasePageData
	TotalUsers    int
	TotalProducts int
	TotalOrders   int
	RecentOrders  []struct {
		OrderID   string
		Customer  string
		Amount    float64
		Status    string
		OrderDate time.Time
	}
	RecentActivities []struct {
		Activity string
		Time     time.Time
	}
	CurrentGlobalDiscount float64
}

type AdminProductPageData struct {
	other.BasePageData
	Products    []models.Product
	ProductData *ProductForm
	IsEdit      bool
	FormAction  string
	Errors      map[string]string
	Categories  []models.Category
}

type ProductForm struct {
	ID          string
	Name        string `form:"name" validate:"required,min=3,max=100"`
	Description string `form:"description" validate:"required,min=10"`
	SKU         string `form:"sku" validate:"required,alphanum,min=3,max=20"`
	Price       string `form:"price" validate:"required,numeric,min=0"`
	Stock       string `form:"stock" validate:"required,numeric,min=0"`
	Weight      string `form:"weight" validate:"required,numeric,min=0"`
	ImagePath   string `form:"image_path"`
	CategoryID  string `form:"category_id" validate:"required"`
}

type AdminCategoryPageData struct {
	other.BasePageData
	Categories   []models.Category
	CategoryData *CategoryForm
	IsEdit       bool
	FormAction   string
	Errors       map[string]string
	Sections     []models.Section
}

type CategoryForm struct {
	ID        string `form:"id"`
	Name      string `form:"name" validate:"required,min=3,max=100"`
	Slug      string
	ParentID  string `form:"parent_id"`
	SectionID string `form:"section_id"`
}

type AdminUserPageData struct {
	other.BasePageData
	Users      []models.User
	UserData   *UserForm
	IsEdit     bool
	FormAction string
	Errors     map[string]string
}

type UserForm struct {
	ID        string `form:"id"`
	FirstName string `form:"first_name" validate:"required,min=2,max=50"`
	LastName  string `form:"last_name" validate:"min=2,max=50"`
	Email     string `form:"email" validate:"required,email"`
	Password  string `form:"password"`
	Role      string `form:"role" validate:"required,oneof=admin customer"`
}

func (h *AdminHandler) populateBaseDataForAdmin(r *http.Request, pageData interface{}) {
	baseDataMap := helpers.GetBaseData(r, nil)

	var base *other.BasePageData
	switch pd := pageData.(type) {
	case *AdminPageData:
		base = &pd.BasePageData
	case *AdminProductPageData:
		base = &pd.BasePageData
	case *AdminCategoryPageData:
		base = &pd.BasePageData
	case *AdminUserPageData:
		base = &pd.BasePageData
	default:
		log.Printf("populateBaseDataForAdmin: Unknown pageData type: %T", pageData)
		return
	}

	if title, ok := baseDataMap["Title"].(string); ok {
		base.Title = title
	}
	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
		base.IsLoggedIn = isLoggedIn
	}
	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok {
		base.User = user
	}
	if userID, ok := baseDataMap["UserID"].(string); ok {
		base.UserID = userID
	}
	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
		base.CartCount = cartCount
	}
	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
		base.CSRFToken = csrfToken
	}
	if message, ok := baseDataMap["Message"].(string); ok {
		base.Message = message
	}
	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
		base.MessageStatus = messageStatus
	}
	if query, ok := baseDataMap["Query"].(url.Values); ok {
		base.Query = query
	}
	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
		base.Breadcrumbs = breadcrumbs
	}
	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
		base.IsAuthPage = isAuthPage
	}
	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
		base.IsAdminPage = isAdminPage
	}
	if hideAdminWelcomeMessage, ok := baseDataMap["HideAdminWelcomeMessage"].(bool); ok {
		base.HideAdminWelcomeMessage = hideAdminWelcomeMessage
	}
	base.CurrentPath = r.URL.Path
	if strings.HasPrefix(r.URL.Path, "/admin/") {
		base.IsAdminRoute = true
	} else {
		base.IsAdminRoute = false
	}
}

func (h *AdminHandler) applyGlobalDiscount(ctx context.Context, discountPercent float64) error {
	discountDecimal := decimal.NewFromFloat(discountPercent)

	products, err := h.productRepo.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all products for global discount: %w", err)
	}

	for _, product := range products {

		productDiscountAmount := calc.CalculateDiscount(product.Price, discountDecimal)

		if err := h.productRepo.UpdateProductDiscount(ctx, product.ID, discountDecimal, productDiscountAmount); err != nil {
			log.Printf("applyGlobalDiscount: Failed to update discount for product %s: %v", product.ID, err)

		}
	}

	carts, err := h.cartRepo.GetAllCarts(ctx)
	if err != nil {
		log.Printf("applyGlobalDiscount: Failed to get all carts to update summaries after product discount: %v", err)
		return fmt.Errorf("failed to update cart summaries after global product discount: %w", err)
	}

	for _, cart := range carts {

		if err := h.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
			log.Printf("applyGlobalDiscount: Failed to update cart summary for cart %s after product discount: %v", cart.ID, err)

		}
	}

	log.Printf("applyGlobalDiscount: Global discount of %.2f%% applied to all products and cart summaries updated.", discountPercent)
	return nil
}

func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	data := &AdminPageData{}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Dashboard Admin"
	data.IsAuthPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Dashboard", URL: "/admin/dashboard"},
	}

	data.TotalUsers = 1250
	data.TotalProducts = 500
	data.TotalOrders = 780

	data.RecentOrders = []struct {
		OrderID   string
		Customer  string
		Amount    float64
		Status    string
		OrderDate time.Time
	}{
		{OrderID: "ORD001", Customer: "Budi Santoso", Amount: 150000.00, Status: "Pending", OrderDate: time.Now().Add(-1 * time.Hour)},
		{OrderID: "ORD002", Customer: "Siti Aminah", Amount: 250000.00, Status: "Completed", OrderDate: time.Now().Add(-6 * time.Hour)},
		{OrderID: "ORD003", Customer: "Joko Susilo", Amount: 80000.00, Status: "Processing", OrderDate: time.Now().Add(-24 * time.Hour)},
	}

	data.RecentActivities = []struct {
		Activity string
		Time     time.Time
	}{
		{Activity: "Produk baru 'Smartwatch X' ditambahkan.", Time: time.Now().Add(-30 * time.Minute)},
		{Activity: "Pengguna 'budi.s@example.com' mendaftar.", Time: time.Now().Add(-2 * time.Hour)},
		{Activity: "Pesanan ORD002 diselesaikan.", Time: time.Now().Add(-7 * time.Hour)},
	}

	products, err := h.productRepo.GetProducts(r.Context())
	if err == nil && len(products) > 0 {

		data.CurrentGlobalDiscount = products[0].DiscountPercent.InexactFloat64()
	} else {
		data.CurrentGlobalDiscount = 0.0
	}

	h.render.HTML(w, http.StatusOK, "admin/dashboard_index", data)
}

func (h *AdminHandler) ApplyGlobalDiscountPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ApplyGlobalDiscountPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	discountPercentStr := r.PostFormValue("global_discount_percent")
	discountPercent, err := strconv.ParseFloat(discountPercentStr, 64)
	if err != nil || discountPercent < 0 || discountPercent > 100 {
		log.Printf("ApplyGlobalDiscountPost: Persentase diskon tidak valid: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	err = h.applyGlobalDiscount(r.Context(), discountPercent)
	if err != nil {
		log.Printf("ApplyGlobalDiscountPost: Gagal menerapkan diskon global: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *AdminHandler) GetProductsPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminProductPageData{}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Manajemen Produk"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Produk", URL: "/admin/products"},
	}

	products, err := h.productRepo.GetProducts(r.Context())
	if err != nil {
		log.Printf("GetProductsPage: Gagal mengambil daftar produk: %v", err)
		data.Message = "Gagal mengambil daftar produk."
		data.MessageStatus = "error"
	} else {
		data.Products = products
	}

	h.render.HTML(w, http.StatusOK, "admin/products/index", data)
}

func (h *AdminHandler) AddProductPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminProductPageData{
		FormAction:  "/admin/products/add",
		IsEdit:      false,
		ProductData: &ProductForm{},
		Errors:      make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	categories, err := h.categoryRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("AddProductPage: Gagal mengambil kategori: %v", err)
		data.Message = "Gagal memuat kategori."
		data.MessageStatus = "error"
	}
	data.Categories = categories

	data.Title = "Tambah Produk Baru"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Produk", URL: "/admin/products"},
		{Name: "Tambah Baru", URL: "/admin/products/add"},
	}

	h.render.HTML(w, http.StatusOK, "admin/products/form", data)
}

func (h *AdminHandler) AddProductPost(w http.ResponseWriter, r *http.Request) {
	var form ProductForm
	if err := r.ParseForm(); err != nil {
		log.Printf("AddProductPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
		return
	}

	form.Name = r.PostFormValue("name")
	form.Description = r.PostFormValue("description")
	form.SKU = r.PostFormValue("sku")
	form.Price = r.PostFormValue("price")
	form.Stock = r.PostFormValue("stock")
	form.Weight = r.PostFormValue("weight")
	form.ImagePath = r.PostFormValue("image_path")
	form.CategoryID = r.PostFormValue("category_id")

	log.Printf("AddProductPost: Form diterima - Nama: %s, SKU: %s, Harga: %s, Stok: %s, Weight: %s, ImagePath: %s, CategoryID: %s",
		form.Name, form.SKU, form.Price, form.Stock, form.Weight, form.ImagePath, form.CategoryID)

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		log.Printf("AddProductPost: Validasi form gagal: %v, Errors: %+v", err, formattedErrors)

		data := &AdminProductPageData{
			FormAction:  "/admin/products/add",
			IsEdit:      false,
			ProductData: &form,
			Errors:      formattedErrors,
		}
		h.populateBaseDataForAdmin(r, data)

		categories, catErr := h.categoryRepo.GetAll(r.Context())
		if catErr != nil {
			log.Printf("AddProductPost: Gagal mengambil kategori saat validasi gagal: %v", catErr)
		}
		data.Categories = categories

		data.Title = "Tambah Produk Baru"
		data.IsAuthPage = true
		data.IsAdminPage = true
		data.HideAdminWelcomeMessage = true
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"},
			{Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Produk", URL: "/admin/products"},
			{Name: "Tambah Baru", URL: "/admin/products/add"},
		}
		h.render.HTML(w, http.StatusOK, "admin/products/form", data)
		return
	}
	log.Printf("AddProductPost: Validasi form berhasil.")

	priceFloat, err := strconv.ParseFloat(form.Price, 64)
	if err != nil {
		log.Printf("AddProductPost: Format harga tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Format harga tidak valid.")), http.StatusSeeOther)
		return
	}
	price := decimal.NewFromFloat(priceFloat)
	log.Printf("AddProductPost: Harga dikonversi: %s", price.String())

	stock, err := strconv.Atoi(form.Stock)
	if err != nil {
		log.Printf("AddProductPost: Format stok tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Format stok tidak valid.")), http.StatusSeeOther)
		return
	}
	log.Printf("AddProductPost: Stok dikonversi: %d", stock)

	weightFloat, err := strconv.ParseFloat(form.Weight, 64)
	if err != nil {
		log.Printf("AddProductPost: Format berat tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Format berat tidak valid.")), http.StatusSeeOther)
		return
	}
	weight := decimal.NewFromFloat(weightFloat)

	category, err := h.categoryRepo.GetByID(r.Context(), form.CategoryID)
	if err != nil || category == nil {
		log.Printf("AddProductPost: Kategori tidak ditemukan atau error: %v, CategoryID: %s", err, form.CategoryID)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Kategori tidak valid.")), http.StatusSeeOther)
		return
	}
	log.Printf("AddProductPost: Kategori ditemukan: %s", category.Name)

	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Printf("AddProductPost: UserID tidak ditemukan di konteks. UserID: '%s', OK: %t", userID, ok)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("User admin tidak terautentikasi.")), http.StatusSeeOther)
		return
	}
	log.Printf("AddProductPost: UserID dari konteks: %s", userID)

	newProductID := uuid.New().String()
	productSlug := helpers.GenerateSlug(form.Name) + "-" + newProductID[:8]

	IsSkuExist, err := h.productRepo.IsSKUExists(r.Context(), form.SKU)
	if err != nil {
		log.Printf("AddProductPost: Gagal mengecek SKU unik: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Gagal mengecek SKU.")), http.StatusSeeOther)
		return
	}
	if IsSkuExist {
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("SKU sudah digunakan, gunakan yang lain.")), http.StatusSeeOther)
		return
	}

	product := &models.Product{
		ID:              newProductID,
		UserID:          userID,
		Name:            form.Name,
		Description:     form.Description,
		Sku:             form.SKU,
		Price:           price,
		Stock:           stock,
		Weight:          weight,
		Slug:            productSlug,
		DiscountPercent: decimal.Zero,
		DiscountAmount:  decimal.Zero,
	}
	product.Categories = []models.Category{*category}

	if form.ImagePath != "" {
		product.ProductImages = []models.ProductImage{
			{
				ID:         uuid.New().String(),
				Path:       form.ImagePath,
				ExtraLarge: form.ImagePath,
				Large:      form.ImagePath,
				Medium:     form.ImagePath,
				Small:      form.ImagePath,
			},
		}
	}
	log.Printf("AddProductPost: Objek produk siap disimpan: %+v", product)

	err = h.productRepo.CreateProduct(r.Context(), product)
	if err != nil {
		log.Printf("AddProductPost: Gagal membuat produk di repository: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Gagal menambahkan produk: "+err.Error())), http.StatusSeeOther)
		return
	}
	log.Printf("AddProductPost: Produk berhasil dibuat di repository.")

	http.Redirect(w, r, fmt.Sprintf("/admin/products?status=success&message=%s", url.QueryEscape("Produk berhasil ditambahkan!")), http.StatusSeeOther)
}

func (h *AdminHandler) EditProductPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		log.Printf("EditProductPage: Error mencari produk %s: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan.")), http.StatusSeeOther)
		return
	}
	if product == nil {
		log.Printf("EditProductPage: Produk %s tidak ditemukan", productID)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan.")), http.StatusSeeOther)
		return
	}

	formData := ProductForm{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.Sku,
		Price:       product.Price.String(),
		Stock:       fmt.Sprintf("%d", product.Stock),
		Weight:      product.Weight.String(),
	}

	if len(product.ProductImages) > 0 {
		formData.ImagePath = product.ProductImages[0].Path
	}
	if len(product.Categories) > 0 {
		formData.CategoryID = product.Categories[0].ID
	}

	data := &AdminProductPageData{
		FormAction:  fmt.Sprintf("/admin/products/edit/%s", productID),
		IsEdit:      true,
		ProductData: &formData,
		Errors:      make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	categories, catErr := h.categoryRepo.GetAll(r.Context())
	if catErr != nil {
		log.Printf("EditProductPage: Gagal mengambil kategori: %v", catErr)
		data.Message = "Gagal memuat kategori."
		data.MessageStatus = "error"
	}
	data.Categories = categories

	data.Title = "Edit Produk"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Produk", URL: "/admin/products"},
		{Name: "Edit", URL: fmt.Sprintf("/admin/products/edit/%s", productID)},
	}

	h.render.HTML(w, http.StatusOK, "admin/products/form", data)
}

func (h *AdminHandler) EditProductPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		log.Printf("EditProductPost: Error mencari produk %s untuk pembaruan: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan.")), http.StatusSeeOther)
		return
	}
	if product == nil {
		log.Printf("EditProductPost: Produk %s tidak ditemukan untuk pembaruan", productID)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan.")), http.StatusSeeOther)
		return
	}

	var form ProductForm
	if err := r.ParseForm(); err != nil {
		log.Printf("EditProductPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
		return
	}

	form.ID = productID
	form.Name = r.PostFormValue("name")
	form.Description = r.PostFormValue("description")
	form.SKU = r.PostFormValue("sku")
	form.Price = r.PostFormValue("price")
	form.Stock = r.PostFormValue("stock")
	form.Weight = r.PostFormValue("weight")
	form.ImagePath = r.PostFormValue("image_path")
	form.CategoryID = r.PostFormValue("category_id")

	log.Printf("EditProductPost: Form diterima untuk produk %s - Nama: %s, SKU: %s, Harga: %s, Stok: %s, Weight: %s, ImagePath: %s, CategoryID: %s",
		productID, form.Name, form.SKU, form.Price, form.Stock, form.Weight, form.ImagePath, form.CategoryID)

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		data := &AdminProductPageData{
			FormAction:  fmt.Sprintf("/admin/products/edit/%s", productID),
			IsEdit:      true,
			ProductData: &form,
			Errors:      formattedErrors,
		}
		h.populateBaseDataForAdmin(r, data)

		categories, catErr := h.categoryRepo.GetAll(r.Context())
		if catErr != nil {
			log.Printf("EditProductPost: Gagal mengambil kategori saat validasi gagal: %v", catErr)
		}
		data.Categories = categories

		data.Title = "Edit Produk"
		data.IsAuthPage = true
		data.IsAdminPage = true
		data.HideAdminWelcomeMessage = true
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"},
			{Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Produk", URL: "/admin/products"},
			{Name: "Edit", URL: fmt.Sprintf("/admin/products/edit/%s", productID)},
		}
		h.render.HTML(w, http.StatusOK, "admin/products/form", data)
		return
	}

	priceFloat, err := strconv.ParseFloat(form.Price, 64)
	if err != nil {
		log.Printf("EditProductPost: Format harga tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Format harga tidak valid.")), http.StatusSeeOther)
		return
	}
	price := decimal.NewFromFloat(priceFloat)

	stock, err := strconv.Atoi(form.Stock)
	if err != nil {
		log.Printf("EditProductPost: Format stok tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Format stok tidak valid.")), http.StatusSeeOther)
		return
	}
	weightFloat, err := strconv.ParseFloat(form.Weight, 64)
	if err != nil {
		log.Printf("EditProductPost: Format berat tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Format berat tidak valid.")), http.StatusSeeOther)
		return
	}
	weight := decimal.NewFromFloat(weightFloat)

	category, err := h.categoryRepo.GetByID(r.Context(), form.CategoryID)
	if err != nil || category == nil {
		log.Printf("EditProductPost: Kategori tidak ditemukan: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Kategori tidak valid.")), http.StatusSeeOther)
		return
	}

	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Printf("EditProductPost: UserID tidak ditemukan di konteks")
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("User admin tidak terautentikasi.")), http.StatusSeeOther)
		return
	}

	if product.Name != form.Name {
		product.Slug = helpers.GenerateSlug(form.Name) + "-" + product.ID[:8]
	}

	product.Name = form.Name
	product.Description = form.Description
	product.Sku = form.SKU
	product.Price = price
	product.Stock = stock
	product.Weight = weight
	product.Categories = []models.Category{*category}
	product.UpdatedAt = time.Now()
	product.UserID = userID

	if form.ImagePath != "" {
		if len(product.ProductImages) > 0 {
			product.ProductImages[0].Path = form.ImagePath
			product.ProductImages[0].ExtraLarge = form.ImagePath
			product.ProductImages[0].Large = form.ImagePath
			product.ProductImages[0].Medium = form.ImagePath
			product.ProductImages[0].Small = form.ImagePath
		} else {
			product.ProductImages = []models.ProductImage{
				{
					ID:         uuid.New().String(),
					Path:       form.ImagePath,
					ExtraLarge: form.ImagePath,
					Large:      form.ImagePath,
					Medium:     form.ImagePath,
					Small:      form.ImagePath,
				},
			}
		}
	} else {
		product.ProductImages = []models.ProductImage{}
	}

	err = h.productRepo.UpdateProduct(r.Context(), product)
	if err != nil {
		log.Printf("EditProductPost: Gagal memperbarui produk %s: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=error&message=%s", productID, url.QueryEscape("Gagal memperbarui produk: "+err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/products?status=success&message=%s", url.QueryEscape("Produk berhasil diperbarui!")), http.StatusSeeOther)
}

func (h *AdminHandler) DeleteProductPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil || product == nil {
		log.Printf("DeleteProductPost: Produk %s tidak ditemukan untuk penghapusan: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan atau sudah dihapus.")), http.StatusSeeOther)
		return
	}

	err = h.productRepo.DeleteProduct(r.Context(), productID)
	if err != nil {
		log.Printf("DeleteProductPost: Gagal menghapus produk %s: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Gagal menghapus produk.")), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/products?status=success&message=%s", url.QueryEscape("Produk berhasil dihapus!")), http.StatusSeeOther)
}

func (h *AdminHandler) GetUsersPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminUserPageData{}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Manajemen Pengguna"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Pengguna", URL: "/admin/users"},
	}

	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("GetUsersPage: Gagal mengambil daftar pengguna: %v", err)
		data.Message = "Gagal mengambil daftar pengguna."
		data.MessageStatus = "error"
	} else {
		data.Users = users
	}

	h.render.HTML(w, http.StatusOK, "admin/users/index", data)
}

func (h *AdminHandler) AddUserPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminUserPageData{
		FormAction: "/admin/users/add",
		IsEdit:     false,
		UserData:   &UserForm{},
		Errors:     make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Tambah Pengguna Baru"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Pengguna", URL: "/admin/users"}, {Name: "Tambah Baru", URL: "/admin/users/add"},
	}

	h.render.HTML(w, http.StatusOK, "admin/users/form", data)
}

func (h *AdminHandler) AddUserPost(w http.ResponseWriter, r *http.Request) {
	var form UserForm
	if err := r.ParseForm(); err != nil {
		log.Printf("AddUserPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users/add?status=error&message=%s", url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
		return
	}

	form.FirstName = r.PostFormValue("first_name")
	form.LastName = r.PostFormValue("last_name")
	form.Email = r.PostFormValue("email")
	form.Password = r.PostFormValue("password")
	form.Role = r.PostFormValue("role")

	if form.Password == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/add?status=error&message=%s", url.QueryEscape("Password harus diisi.")), http.StatusSeeOther)
		return
	}

	if len(form.Password) < 6 {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/add?status=error&message=%s", url.QueryEscape("Password minimal 6 karakter.")), http.StatusSeeOther)
		return
	}

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		data := &AdminUserPageData{
			FormAction: "/admin/users/add",
			IsEdit:     false,
			UserData:   &form,
			Errors:     formattedErrors,
		}
		h.populateBaseDataForAdmin(r, data)
		data.Title = "Tambah Pengguna Baru"
		data.IsAuthPage = true
		data.IsAdminPage = true
		data.HideAdminWelcomeMessage = true
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Pengguna", URL: "/admin/users"}, {Name: "Tambah Baru", URL: "/admin/users/add"},
		}
		h.render.HTML(w, http.StatusOK, "admin/users/form", data)
		return
	}

	existingUser, _ := h.userRepo.FindByEmail(r.Context(), form.Email)
	if existingUser != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/add?status=error&message=%s", url.QueryEscape("Email sudah terdaftar.")), http.StatusSeeOther)
		return
	}

	newUser := &models.User{
		ID:        uuid.New().String(),
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     form.Email,
		Role:      form.Role,
	}

	hashedPassword := helpers.HashPassword(form.Password)
	newUser.Password = hashedPassword

	err := h.userRepo.Create(r.Context(), newUser)
	if err != nil {
		log.Printf("AddUserPost: Gagal membuat pengguna: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users/add?status=error&message=%s", url.QueryEscape("Gagal menambahkan pengguna: "+err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/users?status=success&message=%s", url.QueryEscape("Pengguna berhasil ditambahkan!")), http.StatusSeeOther)
}

func (h *AdminHandler) EditUserPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		log.Printf("EditUserPage: Error mencari pengguna %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Pengguna tidak ditemukan.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("EditUserPage: Pengguna %s tidak ditemukan", userID)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Pengguna tidak ditemukan.")), http.StatusSeeOther)
		return
	}

	formData := UserForm{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	}

	data := &AdminUserPageData{
		FormAction: fmt.Sprintf("/admin/users/edit/%s", userID),
		IsEdit:     true,
		UserData:   &formData,
		Errors:     make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Edit Pengguna"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Pengguna", URL: "/admin/users"}, {Name: "Edit", URL: fmt.Sprintf("/admin/users/edit/%s", userID)},
	}

	h.render.HTML(w, http.StatusOK, "admin/users/form", data)
}

func (h *AdminHandler) EditUserPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		log.Printf("EditUserPost: Error mencari pengguna %s untuk pembaruan: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Pengguna tidak ditemukan.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("EditUserPost: Pengguna %s tidak ditemukan untuk pembaruan", userID)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Pengguna tidak ditemukan.")), http.StatusSeeOther)
		return
	}

	var form UserForm
	if err := r.ParseForm(); err != nil {
		log.Printf("EditUserPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users/edit/%s?status=error&message=%s", userID, url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
		return
	}

	form.ID = userID
	form.FirstName = r.PostFormValue("first_name")
	form.LastName = r.PostFormValue("last_name")
	form.Email = r.PostFormValue("email")
	form.Password = r.PostFormValue("password")
	form.Role = r.PostFormValue("role")

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		data := &AdminUserPageData{
			FormAction: fmt.Sprintf("/admin/users/edit/%s", userID),
			IsEdit:     true,
			UserData:   &form,
			Errors:     formattedErrors,
		}
		h.populateBaseDataForAdmin(r, data)
		data.Title = "Edit Pengguna"
		data.IsAuthPage = true
		data.IsAdminPage = true
		data.HideAdminWelcomeMessage = true
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Pengguna", URL: "/admin/users"}, {Name: "Edit", URL: fmt.Sprintf("/admin/users/edit/%s", userID)},
		}
		h.render.HTML(w, http.StatusOK, "admin/users/form", data)
		return
	}

	if user.Email != form.Email {
		existingUser, _ := h.userRepo.FindByEmail(r.Context(), form.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			http.Redirect(w, r, fmt.Sprintf("/admin/users/edit/%s?status=error&message=%s", userID, url.QueryEscape("Email sudah terdaftar oleh pengguna lain.")), http.StatusSeeOther)
			return
		}
	}

	user.FirstName = form.FirstName
	user.LastName = form.LastName
	user.Email = form.Email
	user.Role = form.Role
	user.UpdatedAt = time.Now()

	if form.Password != "" {
		if len(form.Password) < 6 {
			http.Redirect(w, r, fmt.Sprintf("/admin/users/edit/%s?status=error&message=%s", userID, url.QueryEscape("Password minimal 6 karakter.")), http.StatusSeeOther)
			return
		}
		hashedPassword := helpers.HashPassword(form.Password)
		user.Password = hashedPassword
	}

	err = h.userRepo.UpdateUser(r.Context(), user)
	if err != nil {
		log.Printf("EditUserPost: Gagal memperbarui pengguna %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users/edit/%s?status=error&message=%s", userID, url.QueryEscape("Gagal memperbarui pengguna: "+err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/users?status=success&message=%s", url.QueryEscape("Pengguna berhasil diperbarui!")), http.StatusSeeOther)
}

func (h *AdminHandler) DeleteUserPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	currentUserID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if ok && currentUserID == userID {
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Anda tidak dapat menghapus akun Anda sendiri.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil || user == nil {
		log.Printf("DeleteUserPost: Pengguna %s tidak ditemukan untuk penghapusan: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Pengguna tidak ditemukan atau sudah dihapus.")), http.StatusSeeOther)
		return
	}

	err = h.userRepo.DeleteUser(r.Context(), userID)
	if err != nil {
		log.Printf("DeleteUserPost: Gagal menghapus pengguna %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/users?status=error&message=%s", url.QueryEscape("Gagal menghapus pengguna.")), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/users?status=success&message=%s", url.QueryEscape("Pengguna berhasil dihapus!")), http.StatusSeeOther)
}

func (h *AdminHandler) GetCategoriesPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminCategoryPageData{}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Manajemen Kategori"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Kategori", URL: "/admin/categories"},
	}

	categories, err := h.categoryRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("GetCategoriesPage: Gagal mengambil daftar kategori: %v", err)
		data.Message = "Gagal mengambil daftar kategori."
		data.MessageStatus = "error"
	} else {
		data.Categories = categories
	}

	h.render.HTML(w, http.StatusOK, "admin/categories/index", data)
}

func (h *AdminHandler) AddCategoryPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminCategoryPageData{
		FormAction:   "/admin/categories/add",
		IsEdit:       false,
		CategoryData: &CategoryForm{},
		Errors:       make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	sections, err := h.sectionRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("AddCategoryPage: Gagal mengambil daftar section: %v", err)
		data.Message = "Gagal memuat daftar section."
		data.MessageStatus = "error"
	}
	data.Sections = sections

	data.Title = "Tambah Kategori Baru"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Kategori", URL: "/admin/categories"}, {Name: "Tambah Baru", URL: "/admin/categories/add"},
	}

	h.render.HTML(w, http.StatusOK, "admin/categories/form", data)
}

func (h *AdminHandler) AddCategoryPost(w http.ResponseWriter, r *http.Request) {

	section, secErr := h.sectionRepo.GetOrCreateDefaultSection(r.Context())
	if secErr != nil {
		log.Printf("Gagal mengambil/membuat default section: %v", secErr)

	}
	log.Printf("Section terpilih: ID=%s, Slug=%s", section.ID, section.Slug)

	var data AdminCategoryPageData
	data.IsAdminPage = true
	data.IsAuthPage = true
	data.HideAdminWelcomeMessage = true
	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Kategori", URL: "/admin/categories"},
		{Name: "Tambah Baru", URL: "/admin/categories/add"},
	}

	var form CategoryForm
	if err := r.ParseForm(); err != nil {
		log.Printf("AddCategoryPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, "/admin/categories/add?status=error&message=Kesalahan%20parsing%20form", http.StatusSeeOther)
		return
	}
	form.Name = r.PostFormValue("name")
	form.ParentID = r.PostFormValue("parent_id")
	form.SectionID = section.ID

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		data.FormAction = "/admin/categories/add"
		data.IsEdit = false
		data.CategoryData = &form
		data.Errors = helpers.FormatValidationErrors(validationErrors)
		data.Title = "Tambah Kategori Baru"
		h.populateBaseDataForAdmin(r, &data)
		h.render.HTML(w, http.StatusOK, "admin/categories/form", &data)
		return
	}

	categorySlug := helpers.GenerateSlug(form.Name)

	newCategory := &models.Category{
		ID:        uuid.New().String(),
		Name:      form.Name,
		Slug:      categorySlug,
		SectionID: form.SectionID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if form.ParentID != "" {
		newCategory.ParentID = &form.ParentID
	}

	if err := h.categoryRepo.Create(r.Context(), newCategory); err != nil {
		log.Printf("AddCategoryPost: Gagal membuat kategori: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/categories/add?status=error&message=%s", url.QueryEscape("Gagal menambahkan kategori: "+err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *AdminHandler) EditCategoryPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]

	category, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil {
		log.Printf("EditCategoryPage: Error mencari kategori %s: %v", categoryID, err)
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}
	if category == nil {
		log.Printf("EditCategoryPage: Kategori %s tidak ditemukan", categoryID)
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	formData := CategoryForm{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		SectionID: category.SectionID,
	}
	if category.ParentID != nil {
		formData.ParentID = *category.ParentID
	}

	data := &AdminCategoryPageData{
		FormAction:   fmt.Sprintf("/admin/categories/edit/%s", categoryID),
		IsEdit:       true,
		CategoryData: &formData,
		Errors:       make(map[string]string),
	}
	h.populateBaseDataForAdmin(r, data)

	sections, secErr := h.sectionRepo.GetAll(r.Context())
	if secErr != nil {
		log.Printf("EditCategoryPage: Gagal mengambil daftar section: %v", secErr)
		data.Message = "Gagal memuat daftar section."
		data.MessageStatus = "error"
	}
	data.Sections = sections

	data.Title = "Edit Kategori"
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Kategori", URL: "/admin/categories"}, {Name: "Edit", URL: fmt.Sprintf("/admin/categories/edit/%s", categoryID)},
	}

	h.render.HTML(w, http.StatusOK, "admin/categories/form", data)
}

func (h *AdminHandler) EditCategoryPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]

	category, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil || category == nil {
		log.Printf("EditCategoryPost: Kategori %s tidak ditemukan untuk pembaruan: %v", categoryID, err)
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	var form CategoryForm
	if err := r.ParseForm(); err != nil {
		log.Printf("EditCategoryPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/categories/edit/%s?status=error&message=%s", categoryID, url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
		return
	}

	form.ID = categoryID
	form.Name = r.PostFormValue("name")
	form.ParentID = r.PostFormValue("parent_id")
	form.SectionID = r.PostFormValue("section_id")

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		data := &AdminCategoryPageData{
			FormAction:   fmt.Sprintf("/admin/categories/edit/%s", categoryID),
			IsEdit:       true,
			CategoryData: &form,
			Errors:       formattedErrors,
		}
		h.populateBaseDataForAdmin(r, data)

		sections, secErr := h.sectionRepo.GetAll(r.Context())
		if secErr != nil {
			log.Printf("EditCategoryPost: Gagal mengambil section saat validasi gagal: %v", secErr)
		}
		data.Sections = sections

		data.Title = "Edit Kategori"
		data.IsAuthPage = true
		data.IsAdminPage = true
		data.HideAdminWelcomeMessage = true
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Kategori", URL: "/admin/categories"}, {Name: "Edit", URL: fmt.Sprintf("/admin/categories/edit/%s", categoryID)},
		}
		h.render.HTML(w, http.StatusOK, "admin/categories/form", data)
		return
	}

	if category.Name != form.Name {
		category.Slug = helpers.GenerateSlug(form.Name)
	}

	category.Name = form.Name
	category.SectionID = form.SectionID
	if form.ParentID != "" {
		category.ParentID = &form.ParentID
	} else {
		category.ParentID = nil
	}
	category.UpdatedAt = time.Now()

	err = h.categoryRepo.Update(r.Context(), category)
	if err != nil {
		log.Printf("EditCategoryPost: Gagal memperbarui kategori %s: %v", categoryID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/categories/edit/%s?status=error&message=%s", categoryID, url.QueryEscape("Gagal memperbarui kategori: "+err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteCategoryPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]

	category, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil || category == nil {
		log.Printf("DeleteCategoryPost: Kategori %s tidak ditemukan untuk penghapusan: %v", categoryID, err)
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	err = h.categoryRepo.Delete(r.Context(), categoryID)
	if err != nil {
		log.Printf("DeleteCategoryPost: Gagal menghapus kategori %s: %v", categoryID, err)
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}
