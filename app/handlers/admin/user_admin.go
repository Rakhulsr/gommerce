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
