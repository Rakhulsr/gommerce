package admin

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

const (
	UploadDir = "./static/uploads/products/"
	MaxImages = 3
)

func (h *AdminHandler) GetProductsPage(w http.ResponseWriter, r *http.Request) {

	pageData := &AdminProductPageData{}

	h.populateBaseDataForAdmin(r, pageData)

	pageData.Title = "Manajemen Produk"
	pageData.IsAuthPage = true
	pageData.IsAdminPage = true
	pageData.HideAdminWelcomeMessage = true

	pageData.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Produk", URL: "/admin/products"},
	}

	products, err := h.productRepo.GetProducts(r.Context())
	if err != nil {
		log.Printf("GetProductsPage: Gagal mengambil daftar produk: %v", err)
		pageData.Message = "Gagal mengambil daftar produk."
		pageData.MessageStatus = "error"
	} else {
		pageData.Products = products
	}

	dataMap := make(map[string]interface{})

	dataMap["Title"] = pageData.Title
	dataMap["IsLoggedIn"] = pageData.IsLoggedIn
	dataMap["User"] = pageData.User
	dataMap["UserID"] = pageData.UserID
	dataMap["CartCount"] = pageData.CartCount
	dataMap["CSRFToken"] = pageData.CSRFToken
	dataMap["Message"] = pageData.Message
	dataMap["MessageStatus"] = pageData.MessageStatus
	dataMap["Query"] = pageData.Query
	dataMap["Breadcrumbs"] = pageData.Breadcrumbs
	dataMap["IsAuthPage"] = pageData.IsAuthPage
	dataMap["IsAdminPage"] = pageData.IsAdminPage
	dataMap["HideAdminWelcomeMessage"] = pageData.HideAdminWelcomeMessage
	dataMap["CurrentPath"] = pageData.CurrentPath
	dataMap["IsAdminRoute"] = pageData.IsAdminRoute

	dataMap["Products"] = pageData.Products
	dataMap["ProductData"] = pageData.ProductData
	dataMap["IsEdit"] = pageData.IsEdit
	dataMap["FormAction"] = pageData.FormAction
	dataMap["Errors"] = pageData.Errors
	dataMap["Categories"] = pageData.Categories

	h.render.HTML(w, http.StatusOK, "admin/products/index", dataMap)
}

func (h *AdminHandler) AddProductPage(w http.ResponseWriter, r *http.Request) {
	data := &AdminProductPageData{
		FormAction: "/admin/products/add",
		IsEdit:     false,
		ProductData: &ProductForm{
			DiscountPercent: "0",
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

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("AddProductPost: Kesalahan parsing multipart form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products/add?status=error&message=%s", url.QueryEscape("Kesalahan parsing form (ukuran file terlalu besar?).")), http.StatusSeeOther)
		return
	}

	var form ProductForm
	form.Name = r.PostFormValue("name")
	form.Description = r.PostFormValue("description")
	form.SKU = r.PostFormValue("sku")
	form.Price = r.PostFormValue("price")
	form.Stock = r.PostFormValue("stock")
	form.Weight = r.PostFormValue("weight")
	form.CategoryID = r.PostFormValue("category_id")
	form.DiscountPercent = r.PostFormValue("discount_percent")

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)
		log.Printf("AddProductPost: Validasi form GAGAL: %v, Errors: %+v", err, formattedErrors)

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
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Produk", URL: "/admin/products"}, {Name: "Tambah Baru", URL: "/admin/products/add"},
		}
		h.render.HTML(w, http.StatusOK, "admin/products/form", data)
		return
	}

	priceFloat, err := strconv.ParseFloat(form.Price, 64)
	if err != nil {
		h.handleFormError(w, r, "/admin/products/add", "Format harga tidak valid.", &form, nil)
		return
	}
	price := decimal.NewFromFloat(priceFloat)

	stock, err := strconv.Atoi(form.Stock)
	if err != nil {
		h.handleFormError(w, r, "/admin/products/add", "Format stok tidak valid.", &form, nil)
		return
	}

	weightFloat, err := strconv.ParseFloat(form.Weight, 64)
	if err != nil {
		h.handleFormError(w, r, "/admin/products/add", "Format berat tidak valid.", &form, nil)
		return
	}
	weight := decimal.NewFromFloat(weightFloat)

	discountPercentFloat, err := strconv.ParseFloat(form.DiscountPercent, 64)
	if err != nil {
		log.Printf("AddProductPost: Format diskon tidak valid atau kosong, setting ke 0: %v", err)
		discountPercentFloat = 0
	}
	discountPercent := decimal.NewFromFloat(discountPercentFloat)

	discountAmount := calc.CalculateDiscount(price, discountPercent)

	category, err := h.categoryRepo.GetByID(r.Context(), form.CategoryID)
	if err != nil || category == nil {
		h.handleFormError(w, r, "/admin/products/add", "Kategori tidak valid.", &form, nil)
		return
	}

	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		h.handleFormError(w, r, "/admin/products/add", "User admin tidak terautentikasi.", &form, nil)
		return
	}

	newProductID := uuid.New().String()
	productSlug := helpers.GenerateSlug(form.Name) + "-" + newProductID[:8]

	IsSkuExist, err := h.productRepo.IsSKUExists(r.Context(), form.SKU)
	if err != nil {
		log.Printf("AddProductPost: Gagal mengecek SKU unik: %v", err)
		h.handleFormError(w, r, "/admin/products/add", "Gagal mengecek SKU.", &form, nil)
		return
	}
	if IsSkuExist {
		h.handleFormError(w, r, "/admin/products/add", "SKU sudah digunakan, gunakan yang lain.", &form, map[string]string{"sku": "SKU ini sudah ada."})
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
		DiscountPercent: discountPercent,
		DiscountAmount:  discountAmount,
	}
	product.Categories = []models.Category{*category}

	files := r.MultipartForm.File["product_images"]
	if len(files) > MaxImages {
		h.handleFormError(w, r, "/admin/products/add", fmt.Sprintf("Anda hanya dapat mengunggah maksimal %d gambar.", MaxImages), &form, map[string]string{"product_images": fmt.Sprintf("Maksimal %d gambar.", MaxImages)})
		return
	}

	var productImages []models.ProductImage
	for _, fileHeader := range files {
		log.Printf("AddProductPost: Memproses file: %s", fileHeader.Filename)
		imagePath, err := h.saveProductImage(fileHeader)
		if err != nil {
			log.Printf("AddProductPost: Gagal menyimpan gambar: %v", err)
			h.handleFormError(w, r, "/admin/products/add", "Gagal menyimpan salah satu gambar.", &form, nil)
			return
		}
		productImages = append(productImages, models.ProductImage{
			ID:         uuid.New().String(),
			ProductID:  newProductID,
			Path:       imagePath,
			ExtraLarge: imagePath,
			Large:      imagePath,
			Medium:     imagePath,
			Small:      imagePath,
		})
	}
	product.ProductImages = productImages

	err = h.productRepo.CreateProduct(r.Context(), product)
	if err != nil {
		log.Printf("AddProductPost: Gagal membuat produk di repository: %v", err)
		h.handleFormError(w, r, "/admin/products/add", "Gagal menambahkan produk: "+err.Error(), &form, nil)
		return
	}

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
		ID:              product.ID,
		Name:            product.Name,
		Description:     product.Description,
		SKU:             product.Sku,
		Price:           product.Price.String(),
		Stock:           fmt.Sprintf("%d", product.Stock),
		Weight:          product.Weight.String(),
		DiscountPercent: product.DiscountPercent.String(),
		ExistingImages:  product.ProductImages,
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("EditProductPost: Kesalahan parsing multipart form: %v", err)

		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Kesalahan parsing form (ukuran file terlalu besar?).", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, nil)
		return
	}

	var form ProductForm
	form.ID = productID
	form.Name = r.PostFormValue("name")
	form.Description = r.PostFormValue("description")
	form.SKU = r.PostFormValue("sku")
	form.Price = r.PostFormValue("price")
	form.Stock = r.PostFormValue("stock")
	form.Weight = r.PostFormValue("weight")
	form.CategoryID = r.PostFormValue("category_id")
	form.DiscountPercent = r.PostFormValue("discount_percent")

	log.Printf("EditProductPost: Form diterima untuk produk %s - Nama: %s, SKU: %s, Harga: %s, Stok: %s, Weight: %s, CategoryID: %s, DiscountPercent: %s",
		productID, form.Name, form.SKU, form.Price, form.Stock, form.Weight, form.CategoryID, form.DiscountPercent)

	if err := h.validator.Struct(&form); err != nil {
		log.Printf("EditProductPost: Validasi form GAGAL: %v", err)
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Validasi form gagal.", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, formattedErrors)
		return
	}

	priceFloat, _ := strconv.ParseFloat(form.Price, 64)
	price := decimal.NewFromFloat(priceFloat)
	stock, _ := strconv.Atoi(form.Stock)
	weightFloat, _ := strconv.ParseFloat(form.Weight, 64)
	weight := decimal.NewFromFloat(weightFloat)
	discountPercentFloat, _ := strconv.ParseFloat(form.DiscountPercent, 64)
	discountPercent := decimal.NewFromFloat(discountPercentFloat)

	if form.SKU != product.Sku {
		IsSkuExist, err := h.productRepo.IsSKUExists(r.Context(), form.SKU)
		if err != nil {
			log.Printf("EditProductPost: Gagal mengecek SKU unik: %v", err)

			h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Gagal mengecek SKU.", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, nil)
			return
		}
		if IsSkuExist {

			h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "SKU sudah digunakan oleh produk lain.", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, map[string]string{"sku": "SKU ini sudah ada."})
			return
		}
	}

	category, err := h.categoryRepo.GetByID(r.Context(), form.CategoryID)
	if err != nil || category == nil {

		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Kategori tidak valid.", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, nil)
		return
	}

	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {

		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "User admin tidak terautentikasi.", &ProductForm{ID: productID, ExistingImages: product.ProductImages}, nil)
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
	product.DiscountPercent = discountPercent
	product.DiscountAmount = calc.CalculateDiscount(price, discountPercent)
	product.UpdatedAt = time.Now()
	product.UserID = userID

	product.Categories = []models.Category{*category}

	var finalProductImages []models.ProductImage = make([]models.ProductImage, 0)

	retainedImageIDsStr := r.PostFormValue("retained_image_ids")
	retainedImageIDsMap := make(map[string]bool)
	if retainedImageIDsStr != "" {
		for _, id := range strings.Split(retainedImageIDsStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				retainedImageIDsMap[id] = true
			}
		}
	}

	if product.ProductImages == nil {
		product.ProductImages = []models.ProductImage{}

	}

	for _, img := range product.ProductImages {

		if retainedImageIDsMap[img.ID] {
			finalProductImages = append(finalProductImages, img)
		} else {

			if img.Path != "" {
				fullPath := filepath.Join(".", img.Path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					log.Printf("EditProductPost: Gagal menghapus file fisik gambar lama yang tidak dipertahankan %s: remove %s: The system cannot find the file specified. (File already gone)", fullPath, fullPath)
				} else if err := os.Remove(fullPath); err != nil {
					log.Printf("EditProductPost: Gagal menghapus file fisik gambar lama yang tidak dipertahankan %s: %v", fullPath, err)
				} else {
					log.Printf("EditProductPost: Berhasil menghapus file fisik gambar lama yang tidak dipertahankan: %s", fullPath)
				}
			} else {
				log.Printf("EditProductPost: Melewatkan penghapusan gambar lama karena path kosong untuk ID: %s", img.ID)
			}
		}
	}

	files := r.MultipartForm.File["product_images"]

	var uploadedProductImages []models.ProductImage = make([]models.ProductImage, 0)
	for _, fileHeader := range files {
		if fileHeader.Size == 0 {
			continue
		}
		imagePath, err := h.saveProductImage(fileHeader)
		if err != nil {
			log.Printf("EditProductPost: Gagal menyimpan gambar baru: %v", err)

			form.ExistingImages = finalProductImages
			h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Gagal menyimpan salah satu gambar baru.", &form, nil)
			return
		}
		uploadedProductImages = append(uploadedProductImages, models.ProductImage{
			ID:         uuid.New().String(),
			ProductID:  productID,
			Path:       imagePath,
			ExtraLarge: imagePath,
			Large:      imagePath,
			Medium:     imagePath,
			Small:      imagePath,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
	}

	product.ProductImages = append(finalProductImages, uploadedProductImages...)

	if len(product.ProductImages) > MaxImages {

		for _, img := range uploadedProductImages {
			fullPath := filepath.Join(".", img.Path)
			if err := os.Remove(fullPath); err != nil {
				log.Printf("EditProductPost: Gagal menghapus file fisik baru karena melebihi batas %s: %v", fullPath, err)
			} else {
				log.Printf("EditProductPost: Berhasil menghapus file fisik baru karena melebihi batas: %s", fullPath)
			}
		}

		product.ProductImages = finalProductImages
		form.ExistingImages = finalProductImages
		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), fmt.Sprintf("Anda hanya dapat memiliki total %d gambar. Total gambar akan menjadi %d (termasuk %d yang baru diunggah).", MaxImages, len(product.ProductImages)+len(files), len(files)), &form, map[string]string{"product_images": fmt.Sprintf("Maksimal %d gambar.", MaxImages)})
		return
	}

	err = h.productRepo.UpdateProduct(r.Context(), product)
	if err != nil {
		log.Printf("EditProductPost: GAGAL memperbarui produk %s: %v", productID, err)

		form.ExistingImages = product.ProductImages
		h.handleFormError(w, r, fmt.Sprintf("/admin/products/edit/%s", productID), "Gagal memperbarui produk: "+err.Error(), &form, nil)
		return
	}

	http.Redirect(w, r, "/admin/products?status=success&message="+url.QueryEscape("Produk berhasil diperbarui!"), http.StatusSeeOther)
}

func (h *AdminHandler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]
	imageID := vars["image_id"]

	if r.Method != http.MethodDelete {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	err := h.productRepo.DeleteProductImage(r.Context(), imageID)
	if err != nil {
		log.Printf("DeleteProductImage: Gagal menghapus gambar: %v", err)
		http.Error(w, "Gagal menghapus gambar.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/products/edit/%s?status=success&message=%s", productID, url.QueryEscape("Gambar berhasil dihapus.")), http.StatusSeeOther)
}

func (h *AdminHandler) DeleteProductPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		log.Printf("DeleteProductPost: Produk %s tidak ditemukan untuk penghapusan: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan atau sudah dihapus.")), http.StatusSeeOther)
		return
	}
	if product == nil {
		log.Printf("DeleteProductPost: Produk %s tidak ditemukan untuk penghapusan (nil product)", productID)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Produk tidak ditemukan atau sudah dihapus.")), http.StatusSeeOther)
		return
	}

	for _, img := range product.ProductImages {
		fullPath := filepath.Join(".", img.Path)
		if err := os.Remove(fullPath); err != nil {
			log.Printf("DeleteProductPost: Gagal menghapus file gambar fisik %s untuk produk %s: %v", fullPath, productID, err)
		} else {
			log.Printf("DeleteProductPost: Berhasil menghapus file gambar fisik: %s", fullPath)
		}
	}

	err = h.productRepo.DeleteProduct(r.Context(), productID)
	if err != nil {
		log.Printf("DeleteProductPost: Gagal menghapus produk %s dari database: %v", productID, err)
		http.Redirect(w, r, fmt.Sprintf("/admin/products?status=error&message=%s", url.QueryEscape("Gagal menghapus produk dari database.")), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/products?status=success&message=%s", url.QueryEscape("Produk berhasil dihapus!")), http.StatusSeeOther)
}

func (h *AdminHandler) handleFormError(w http.ResponseWriter, r *http.Request, redirectURL string, msg string, formData *ProductForm, validationErrors map[string]string) {
	log.Printf("%s: %s", redirectURL, msg)

	data := &AdminProductPageData{
		FormAction:  redirectURL,
		IsEdit:      (redirectURL != "/admin/products/add"),
		ProductData: formData,
		Errors:      validationErrors,
	}
	h.populateBaseDataForAdmin(r, data)

	categories, catErr := h.categoryRepo.GetAll(r.Context())
	if catErr != nil {
		log.Printf("handleFormError: Gagal mengambil kategori: %v", catErr)
	}
	data.Categories = categories
	data.Message = msg
	data.MessageStatus = "error"
	data.Title = "Tambah Produk Baru"
	if data.IsEdit {
		data.Title = "Edit Produk"
	}
	data.IsAuthPage = true
	data.IsAdminPage = true
	data.HideAdminWelcomeMessage = true

	if data.IsEdit && formData.ID != "" {
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Produk", URL: "/admin/products"}, {Name: "Edit", URL: redirectURL},
		}

		if len(formData.ExistingImages) == 0 {
			product, err := h.productRepo.GetByID(r.Context(), formData.ID)
			if err == nil && product != nil {
				formData.ExistingImages = product.ProductImages
			}
		}
	} else {
		data.Breadcrumbs = []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"}, {Name: "Admin", URL: "/admin/dashboard"},
			{Name: "Produk", URL: "/admin/products"}, {Name: "Tambah Baru", URL: redirectURL},
		}
	}
	h.render.HTML(w, http.StatusOK, "admin/products/form", data)
}

func (h *AdminHandler) saveProductImage(fileHeader *multipart.FileHeader) (string, error) {

	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		log.Printf("saveProductImage: Gagal membuat direktori upload: %v", err)
		return "", fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("saveProductImage: Gagal membuka file yang diunggah: %v", err)
		return "", fmt.Errorf("gagal membuka file yang diunggah: %w", err)
	}
	defer file.Close()

	extension := filepath.Ext(fileHeader.Filename)
	uniqueFileName := uuid.New().String() + extension
	filePath := filepath.Join(UploadDir, uniqueFileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("saveProductImage: Gagal menyalin file yang diunggah: %v", err)
		log.Printf("saveProductImage: Menyimpan file ke: %s", filePath)
		return "", fmt.Errorf("gagal membuat file di server: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", fmt.Errorf("gagal menyalin file yang diunggah: %w", err)
	}

	return "/static/uploads/products/" + uniqueFileName, nil
}
