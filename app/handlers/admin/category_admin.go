package admin

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

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
