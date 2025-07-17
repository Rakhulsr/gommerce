package handlers

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"strconv"

// 	"github.com/Rakhulsr/go-ecommerce/app/configs"
// 	"github.com/Rakhulsr/go-ecommerce/app/helpers"
// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/Rakhulsr/go-ecommerce/app/models/other"
// 	"github.com/Rakhulsr/go-ecommerce/app/repositories"
// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
// 	"github.com/shopspring/decimal"
// 	"github.com/unrolled/render"
// )

// type CartHandler struct {
// 	productRepo  repositories.ProductRepositoryImpl
// 	cartRepo     repositories.CartRepositoryImpl
// 	cartItemRepo repositories.CartItemRepositoryImpl
// 	render       *render.Render
// 	locationSvc  services.RajaOngkirClient
// 	userRepo     repositories.UserRepositoryImpl
// 	addressRepo  repositories.AddressRepository
// 	cartSvc      *services.CartService
// }

// func NewCartHandler(
// 	productRepo repositories.ProductRepositoryImpl,
// 	cartRepo repositories.CartRepositoryImpl,
// 	render *render.Render,
// 	cartItemRepo repositories.CartItemRepositoryImpl,
// 	locationSvc services.RajaOngkirClient,
// 	userRepo repositories.UserRepositoryImpl,
// 	addressRepo repositories.AddressRepository,

// 	cartSvc *services.CartService,
// ) *CartHandler {
// 	return &CartHandler{
// 		productRepo:  productRepo,
// 		cartRepo:     cartRepo,
// 		cartItemRepo: cartItemRepo,
// 		render:       render,
// 		locationSvc:  locationSvc,
// 		userRepo:     userRepo,
// 		addressRepo:  addressRepo,
// 		cartSvc:      cartSvc,
// 	}
// }

// func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !userOk || userID == "" {
// 		log.Printf("GetCart: UserID not found in context. Rendering empty cart for non-logged in user.")
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong. Mohon login untuk menyimpan keranjang Anda.")
// 		return
// 	}

// 	cart, err := h.cartSvc.GetUserCart(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("GetCart: Gagal mengambil data cart untuk user %s: %v", userID, err)
// 		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
// 		return
// 	}

// 	if cart == nil || len(cart.CartItems) == 0 {
// 		log.Printf("GetCart: Cart for user %s is empty or not found after service call. Rendering empty cart.", userID)
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
// 		return
// 	}

// 	status := r.URL.Query().Get("status")
// 	message := r.URL.Query().Get("message")

// 	provinces, err := h.locationSvc.GetProvincesFromAPI()
// 	if err != nil {
// 		log.Printf("GetCart: Gagal mengambil daftar provinsi dari RajaOngkir API: %v", err)
// 		status = "error"
// 		message = "Gagal memuat daftar provinsi untuk pengiriman. Coba lagi nanti."
// 		provinces = []other.Province{}
// 	}

// 	var userAddresses []models.Address
// 	userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("GetCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
// 	} else if userWithAddresses != nil {
// 		userAddresses = userWithAddresses.Address

// 		cityMap := make(map[string]other.City)
// 		for _, addr := range userAddresses {
// 			if addr.ProvinceID != "" {
// 				citiesInProvince, err := h.locationSvc.GetCitiesFromAPI(addr.ProvinceID)
// 				if err != nil {
// 					log.Printf("GetCart: Failed to fetch cities for province %s: %v", addr.ProvinceID, err)
// 					continue
// 				}
// 				for _, city := range citiesInProvince {
// 					cityMap[city.ID] = city
// 				}
// 			}
// 		}

// 		provinceMap := make(map[string]other.Province)
// 		for _, prov := range provinces {
// 			provinceMap[prov.ID] = prov
// 		}

// 		for i := range userAddresses {
// 			if prov, found := provinceMap[userAddresses[i].ProvinceID]; found {
// 				userAddresses[i].ProvinceName = prov.Name
// 			} else {
// 				log.Printf("GetCart: Province name not found for ID %s", userAddresses[i].ProvinceID)
// 			}

// 			if city, found := cityMap[userAddresses[i].CityID]; found {
// 				userAddresses[i].CityName = fmt.Sprintf("%s %s", city.Type, city.Name)
// 			} else {
// 				log.Printf("GetCart: City name not found for ID %s", userAddresses[i].CityID)
// 			}
// 		}
// 	} else {
// 		log.Printf("GetCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
// 	}

// 	supportedCouriers := []other.Courier{
// 		{Code: "jne", Name: "JNE"},
// 		{Code: "tiki", Name: "TIKI"},
// 		{Code: "pos", Name: "POS"},
// 	}

// 	originCity := configs.LoadENV.API_ONGKIR_ORIGIN

// 	pageSpecificData := map[string]interface{}{
// 		"title":                 "Keranjang Belanja",
// 		"cart":                  cart,
// 		"totalWeight":           cart.TotalWeight,
// 		"grandTotal":            cart.GrandTotal,
// 		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
// 		"MessageStatus":         status,
// 		"Message":               message,
// 		"provinces":             provinces,
// 		"couriers":              supportedCouriers,
// 		"OriginCityID":          originCity,
// 		"finalPrice":            cart.GrandTotal,
// 		"GrandTotalAmountForJS": cart.GrandTotal.InexactFloat64(),
// 		"Addresses":             userAddresses,
// 	}

// 	datas := helpers.GetBaseData(r, pageSpecificData)
// 	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
// }

// func (h *CartHandler) renderEmptyCart(w http.ResponseWriter, r *http.Request, status, message string) {
// 	emptyCart := &models.Cart{
// 		BaseTotalPrice:  decimal.Zero,
// 		TaxAmount:       decimal.Zero,
// 		TaxPercent:      decimal.Zero,
// 		DiscountAmount:  decimal.Zero,
// 		DiscountPercent: decimal.Zero,
// 		GrandTotal:      decimal.Zero,
// 		TotalWeight:     0,
// 		CartItems:       []models.CartItem{},
// 	}

// 	provinces, err := h.locationSvc.GetProvincesFromAPI()
// 	if err != nil {
// 		log.Printf("renderEmptyCart: Gagal mengambil daftar provinsi dari RajaOngkir API: %v", err)
// 		status = "error"
// 		message = "Gagal memuat daftar provinsi untuk pengiriman. Coba lagi nanti."
// 		provinces = []other.Province{}
// 	}

// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	var userAddresses []models.Address
// 	if userOk && userID != "" {
// 		userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(r.Context(), userID)
// 		if err != nil {
// 			log.Printf("renderEmptyCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
// 		} else if userWithAddresses != nil {
// 			userAddresses = userWithAddresses.Address

// 			cityMap := make(map[string]other.City)
// 			for _, addr := range userAddresses {
// 				if addr.ProvinceID != "" {
// 					citiesInProvince, err := h.locationSvc.GetCitiesFromAPI(addr.ProvinceID)
// 					if err != nil {
// 						log.Printf("renderEmptyCart: Failed to fetch cities for province %s: %v", addr.ProvinceID, err)
// 						continue
// 					}
// 					for _, city := range citiesInProvince {
// 						cityMap[city.ID] = city
// 					}
// 				}
// 			}

// 			provinceMap := make(map[string]other.Province)
// 			for _, prov := range provinces {
// 				provinceMap[prov.ID] = prov
// 			}

// 			for i := range userAddresses {
// 				if prov, found := provinceMap[userAddresses[i].ProvinceID]; found {
// 					userAddresses[i].ProvinceName = prov.Name
// 				} else {
// 					log.Printf("renderEmptyCart: Province name not found for ID %s", userAddresses[i].ProvinceID)
// 				}

// 				if city, found := cityMap[userAddresses[i].CityID]; found {
// 					userAddresses[i].CityName = fmt.Sprintf("%s %s", city.Type, city.Name)
// 				} else {
// 					log.Printf("renderEmptyCart: City name not found for ID %s", userAddresses[i].CityID)
// 				}
// 			}
// 		}
// 	} else {
// 		log.Printf("renderEmptyCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
// 	}

// 	supportedCouriers := []other.Courier{
// 		{Code: "jne", Name: "JNE"},
// 		{Code: "tiki", Name: "TIKI"},
// 		{Code: "pos", Name: "POS"},
// 	}

// 	originCity := configs.LoadENV.API_ONGKIR_ORIGIN

// 	pageSpecificData := map[string]interface{}{
// 		"title":                 "Keranjang Belanja",
// 		"cart":                  emptyCart,
// 		"totalWeight":           emptyCart.TotalWeight,
// 		"grandTotal":            emptyCart.GrandTotal,
// 		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
// 		"MessageStatus":         status,
// 		"Message":               message,
// 		"provinces":             provinces,
// 		"couriers":              supportedCouriers,
// 		"OriginCityID":          originCity,
// 		"finalPrice":            emptyCart.GrandTotal,
// 		"GrandTotalAmountForJS": emptyCart.GrandTotal.InexactFloat64(),
// 		"Addresses":             userAddresses,
// 	}

// 	datas := helpers.GetBaseData(r, pageSpecificData)
// 	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
// }

// func (h *CartHandler) AddItemCart(w http.ResponseWriter, r *http.Request) {
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Gagal membaca data", http.StatusBadRequest)
// 		return
// 	}

// 	productID := r.FormValue("product_id")
// 	qtyStr := r.FormValue("qty")
// 	action := r.FormValue("action")

// 	log.Println("AddItemCart - Product ID:", productID)
// 	log.Println("AddItemCart - Qty:", qtyStr)
// 	log.Println("AddItemCart - Action:", action)

// 	if productID == "" || qtyStr == "" {
// 		log.Printf("AddItemCart: Data tidak lengkap (productID: '%s', qtyStr: '%s')", productID, qtyStr)
// 		redirectBackWithError(w, r, productID, "Data produk atau kuantitas tidak lengkap.", "error", h.productRepo)
// 		return
// 	}

// 	qty, err := strconv.Atoi(qtyStr)
// 	if err != nil || qty <= 0 {
// 		log.Printf("AddItemCart: Jumlah tidak valid (qtyStr: '%s', error: %v)", qtyStr, err)
// 		redirectBackWithError(w, r, productID, "Jumlah tidak valid, harus lebih dari 0.", "error", h.productRepo)
// 		return
// 	}

// 	product, err := h.productRepo.GetByID(r.Context(), productID)
// 	if err != nil || product == nil {
// 		log.Printf("AddItemCart: Produk tidak ditemukan: %v", err)
// 		redirectBackWithError(w, r, productID, "Produk tidak ditemukan.", "error", h.productRepo)
// 		return
// 	}

// 	cartID, _ := r.Context().Value(helpers.ContextKeyCartID).(string)
// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)

// 	if !userOk || userID == "" {

// 		log.Printf("AddItemCart: UserID not found in context. Redirecting to login.")
// 		redirectBackWithError(w, r, productID, "Anda harus login untuk menambahkan produk ke keranjang.", "warning", h.productRepo)
// 		return
// 	}

// 	err = h.cartSvc.AddItemToCart(r.Context(), cartID, userID, productID, qty)
// 	if err != nil {
// 		log.Printf("AddItemCart: Gagal menambahkan item ke keranjang melalui service: %v", err)
// 		redirectBackWithError(w, r, productID, fmt.Sprintf("Gagal menambahkan produk ke keranjang: %v", err), "error", h.productRepo)
// 		return
// 	}

// 	switch action {
// 	case "buy":
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
// 		return
// 	default:
// 		http.Redirect(w, r, fmt.Sprintf("/products/%s?status=success&message=%s", product.Slug, url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
// 	}
// }

// func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
// 	productID := r.FormValue("product_id")
// 	qtyStr := r.FormValue("qty")

// 	qty, err := strconv.Atoi(qtyStr)
// 	if err != nil || qty <= 0 {
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Kuantitas tidak valid!")), http.StatusSeeOther)
// 		return
// 	}

// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !userOk || userID == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	updatedCart, err := h.cartSvc.UpdateCartItemQty(r.Context(), userID, productID, qty)
// 	if err != nil {
// 		log.Printf("UpdateCartItem: Gagal memperbarui item keranjang melalui service: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal memperbarui item: %v", err))), http.StatusSeeOther)
// 		return
// 	}

// 	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus atau kuantitas diubah menjadi nol!")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Kuantitas item keranjang berhasil diperbarui!")), http.StatusSeeOther)
// }

// func (h *CartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
// 	productID := r.FormValue("product_id")
// 	if productID == "" {
// 		http.Error(w, "Produk tidak valid", http.StatusBadRequest)
// 		return
// 	}

// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !userOk || userID == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	updatedCart, err := h.cartSvc.RemoveItemFromCart(r.Context(), userID, productID)
// 	if err != nil {
// 		log.Printf("DeleteCartItem: Gagal menghapus item keranjang melalui service: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal menghapus item: %v", err))), http.StatusSeeOther)
// 		return
// 	}

// 	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus dan keranjang kosong!")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item keranjang berhasil dihapus!")), http.StatusSeeOther)
// }

// func (h *CartHandler) GetCartCount(w http.ResponseWriter, r *http.Request) {
// 	if cartCountVal := r.Context().Value(helpers.CartCountKey); cartCountVal != nil {
// 		if count, ok := cartCountVal.(int); ok {
// 			w.Write([]byte(strconv.Itoa(count)))
// 			return
// 		}
// 	}
// 	w.Write([]byte("0"))
// }

// func redirectBackWithError(w http.ResponseWriter, r *http.Request, productID string, msg string, status string, productRepo repositories.ProductRepositoryImpl) {
// 	if productID != "" {
// 		product, err := productRepo.GetByID(r.Context(), productID)
// 		if err == nil && product != nil {
// 			http.Redirect(w, r, fmt.Sprintf("/products/%s?status=%s&message=%s", product.Slug, status, url.QueryEscape(msg)), http.StatusSeeOther)
// 			return
// 		}
// 	}
// 	http.Redirect(w, r, fmt.Sprintf("/?status=%s&message=%s", status, url.QueryEscape(msg)), http.StatusSeeOther)
// }
