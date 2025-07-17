package handlers

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"time"

// 	"github.com/Rakhulsr/go-ecommerce/app/helpers"
// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/Rakhulsr/go-ecommerce/app/models/other"
// 	"github.com/Rakhulsr/go-ecommerce/app/repositories"
// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
// 	"github.com/go-playground/validator/v10"
// 	"github.com/gorilla/mux"
// 	"github.com/unrolled/render"
// )

// type AddressHandler struct {
// 	render        *render.Render
// 	addressRepo   repositories.AddressRepository
// 	userRepo      repositories.UserRepositoryImpl
// 	rajaOngkirSvc services.RajaOngkirClient
// 	validator     *validator.Validate
// }

// func NewAddressHandler(
// 	render *render.Render,
// 	addressRepo repositories.AddressRepository,
// 	userRepo repositories.UserRepositoryImpl,
// 	rajaOngkirSvc services.RajaOngkirClient,
// 	validator *validator.Validate,
// ) *AddressHandler {
// 	return &AddressHandler{
// 		render:        render,
// 		addressRepo:   addressRepo,
// 		userRepo:      userRepo,
// 		rajaOngkirSvc: rajaOngkirSvc,
// 		validator:     validator,
// 	}
// }

// type AddressWithNames struct {
// 	models.Address
// 	ProvinceName string
// 	CityName     string
// }

// type AddressPageData struct {
// 	other.BasePageData
// 	Addresses   []AddressWithNames
// 	AddressData *AddressForm
// 	Provinces   []other.Province
// 	Cities      []other.City
// 	IsEdit      bool
// 	FormAction  string
// 	Errors      map[string]string
// }

// type AddressForm struct {
// 	Name       string `form:"name" validate:"required"`
// 	Phone      string `form:"phone" validate:"required,numeric,min=10,max=15"`
// 	Email      string `form:"email" validate:"required,email"`
// 	Address1   string `form:"address1" validate:"required"`
// 	Address2   string `form:"address2"`
// 	ProvinceID string `form:"province_id" validate:"required"`
// 	CityID     string `form:"city_id" validate:"required"`
// 	CityName   string
// 	PostCode   string `form:"post_code" validate:"required,numeric"`
// }

// func (h *AddressHandler) populateBaseData(r *http.Request, pageData *AddressPageData) {
// 	baseDataMap := helpers.GetBaseData(r, nil)

// 	if title, ok := baseDataMap["Title"].(string); ok {
// 		pageData.Title = title
// 	}
// 	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
// 		pageData.IsLoggedIn = isLoggedIn
// 	}

// 	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok {
// 		pageData.User = user
// 	}

// 	if userID, ok := baseDataMap["UserID"].(string); ok {
// 		pageData.UserID = userID
// 	}
// 	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
// 		pageData.CartCount = cartCount
// 	}
// 	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
// 		pageData.CSRFToken = csrfToken
// 	}
// 	if message, ok := baseDataMap["Message"].(string); ok {
// 		pageData.Message = message
// 	}
// 	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
// 		pageData.MessageStatus = messageStatus
// 	}
// 	if query, ok := baseDataMap["Query"].(url.Values); ok {
// 		pageData.Query = query
// 	}
// 	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
// 		pageData.Breadcrumbs = breadcrumbs
// 	}
// 	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
// 		pageData.IsAuthPage = isAuthPage
// 	}

// 	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
// 		pageData.IsAdminPage = isAdminPage
// 	}
// }

// func (h *AddressHandler) GetAddressesPage(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	addresses, err := h.addressRepo.FindAddressesByUserID(ctx, userID)
// 	if err != nil {
// 		log.Printf("GetAddressesPage: Failed to retrieve addresses for user %s: %v", userID, err)
// 		data := AddressPageData{}
// 		h.populateBaseData(r, &data)
// 		data.Message = "Gagal mengambil daftar alamat."
// 		data.MessageStatus = "error"
// 		data.Title = "Error"
// 		h.render.HTML(w, http.StatusInternalServerError, "error", data)
// 		return
// 	}

// 	addressesWithNames := make([]AddressWithNames, len(addresses))
// 	allProvinces, err := h.rajaOngkirSvc.GetProvincesFromAPI()
// 	if err != nil {
// 		log.Printf("GetAddressesPage: Error fetching all provinces from RajaOngkir: %v", err)

// 	}
// 	provinceMap := make(map[string]string)
// 	for _, p := range allProvinces {
// 		provinceMap[p.ID] = p.Name
// 	}

// 	citiesCache := make(map[string][]other.City)

// 	for i, addr := range addresses {
// 		addressesWithNames[i].Address = addr

// 		if name, ok := provinceMap[addr.ProvinceID]; ok {
// 			addressesWithNames[i].ProvinceName = name
// 		} else {
// 			addressesWithNames[i].ProvinceName = "Provinsi Tidak Dikenal"
// 		}

// 		var citiesInProvince []other.City
// 		var foundCity bool
// 		if cachedCities, ok := citiesCache[addr.ProvinceID]; ok {
// 			citiesInProvince = cachedCities
// 			foundCity = true
// 		} else {
// 			c, cErr := h.rajaOngkirSvc.GetCitiesFromAPI(addr.ProvinceID)
// 			if cErr != nil {
// 				log.Printf("GetAddressesPage: Error fetching cities for province %s: %v", addr.ProvinceID, cErr)
// 				addressesWithNames[i].CityName = "Kota Tidak Dikenal"
// 			} else {
// 				citiesInProvince = c
// 				citiesCache[addr.ProvinceID] = c
// 				foundCity = true
// 			}
// 		}

// 		if foundCity {
// 			for _, city := range citiesInProvince {
// 				if city.ID == addr.CityID {
// 					addressesWithNames[i].CityName = city.Name
// 					break
// 				}
// 			}
// 			if addressesWithNames[i].CityName == "" {
// 				addressesWithNames[i].CityName = "Kota Tidak Dikenal"
// 			}
// 		} else {
// 			addressesWithNames[i].CityName = "Kota Tidak Dikenal"
// 		}
// 	}

// 	data := AddressPageData{
// 		Addresses: addressesWithNames,
// 	}
// 	h.populateBaseData(r, &data)

// 	data.Title = "Daftar Alamat Saya"
// 	data.Breadcrumbs = []breadcrumb.Breadcrumb{
// 		{Name: "Beranda", URL: "/"},
// 		{Name: "Profile", URL: "/profile"},
// 		{Name: "Alamat Saya", URL: "/addresses"},
// 	}
// 	data.IsAuthPage = true

// 	h.render.HTML(w, http.StatusOK, "auth/addresses_index", data)
// }

// func (h *AddressHandler) AddAddressPage(w http.ResponseWriter, r *http.Request) {
// 	provinces, err := h.rajaOngkirSvc.GetProvincesFromAPI()
// 	if err != nil {
// 		log.Printf("AddAddressPage: Failed to get provinces from RajaOngkir: %v", err)
// 		provinces = []other.Province{}
// 	}

// 	data := AddressPageData{
// 		FormAction:  "/addresses/add",
// 		IsEdit:      false,
// 		AddressData: nil,
// 		Provinces:   provinces,
// 		Cities:      []other.City{},
// 		Errors:      make(map[string]string),
// 	}
// 	h.populateBaseData(r, &data)

// 	data.Title = "Tambah Alamat Baru"
// 	data.Breadcrumbs = []breadcrumb.Breadcrumb{
// 		{Name: "Beranda", URL: "/"},
// 		{Name: "Profile ", URL: "/profile"},
// 		{Name: "Alamat Saya", URL: "/addresses"},
// 		{Name: "Tambah Baru", URL: "/addresses/add"},
// 	}
// 	data.IsAuthPage = true

// 	h.render.HTML(w, http.StatusOK, "auth/addresses_form", data)
// }

// func (h *AddressHandler) AddAddressPost(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	var form AddressForm
// 	if err := r.ParseForm(); err != nil {
// 		log.Printf("AddAddressPost: Error parsing form: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses/add?status=error&message=%s", url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
// 		return
// 	}

// 	form.Name = r.PostFormValue("name")
// 	form.Phone = r.PostFormValue("phone")
// 	form.Email = r.PostFormValue("email")
// 	form.Address1 = r.PostFormValue("address1")
// 	form.Address2 = r.PostFormValue("address2")
// 	form.ProvinceID = r.PostFormValue("province_id")
// 	form.CityID = r.PostFormValue("city_id")
// 	form.PostCode = r.PostFormValue("post_code")

// 	log.Printf("AddAddressPost: Form received - Name: %s, Phone: %s, Email: %s, ProvinceID: %s, CityID: %s, PostCode: %s, IsPrimary: %t",
// 		form.Name, form.Phone, form.Email, form.ProvinceID, form.CityID, form.PostCode)

// 	if err := h.validator.Struct(&form); err != nil {
// 		validationErrors := err.(validator.ValidationErrors)
// 		formattedErrors := helpers.FormatValidationErrors(validationErrors)

// 		provinces, pErr := h.rajaOngkirSvc.GetProvincesFromAPI()
// 		if pErr != nil {
// 			log.Printf("AddAddressPost: Failed to get provinces on validation error: %v", pErr)
// 			provinces = []other.Province{}
// 		}

// 		cities := []other.City{}
// 		var currentCityName string
// 		if form.ProvinceID != "" {
// 			c, cErr := h.rajaOngkirSvc.GetCitiesFromAPI(form.ProvinceID)
// 			if cErr != nil {
// 				log.Printf("AddAddressPost: Failed to get cities on validation error for province %s: %v", form.ProvinceID, cErr)
// 			} else {
// 				cities = c
// 				for _, city := range cities {
// 					if city.ID == form.CityID {
// 						currentCityName = city.Name
// 						break
// 					}
// 				}
// 			}
// 		}
// 		form.CityName = currentCityName

// 		data := AddressPageData{
// 			FormAction:  "/addresses/add",
// 			IsEdit:      false,
// 			AddressData: &form,
// 			Provinces:   provinces,
// 			Cities:      cities,
// 			Errors:      formattedErrors,
// 		}
// 		h.populateBaseData(r, &data)

// 		data.Title = "Tambah Alamat Baru"
// 		data.Breadcrumbs = []breadcrumb.Breadcrumb{
// 			{Name: "Beranda", URL: "/"},
// 			{Name: "Profile", URL: "/profile"},
// 			{Name: "Alamat Saya", URL: "/addresses"},
// 			{Name: "Tambah Baru", URL: "/addresses/add"},
// 		}
// 		data.IsAuthPage = true

// 		h.render.HTML(w, http.StatusOK, "auth/addresses_form", data)
// 		return
// 	}

// 	address := &models.Address{
// 		UserID:     userID,
// 		Name:       form.Name,
// 		Phone:      form.Phone,
// 		Email:      form.Email,
// 		Address1:   form.Address1,
// 		Address2:   form.Address2,
// 		ProvinceID: form.ProvinceID,
// 		CityID:     form.CityID,
// 		PostCode:   form.PostCode,

// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	err := h.addressRepo.CreateAddress(ctx, address)
// 	if err != nil {
// 		log.Printf("AddAddressPost: Failed to create address for user %s: %v", userID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses/add?status=error&message=%s", url.QueryEscape("Gagal menambahkan alamat: "+err.Error())), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil ditambahkan!")), http.StatusSeeOther)
// }

// func (h *AddressHandler) EditAddressPage(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	vars := mux.Vars(r)
// 	addressID := vars["id"]

// 	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
// 	if err != nil {
// 		log.Printf("EditAddressPage: Error finding address %s: %v", addressID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan.")), http.StatusSeeOther)
// 		return
// 	}
// 	if address == nil || address.UserID != userID {
// 		log.Printf("EditAddressPage: Address %s not found or unauthorized for user %s", addressID, userID)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau tidak berhak mengakses.")), http.StatusSeeOther)
// 		return
// 	}

// 	provinces, pErr := h.rajaOngkirSvc.GetProvincesFromAPI()
// 	if pErr != nil {
// 		log.Printf("EditAddressPage: Failed to get provinces from RajaOngkir: %v", pErr)
// 		provinces = []other.Province{}
// 	}

// 	cities := []other.City{}
// 	var currentCityName string
// 	if address.ProvinceID != "" {
// 		c, cErr := h.rajaOngkirSvc.GetCitiesFromAPI(address.ProvinceID)
// 		if cErr != nil {
// 			log.Printf("EditAddressPage: Failed to get cities for province %s: %v", address.ProvinceID, cErr)
// 		} else {
// 			cities = c
// 			for _, city := range cities {
// 				if city.ID == address.CityID {
// 					currentCityName = city.Name
// 					break
// 				}
// 			}
// 		}
// 	}

// 	formData := AddressForm{
// 		Name:       address.Name,
// 		Phone:      address.Phone,
// 		Email:      address.Email,
// 		Address1:   address.Address1,
// 		Address2:   address.Address2,
// 		ProvinceID: address.ProvinceID,
// 		CityID:     address.CityID,
// 		CityName:   currentCityName,
// 		PostCode:   address.PostCode,
// 	}

// 	data := AddressPageData{
// 		FormAction:  fmt.Sprintf("/addresses/edit/%s", addressID),
// 		IsEdit:      true,
// 		AddressData: &formData,
// 		Provinces:   provinces,
// 		Cities:      cities,
// 		Errors:      make(map[string]string),
// 	}
// 	h.populateBaseData(r, &data)

// 	data.Title = "Edit Alamat"
// 	data.Breadcrumbs = []breadcrumb.Breadcrumb{
// 		{Name: "Beranda", URL: "/"},
// 		{Name: "Profile", URL: "/profile"},
// 		{Name: "Alamat Saya", URL: "/addresses"},
// 		{Name: "Edit Alamat", URL: fmt.Sprintf("/addresses/edit/%s", addressID)},
// 	}
// 	data.IsAuthPage = true

// 	h.render.HTML(w, http.StatusOK, "auth/addresses_form", data)
// }

// func (h *AddressHandler) EditAddressPost(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	vars := mux.Vars(r)
// 	addressID := vars["id"]

// 	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
// 	if err != nil {
// 		log.Printf("EditAddressPost: Error finding address %s for update: %v", addressID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan.")), http.StatusSeeOther)
// 		return
// 	}
// 	if address == nil || address.UserID != userID {
// 		log.Printf("EditAddressPost: Address %s not found or unauthorized for user %s for update", addressID, userID)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau tidak berhak mengakses.")), http.StatusSeeOther)
// 		return
// 	}

// 	var form AddressForm
// 	if err := r.ParseForm(); err != nil {
// 		log.Printf("EditAddressPost: Error parsing form: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses/edit/%s?status=error&message=%s", addressID, url.QueryEscape("Kesalahan parsing form.")), http.StatusSeeOther)
// 		return
// 	}

// 	form.Name = r.PostFormValue("name")
// 	form.Phone = r.PostFormValue("phone")
// 	form.Email = r.PostFormValue("email")
// 	form.Address1 = r.PostFormValue("address1")
// 	form.Address2 = r.PostFormValue("address2")
// 	form.ProvinceID = r.PostFormValue("province_id")
// 	form.CityID = r.PostFormValue("city_id")
// 	form.PostCode = r.PostFormValue("post_code")

// 	log.Printf("EditAddressPost: Form received - Name: %s, Phone: %s, Email: %s, ProvinceID: %s, CityID: %s, PostCode: %s, IsPrimary: %t",
// 		form.Name, form.Phone, form.Email, form.ProvinceID, form.CityID, form.PostCode)

// 	if err := h.validator.Struct(&form); err != nil {
// 		validationErrors := err.(validator.ValidationErrors)
// 		formattedErrors := helpers.FormatValidationErrors(validationErrors)

// 		provinces, pErr := h.rajaOngkirSvc.GetProvincesFromAPI()
// 		if pErr != nil {
// 			log.Printf("EditAddressPost: Failed to get provinces on validation error: %v", pErr)
// 			provinces = []other.Province{}
// 		}

// 		cities := []other.City{}
// 		var currentCityName string
// 		if form.ProvinceID != "" {
// 			c, cErr := h.rajaOngkirSvc.GetCitiesFromAPI(form.ProvinceID)
// 			if cErr != nil {
// 				log.Printf("EditAddressPost: Failed to get cities on validation error for province %s: %v", form.ProvinceID, cErr)
// 			} else {
// 				cities = c
// 				for _, city := range cities {
// 					if city.ID == form.CityID {
// 						currentCityName = city.Name
// 						break
// 					}
// 				}
// 			}
// 		}
// 		form.CityName = currentCityName

// 		data := AddressPageData{
// 			FormAction:  fmt.Sprintf("/addresses/edit/%s", addressID),
// 			IsEdit:      true,
// 			AddressData: &form,
// 			Provinces:   provinces,
// 			Cities:      cities,
// 			Errors:      formattedErrors,
// 		}
// 		h.populateBaseData(r, &data)

// 		data.Title = "Edit Alamat"
// 		data.Breadcrumbs = []breadcrumb.Breadcrumb{
// 			{Name: "Beranda", URL: "/"},
// 			{Name: "Profile", URL: "/profile"},
// 			{Name: "Alamat Saya", URL: "/addresses"},
// 			{Name: "Edit Alamat", URL: fmt.Sprintf("/addresses/edit/%s", addressID)},
// 		}
// 		data.IsAuthPage = true

// 		h.render.HTML(w, http.StatusOK, "auth/addresses_form", data)
// 		return
// 	}

// 	address.Name = form.Name
// 	address.Phone = form.Phone
// 	address.Email = form.Email
// 	address.Address1 = form.Address1
// 	address.Address2 = form.Address2
// 	address.ProvinceID = form.ProvinceID
// 	address.CityID = form.CityID
// 	address.PostCode = form.PostCode

// 	address.UpdatedAt = time.Now()

// 	err = h.addressRepo.UpdateAddress(ctx, address)
// 	if err != nil {
// 		log.Printf("EditAddressPost: Failed to update address %s for user %s: %v", addressID, userID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses/edit/%s?status=error&message=%s", addressID, url.QueryEscape("Gagal memperbarui alamat: "+err.Error())), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil diperbarui!")), http.StatusSeeOther)
// }

// func (h *AddressHandler) DeleteAddressPost(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	vars := mux.Vars(r)
// 	addressID := vars["id"]

// 	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
// 	if err != nil || address == nil || address.UserID != userID {
// 		log.Printf("DeleteAddressPost: Address %s not found or unauthorized for user %s for deletion", addressID, userID)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau tidak berhak menghapus.")), http.StatusSeeOther)
// 		return
// 	}

// 	err = h.addressRepo.DeleteAddress(ctx, addressID)
// 	if err != nil {
// 		log.Printf("DeleteAddressPost: Failed to delete address %s for user %s: %v", addressID, userID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Gagal menghapus alamat.")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil dihapus!")), http.StatusSeeOther)
// }

// func (h *AddressHandler) SetPrimaryAddressPost(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID := ctx.Value(helpers.ContextKeyUserID).(string)

// 	vars := mux.Vars(r)
// 	addressID := vars["id"]

// 	address, err := h.addressRepo.FindAddressByID(ctx, addressID)
// 	if err != nil || address == nil || address.UserID != userID {
// 		log.Printf("SetPrimaryAddressPost: Address %s not found or unauthorized for user %s for setting primary", addressID, userID)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Alamat tidak ditemukan atau tidak berhak mengubah.")), http.StatusSeeOther)
// 		return
// 	}

// 	err = h.addressRepo.SetPrimaryAddress(ctx, userID, addressID)
// 	if err != nil {
// 		log.Printf("SetPrimaryAddressPost: Failed to set address %s as primary for user %s: %v", addressID, userID, err)
// 		http.Redirect(w, r, fmt.Sprintf("/addresses?status=error&message=%s", url.QueryEscape("Gagal mengatur alamat sebagai utama.")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/addresses?status=success&message=%s", url.QueryEscape("Alamat berhasil diatur sebagai utama!")), http.StatusSeeOther)
// }

// func (h *AddressHandler) GetCitiesByProvinceIDAPI(w http.ResponseWriter, r *http.Request) {
// 	provinceID := r.URL.Query().Get("province_id")
// 	log.Printf("GetCitiesByProvinceIDAPI: Received request for province_id: %s", provinceID)

// 	if provinceID == "" {
// 		log.Println("GetCitiesByProvinceIDAPI: Province ID is empty, returning 400.")
// 		h.render.JSON(w, http.StatusBadRequest, map[string]string{"error": "Province ID is required"})
// 		return
// 	}

// 	cities, err := h.rajaOngkirSvc.GetCitiesFromAPI(provinceID)
// 	if err != nil {
// 		log.Printf("GetCitiesByProvinceIDAPI: Error fetching cities for province %s from service: %v", provinceID, err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cities"})
// 		return
// 	}

// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"cities": cities,
// 	})
// }
