package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type KomerceAddressHandler struct {
	render      *render.Render
	addressRepo repositories.AddressRepository
	userRepo    repositories.UserRepositoryImpl
	validate    *validator.Validate
}

func NewKomerceAddressHandler(
	render *render.Render,
	addressRepo repositories.AddressRepository,
	userRepo repositories.UserRepositoryImpl,
	komerceLocationSvc services.KomerceRajaOngkirClient,
	validate *validator.Validate,
) *KomerceAddressHandler {
	return &KomerceAddressHandler{
		render:      render,
		addressRepo: addressRepo,
		userRepo:    userRepo,
		validate:    validate,
	}
}

func (h *KomerceAddressHandler) GetAddressesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("GetAddressesPage: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	addresses, err := h.addressRepo.FindAddressesByUserID(ctx, userID)
	if err != nil {
		log.Printf("GetAddressesPage: Gagal mengambil alamat: %v", err)
		h.render.HTML(w, http.StatusInternalServerError, "error", nil)
		return
	}

	log.Printf("DEBUG: Type of addresses passed to template: %T, Value: %+v", addresses, addresses)

	pageSpecificData := map[string]interface{}{
		"Title":     "Daftar Alamat",
		"Addresses": addresses,
		"Breadcrumbs": []breadcrumb.Breadcrumb{
			{Name: "Home", URL: "/"},
			{Name: "Profile", URL: "/profile"},
			{Name: "Addressses", URL: "/addresses"},
		},
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_list", datas)
}

func (h *KomerceAddressHandler) AddAddressPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("AddAddressPage: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	pageSpecificData := map[string]interface{}{
		"Title":      "Tambah Alamat Baru",
		"FormAction": "/addresses/add",
		"Address":    models.Address{},
		"Errors":     map[string]string{},
		"Breadcrumbs": []breadcrumb.Breadcrumb{
			{Name: "Home", URL: "/"},
			{Name: "Addresses", URL: "/addresses"},
			{Name: "Add Address", URL: "/addresses/add"},
		},
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
}

func (h *KomerceAddressHandler) AddAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("AddAddressPost: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("AddAddressPost: Gagal parse form: %v", err)
		h.render.HTML(w, http.StatusBadRequest, "error", nil)
		return
	}

	province := r.FormValue("province_input")
	city := r.FormValue("city_input")
	district := r.FormValue("district_input")
	subdistrict := r.FormValue("subdistrict_input")

	newAddress := models.Address{
		UserID:     userID,
		Address1:   r.FormValue("address1"),
		Address2:   r.FormValue("address2"),
		Phone:      r.FormValue("phone"),
		LocationID: "",
		PostCode:   r.FormValue("post_code"),
	}

	isPrimary := r.FormValue("is_primary") == "on"
	newAddress.IsPrimary = isPrimary

	errors := make(map[string]string)
	if newAddress.Address1 == "" {
		errors["Address1"] = "Alamat lengkap tidak boleh kosong."
	}
	if newAddress.Phone == "" {
		errors["Phone"] = "Nomor telepon tidak boleh kosong."
	}
	if province == "" {
		errors["Province"] = "Provinsi tidak boleh kosong."
	}
	if city == "" {
		errors["City"] = "Kota/Kabupaten tidak boleh kosong."
	}
	if district == "" {
		errors["District"] = "Kecamatan tidak boleh kosong."
	}
	if subdistrict == "" {
		errors["Subdistrict"] = "Kelurahan/Desa tidak boleh kosong."
	}
	if newAddress.PostCode == "" {
		errors["PostCode"] = "Kode pos tidak boleh kosong."
	}

	if len(errors) > 0 {
		pageSpecificData := map[string]interface{}{
			"Title":         "Tambah Alamat Baru",
			"FormAction":    "/addresses/add",
			"Address":       newAddress,
			"Errors":        errors,
			"Message":       "Mohon periksa kembali input Anda.",
			"MessageStatus": "error",
			"Breadcrumbs": []breadcrumb.Breadcrumb{
				{Name: "Home", URL: "/"},
				{Name: "Alamat", URL: "/addresses"},
				{Name: "Tambah", URL: "/addresses/add"},
			},
		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusBadRequest, "auth/addresses_form", datas)
		return
	}

	newAddress.LocationName = fmt.Sprintf("Kel. %s, Kec. %s, Kota %s, Prov. %s",
		subdistrict,
		district,
		city,
		province,
	)

	if newAddress.IsPrimary {
		err := h.addressRepo.SetAllAddressesNonPrimary(ctx, userID)
		if err != nil {
			log.Printf("AddAddressPost: Gagal mengatur alamat lain menjadi non-utama: %v", err)
		}
	}

	err := h.addressRepo.CreateAddress(ctx, &newAddress)
	if err != nil {
		log.Printf("AddAddressPost: Gagal menambahkan alamat: %v", err)
		h.render.HTML(w, http.StatusInternalServerError, "error", nil)
		return
	}

	http.Redirect(w, r, "/addresses?message=Alamat berhasil ditambahkan&status=success", http.StatusSeeOther)
}

func (h *KomerceAddressHandler) EditAddressPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("EditAddressPage: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil {
		log.Printf("EditAddressPage: Gagal mengambil alamat dengan ID %s: %v", addressID, err)
		http.Redirect(w, r, "/addresses?message=Alamat tidak ditemukan&status=error", http.StatusSeeOther)
		return
	}

	if address.UserID != userID {
		log.Printf("EditAddressPage: User %s mencoba mengedit alamat milik user lain %s", userID, address.UserID)
		http.Redirect(w, r, "/addresses?message=Anda tidak memiliki izin untuk mengedit alamat ini&status=error", http.StatusForbidden)
		return
	}

	pageSpecificData := map[string]interface{}{
		"Title":      "Edit Alamat",
		"FormAction": fmt.Sprintf("/addresses/edit/%s", addressID),
		"Address":    address,
		"Errors":     map[string]string{},
		"Breadcrumbs": []breadcrumb.Breadcrumb{
			{Name: "Home", URL: "/"},
			{Name: "Profile", URL: "/profile"},
			{Name: "Adresses", URL: "/addresses"},
			{Name: "Edit", URL: fmt.Sprintf("/addresses/edit/%s", addressID)},
		},
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
}

func (h *KomerceAddressHandler) EditAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("EditAddressPost: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	if err := r.ParseForm(); err != nil {
		log.Printf("EditAddressPost: Gagal parse form: %v", err)
		h.render.HTML(w, http.StatusBadRequest, "error", nil)
		return
	}

	existingAddress, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil {
		log.Printf("EditAddressPost: Gagal mengambil alamat dengan ID %s: %v", addressID, err)
		http.Redirect(w, r, "/addresses?message=Alamat tidak ditemukan&status=error", http.StatusSeeOther)
		return
	}

	if existingAddress.UserID != userID {
		log.Printf("EditAddressPost: User %s mencoba mengedit alamat milik user lain %s", userID, existingAddress.UserID)
		http.Redirect(w, r, "/addresses?message=Anda tidak memiliki izin untuk mengedit alamat ini&status=error", http.StatusForbidden)
		return
	}

	province := r.FormValue("province_input")
	city := r.FormValue("city_input")
	district := r.FormValue("district_input")
	subdistrict := r.FormValue("subdistrict_input")

	existingAddress.Address1 = r.FormValue("address1")
	existingAddress.Address2 = r.FormValue("address2")
	existingAddress.Phone = r.FormValue("phone")
	existingAddress.LocationID = ""
	existingAddress.PostCode = r.FormValue("post_code")

	isPrimary := r.FormValue("is_primary") == "on"
	existingAddress.IsPrimary = isPrimary

	errors := make(map[string]string)
	if existingAddress.Address1 == "" {
		errors["Address1"] = "Alamat lengkap tidak boleh kosong."
	}
	if existingAddress.Phone == "" {
		errors["Phone"] = "Nomor telepon tidak boleh kosong."
	}
	if province == "" {
		errors["Province"] = "Provinsi tidak boleh kosong."
	}
	if city == "" {
		errors["City"] = "Kota/Kabupaten tidak boleh kosong."
	}
	if district == "" {
		errors["District"] = "Kecamatan tidak boleh kosong."
	}
	if subdistrict == "" {
		errors["Subdistrict"] = "Kelurahan/Desa tidak boleh kosong."
	}
	if existingAddress.PostCode == "" {
		errors["PostCode"] = "Kode pos tidak boleh kosong."
	}

	if len(errors) > 0 {
		pageSpecificData := map[string]interface{}{
			"Title":         "Edit Alamat",
			"FormAction":    fmt.Sprintf("/addresses/edit/%s", addressID),
			"Address":       existingAddress,
			"Errors":        errors,
			"Message":       "Mohon periksa kembali input Anda.",
			"MessageStatus": "error",
			"Breadcrumbs": []breadcrumb.Breadcrumb{
				{Name: "Home", URL: "/"},
				{Name: "Addresses", URL: "/addresses"},
				{Name: "Edit", URL: fmt.Sprintf("/addresses/edit/%s", addressID)},
			},
		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusBadRequest, "auth/addresses_form", datas)
		return
	}

	existingAddress.LocationName = fmt.Sprintf("Kel. %s, Kec. %s, Kota %s, Prov. %s",
		subdistrict,
		district,
		city,
		province,
	)

	if existingAddress.IsPrimary {
		err := h.addressRepo.SetAllAddressesNonPrimary(ctx, userID)
		if err != nil {
			log.Printf("EditAddressPost: Gagal mengatur alamat lain menjadi non-utama: %v", err)
		}
	}

	err = h.addressRepo.UpdateAddress(ctx, existingAddress)
	if err != nil {
		log.Printf("EditAddressPost: Gagal memperbarui alamat: %v", err)
		h.render.HTML(w, http.StatusInternalServerError, "error", nil)
		return
	}

	http.Redirect(w, r, "/addresses?message=Alamat berhasil diperbarui&status=success", http.StatusSeeOther)
}

func (h *KomerceAddressHandler) DeleteAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("DeleteAddressPost: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil {
		log.Printf("DeleteAddressPost: Gagal mengambil alamat dengan ID %s: %v", addressID, err)
		http.Redirect(w, r, "/addresses?message=Alamat tidak ditemukan&status=error", http.StatusSeeOther)
		return
	}

	if address.UserID != userID {
		log.Printf("DeleteAddressPost: User %s mencoba menghapus alamat milik user lain %s", userID, address.UserID)
		http.Redirect(w, r, "/addresses?message=Anda tidak memiliki izin untuk menghapus alamat ini&status=error", http.StatusForbidden)
		return
	}

	err = h.addressRepo.DeleteAddress(ctx, addressID)
	if err != nil {
		log.Printf("DeleteAddressPost: Gagal menghapus alamat: %v", err)
		h.render.HTML(w, http.StatusInternalServerError, "error", nil)
		return
	}

	http.Redirect(w, r, "/addresses?message=Alamat berhasil dihapus&status=success", http.StatusSeeOther)
}

func (h *KomerceAddressHandler) SetPrimaryAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("SetPrimaryAddressPost: UserID tidak ditemukan di konteks atau kosong. Mengarahkan ke login.")
		http.Redirect(w, r, fmt.Sprintf("/login?message=%s&status=error", url.QueryEscape("Sesi Anda telah berakhir atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil {
		log.Printf("SetPrimaryAddressPost: Gagal mengambil alamat dengan ID %s: %v", addressID, err)
		http.Redirect(w, r, "/addresses?message=Alamat tidak ditemukan&status=error", http.StatusSeeOther)
		return
	}

	if address.UserID != userID {
		log.Printf("SetPrimaryAddressPost: User %s mencoba mengatur alamat milik user lain %s sebagai utama", userID, address.UserID)
		http.Redirect(w, r, "/addresses?message=Anda tidak memiliki izin untuk mengatur alamat ini&status=error", http.StatusForbidden)
		return
	}

	err = h.addressRepo.SetAllAddressesNonPrimary(ctx, userID)
	if err != nil {
		log.Printf("SetPrimaryAddressPost: Gagal mengatur alamat lain menjadi non-utama: %v", err)
	}

	address.IsPrimary = true
	err = h.addressRepo.UpdateAddress(ctx, address)
	if err != nil {
		log.Printf("SetPrimaryAddressPost: Gagal mengatur alamat %s sebagai utama: %v", addressID, err)
		h.render.HTML(w, http.StatusInternalServerError, "error", nil)
		return
	}

	http.Redirect(w, r, "/addresses?message=Alamat utama berhasil diubah&status=success", http.StatusSeeOther)
}
