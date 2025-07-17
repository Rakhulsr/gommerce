package admin

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

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
		FormAction: "/admin/products/add",
		IsEdit:     false,
		ProductData: &ProductForm{
			DiscountPercent: "0", // Default value for new product
		},
		Errors: make(map[string]string),
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
	form.DiscountPercent = r.PostFormValue("discount_percent") // <-- AMBIL DARI FORM

	log.Printf("AddProductPost: Form diterima - Nama: %s, SKU: %s, Harga: %s, Stok: %s, Weight: %s, ImagePath: %s, CategoryID: %s, DiscountPercent: %s",
		form.Name, form.SKU, form.Price, form.Stock, form.Weight, form.ImagePath, form.CategoryID, form.DiscountPercent)

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

	discountPercentFloat, err := strconv.ParseFloat(form.DiscountPercent, 64) // <-- KONVERSI DISKON PERSEN
	if err != nil {
		log.Printf("AddProductPost: Format persentase diskon tidak valid: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Format persentase diskon tidak valid.")), http.StatusSeeOther)
		return
	}
	discountPercent := decimal.NewFromFloat(discountPercentFloat)
	// Hitung DiscountAmount berdasarkan Price dan DiscountPercent
	discountAmount := calc.CalculateDiscount(price, discountPercent) // <-- HITUNG DISKON AMOUNT

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
		DiscountPercent: discountPercent, // <-- SET DISKON PERSEN
		DiscountAmount:  discountAmount,  // <-- SET DISKON AMOUNT
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
