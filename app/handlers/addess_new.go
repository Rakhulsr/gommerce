package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type KomerceAddressHandler struct {
	render             *render.Render
	addressRepo        repositories.AddressRepository
	userRepo           repositories.UserRepositoryImpl
	komerceShippingSvc services.KomerceRajaOngkirClient
	validate           *validator.Validate
}

func NewKomerceAddressHandler(
	render *render.Render,
	addressRepo repositories.AddressRepository,
	userRepo repositories.UserRepositoryImpl,
	komerceShippingSvc services.KomerceRajaOngkirClient,
	validate *validator.Validate,
) *KomerceAddressHandler {
	return &KomerceAddressHandler{
		render:             render,
		addressRepo:        addressRepo,
		userRepo:           userRepo,
		komerceShippingSvc: komerceShippingSvc,
		validate:           validate,
	}
}

func parseLocationNameForDisplay(locationName string) (subdistrict, district, city, province string) {
	if locationName == "" {
		return "", "", "", ""
	}
	parts := strings.Split(locationName, ", ")

	if len(parts) > 0 {
		subdistrict = parts[0]
	}
	if len(parts) > 1 {
		district = parts[1]
	}
	if len(parts) > 2 {
		city = parts[2]
	}
	if len(parts) > 3 {
		province = parts[3]
	}
	return subdistrict, district, city, province
}

func (h *KomerceAddressHandler) GetAddressesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	addresses, err := h.addressRepo.FindAddressesByUserID(ctx, userID)
	if err != nil {
		log.Printf("GetAddressesPage: Gagal mengambil alamat untuk user %s: %v", userID, err)
		http.Error(w, "Gagal memuat alamat", http.StatusInternalServerError)
		return
	}

	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	pageSpecificData := map[string]interface{}{
		"Title":         "Daftar Alamat",
		"Addresses":     addresses,
		"Breadcrumbs":   []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}},
		"MessageStatus": status,
		"Message":       message,
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_list", datas)
}

func (h *KomerceAddressHandler) AddAddressPage(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	pageSpecificData := map[string]interface{}{
		"Title":           "Tambah Alamat Baru",
		"Breadcrumbs":     []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Add", URL: "/addresses/add"}},
		"MessageStatus":   status,
		"Message":         message,
		"Address":         &models.Address{},
		"SubdistrictName": "",
		"DistrictName":    "",
		"CityName":        "",
		"ProvinceName":    "",
		"PostCode":        "",
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
}

func (h *KomerceAddressHandler) AddAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("AddAddressPost: Gagal parse form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses/add?status=error&message=%s", url.QueryEscape("Gagal memproses data alamat.")), http.StatusSeeOther)
		return
	}

	subdistrictID := r.FormValue("subdistrict_id")
	provinceName := r.FormValue("province_name")
	cityName := r.FormValue("city_name")
	districtName := r.FormValue("district_name")
	subdistrictName := r.FormValue("subdistrict_name")
	postCode := r.FormValue("post_code")

	locationParts := []string{}
	if subdistrictName != "" {
		locationParts = append(locationParts, subdistrictName)
	}
	if districtName != "" {
		locationParts = append(locationParts, districtName)
	}
	if cityName != "" {
		locationParts = append(locationParts, cityName)
	}
	if provinceName != "" {
		locationParts = append(locationParts, provinceName)
	}
	locationName := strings.Join(locationParts, ", ")

	if subdistrictID == "" || locationName == "" || postCode == "" {
		log.Printf("AddAddressPost: Data lokasi tidak lengkap. Subdistrict ID: '%s', Location Name: '%s', Post Code: '%s'", subdistrictID, locationName, postCode)
		pageSpecificData := map[string]interface{}{
			"Title":         "Tambah Alamat Baru",
			"Breadcrumbs":   []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Add", URL: "/addresses/add"}},
			"MessageStatus": "error",
			"Message":       url.QueryEscape("Lokasi (Provinsi, Kota, Kecamatan, Kelurahan/Desa) dan Kode Pos harus dipilih."),
			"Address": &models.Address{
				Name:      r.FormValue("name"),
				Address1:  r.FormValue("address1"),
				Address2:  r.FormValue("address2"),
				Phone:     r.FormValue("phone"),
				Email:     r.FormValue("email"),
				IsPrimary: r.FormValue("is_primary") == "on",
			},
			"SubdistrictName": subdistrictName,
			"DistrictName":    districtName,
			"CityName":        cityName,
			"ProvinceName":    provinceName,
			"PostCode":        postCode,
			"Errors":          "error disini add adress post",
			// "Errors":          helpers.ParseValidationErrors(h.validate.Struct(&models.Address{})),

		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
		return
	}

	newAddress := models.Address{
		UserID:       userID,
		Name:         r.FormValue("name"),
		Address1:     r.FormValue("address1"),
		Address2:     r.FormValue("address2"),
		LocationID:   subdistrictID,
		LocationName: locationName,
		PostCode:     postCode,
		Phone:        r.FormValue("phone"),
		Email:        r.FormValue("email"),
		IsPrimary:    r.FormValue("is_primary") == "on",
	}

	if err := h.validate.Struct(newAddress); err != nil {
		log.Printf("AddAddressPost: Validasi alamat gagal: %v", err)

		pageSpecificData := map[string]interface{}{
			"Title":           "Tambah Alamat Baru",
			"Breadcrumbs":     []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Add", URL: "/addresses/add"}},
			"MessageStatus":   "error",
			"Message":         fmt.Sprintf("Data alamat tidak valid: %v", err),
			"Address":         &newAddress,
			"SubdistrictName": subdistrictName,
			"DistrictName":    districtName,
			"CityName":        cityName,
			"ProvinceName":    provinceName,
			"PostCode":        postCode,
			"Errors":          err,
		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
		return
	}

	err := h.addressRepo.CreateAddress(ctx, &newAddress)
	if err != nil {
		log.Printf("AddAddressPost: Gagal menyimpan alamat: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses/add?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal menyimpan alamat: %v", err))), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil ditambahkan!")), http.StatusSeeOther)
}

func (h *KomerceAddressHandler) EditAddressPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil || address.UserID != userID {
		log.Printf("EditAddressPage: Alamat tidak ditemukan atau tidak diizinkan: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	subdistrictName, districtName, cityName, provinceName := parseLocationNameForDisplay(address.LocationName)

	pageSpecificData := map[string]interface{}{
		"Title":           "Edit Alamat",
		"Address":         address,
		"Breadcrumbs":     []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Edit", URL: fmt.Sprintf("/addresses/edit/%s", addressID)}},
		"MessageStatus":   status,
		"Message":         message,
		"SubdistrictName": subdistrictName,
		"DistrictName":    districtName,
		"CityName":        cityName,
		"ProvinceName":    provinceName,
		"PostCode":        address.PostCode,
	}
	datas := helpers.GetBaseData(r, pageSpecificData)
	h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
}

func (h *KomerceAddressHandler) EditAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	if err := r.ParseForm(); err != nil {
		log.Printf("EditAddressPost: Gagal parse form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses/edit/%s?status=error&message=%s", addressID, url.QueryEscape("Gagal memproses data alamat.")), http.StatusSeeOther)
		return
	}

	existingAddress, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil || existingAddress.UserID != userID {
		log.Printf("EditAddressPost: Alamat tidak ditemukan atau tidak diizinkan: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	subdistrictID := r.FormValue("subdistrict_id")
	provinceName := r.FormValue("province_name")
	cityName := r.FormValue("city_name")
	districtName := r.FormValue("district_name")
	subdistrictName := r.FormValue("subdistrict_name")
	postCode := r.FormValue("post_code")

	locationParts := []string{}
	if subdistrictName != "" {
		locationParts = append(locationParts, subdistrictName)
	}
	if districtName != "" {
		locationParts = append(locationParts, districtName)
	}
	if cityName != "" {
		locationParts = append(locationParts, cityName)
	}
	if provinceName != "" {
		locationParts = append(locationParts, provinceName)
	}
	locationName := strings.Join(locationParts, ", ")

	if subdistrictID == "" || locationName == "" || postCode == "" {
		log.Printf("EditAddressPost: Data lokasi tidak lengkap. Subdistrict ID: '%s', Location Name: '%s', Post Code: '%s'", subdistrictID, locationName, postCode)
		pageSpecificData := map[string]interface{}{
			"Title":           "Edit Alamat",
			"Breadcrumbs":     []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Edit", URL: fmt.Sprintf("/addresses/edit/%s", addressID)}},
			"MessageStatus":   "error",
			"Message":         url.QueryEscape("Lokasi (Provinsi, Kota, Kecamatan, Kelurahan/Desa) dan Kode Pos harus dipilih."),
			"Address":         existingAddress,
			"SubdistrictName": subdistrictName,
			"DistrictName":    districtName,
			"CityName":        cityName,
			"ProvinceName":    provinceName,
			"PostCode":        postCode,
			"Errors":          err,
		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
		return
	}

	existingAddress.Name = r.FormValue("name")
	existingAddress.Address1 = r.FormValue("address1")
	existingAddress.Address2 = r.FormValue("address2")
	existingAddress.LocationID = subdistrictID
	existingAddress.LocationName = locationName
	existingAddress.PostCode = postCode
	existingAddress.Phone = r.FormValue("phone")
	existingAddress.Email = r.FormValue("email")
	existingAddress.IsPrimary = r.FormValue("is_primary") == "on"

	if err := h.validate.Struct(existingAddress); err != nil {
		log.Printf("EditAddressPost: Validasi alamat gagal: %v", err)

		pageSpecificData := map[string]interface{}{
			"Title":           "Edit Alamat",
			"Breadcrumbs":     []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Profile", URL: "/profile"}, {Name: "Addresses", URL: "/addresses"}, {Name: "Edit", URL: fmt.Sprintf("/addresses/edit/%s", addressID)}},
			"MessageStatus":   "error",
			"Message":         fmt.Sprintf("Data alamat tidak valid: %v", err),
			"Address":         existingAddress,
			"SubdistrictName": subdistrictName,
			"DistrictName":    districtName,
			"CityName":        cityName,
			"ProvinceName":    provinceName,
			"PostCode":        postCode,
			"Errors":          err,
		}
		datas := helpers.GetBaseData(r, pageSpecificData)
		h.render.HTML(w, http.StatusOK, "auth/addresses_form", datas)
		return
	}

	err = h.addressRepo.UpdateAddress(ctx, existingAddress)
	if err != nil {
		log.Printf("EditAddressPost: Gagal memperbarui alamat: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses/edit/%s?status=error&message=%s", addressID, url.QueryEscape(fmt.Sprintf("Gagal memperbarui alamat: %v", err))), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil diperbarui!")), http.StatusSeeOther)
}

func (h *KomerceAddressHandler) DeleteAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil || address.UserID != userID {
		log.Printf("DeleteAddressPost: Alamat tidak ditemukan atau tidak diizinkan: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	err = h.addressRepo.DeleteAddress(ctx, addressID)
	if err != nil {
		log.Printf("DeleteAddressPost: Gagal menghapus alamat: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal menghapus alamat: %v", err))), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil dihapus!")), http.StatusSeeOther)
}

func (h *KomerceAddressHandler) SetPrimaryAddressPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	addressID := vars["id"]

	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil || address.UserID != userID {
		log.Printf("SetPrimaryAddressPost: Alamat tidak ditemukan atau tidak diizinkan: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau Anda tidak memiliki izin.")), http.StatusSeeOther)
		return
	}

	err = h.addressRepo.SetPrimaryAddress(ctx, userID, addressID)
	if err != nil {
		log.Printf("SetPrimaryAddressPost: Gagal mengatur alamat utama: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal mengatur alamat utama: %v", err))), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat utama berhasil diatur!")), http.StatusSeeOther)
}

func (h *KomerceAddressHandler) SearchDomesticDestinationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	if query == "" {
		h.render.JSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    []other.KomerceDomesticDestination{},
		})
		return
	}

	destinations, err := h.komerceShippingSvc.SearchDomesticDestinations(ctx, query, limit, offset)
	if err != nil {
		log.Printf("SearchDomesticDestinationsHandler: Gagal mencari destinasi: %v", err)
		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Gagal mencari destinasi: %v", err),
		})
		return
	}

	h.render.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    destinations,
	})
}
