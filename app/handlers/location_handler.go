package handlers

// import (
// 	"encoding/json" // Import for JSON handling
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"os" // Import for os.Getenv
// 	"strconv"

// 	"github.com/Rakhulsr/go-ecommerce/app/helpers"
// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/Rakhulsr/go-ecommerce/app/models/other"
// 	"github.com/Rakhulsr/go-ecommerce/app/repositories"
// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
// 	"github.com/shopspring/decimal"
// 	"github.com/unrolled/render"
// )

// type KomerceCartHandler struct {
// 	productRepo        repositories.ProductRepositoryImpl
// 	cartRepo           repositories.CartRepositoryImpl
// 	cartItemRepo       repositories.CartItemRepositoryImpl
// 	render             *render.Render
// 	komerceLocationSvc services.KomerceRajaOngkirClient // Ini adalah client untuk RajaOngkir
// 	userRepo           repositories.UserRepositoryImpl
// 	addressRepo        repositories.AddressRepository
// 	cartSvc            *services.CartService
// }

// func NewKomerceCartHandler(
// 	productRepo repositories.ProductRepositoryImpl,
// 	cartRepo repositories.CartRepositoryImpl,
// 	render *render.Render,
// 	cartItemRepo repositories.CartItemRepositoryImpl,
// 	komerceLocationSvc services.KomerceRajaOngkirClient,
// 	userRepo repositories.UserRepositoryImpl,
// 	addressRepo repositories.AddressRepository,
// 	cartSvc *services.CartService,
// ) *KomerceCartHandler {
// 	return &KomerceCartHandler{
// 		productRepo:        productRepo,
// 		cartRepo:           cartRepo,
// 		cartItemRepo:       cartItemRepo,
// 		render:             render,
// 		komerceLocationSvc: komerceLocationSvc,
// 		userRepo:           userRepo,
// 		addressRepo:        addressRepo,
// 		cartSvc:            cartSvc,
// 	}
// }

// func (h *KomerceCartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userID, userOk := ctx.Value(helpers.ContextKeyUserID).(string)
// 	if !userOk || userID == "" {
// 		log.Printf("KomerceCartHandler.GetCart: UserID not found in context. Rendering empty cart for non-logged in user.")
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong. Mohon login untuk menyimpan keranjang Anda.")
// 		return
// 	}

// 	cart, err := h.cartSvc.GetUserCart(ctx, userID)
// 	if err != nil {
// 		log.Printf("KomerceCartHandler.GetCart: Gagal mengambil data cart untuk user %s: %v", userID, err)
// 		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
// 		return
// 	}

// 	if cart == nil || len(cart.CartItems) == 0 {
// 		log.Printf("KomerceCartHandler.GetCart: Cart for user %s is empty or not found after service call. Rendering empty cart.", userID)
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
// 		return
// 	}

// 	status := r.URL.Query().Get("status")
// 	message := r.URL.Query().Get("message")

// 	var userAddresses []models.Address
// 	userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(ctx, userID)
// 	if err != nil {
// 		log.Printf("KomerceCartHandler.GetCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
// 	} else if userWithAddresses != nil {
// 		userAddresses = userWithAddresses.Address
// 	} else {
// 		log.Printf("KomerceCartHandler.GetCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
// 	}

// 	supportedCouriers := []other.Courier{
// 		{Code: "jne", Name: "JNE"},
// 		{Code: "tiki", Name: "TIKI"},
// 		{Code: "pos", Name: "POS"},
// 	}

// 	// Ambil OriginLocationID dari environment variable
// 	originLocationID := os.Getenv("RAJAONGKIR_ORIGIN_LOCATION_ID")
// 	if originLocationID == "" {
// 		originLocationID = "25986" // Default ke Depok (contoh) jika tidak diatur
// 		log.Println("RAJAONGKIR_ORIGIN_LOCATION_ID tidak diatur, menggunakan default Depok (25986)")
// 	}

// 	pageSpecificData := map[string]interface{}{
// 		"title":                 "Keranjang Belanja",
// 		"cart":                  cart,
// 		"totalWeight":           cart.TotalWeight,
// 		"baseTotalPrice":        cart.BaseTotalPrice,
// 		"totalDiscountAmount":   cart.DiscountAmount,
// 		"taxAmount":             cart.TaxAmount,
// 		"taxPercent":            cart.TaxPercent,
// 		"grandTotal":            cart.GrandTotal,
// 		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Carts", URL: "/carts"}},
// 		"MessageStatus":         status,
// 		"Message":               message,
// 		"couriers":              supportedCouriers,
// 		"OriginLocationID":      originLocationID,
// 		"finalPrice":            cart.GrandTotal,
// 		"GrandTotalAmountForJS": cart.GrandTotal.InexactFloat64(),
// 		"Addresses":             userAddresses,
// 	}

// 	datas := helpers.GetBaseData(r, pageSpecificData)
// 	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
// }

// func (h *KomerceCartHandler) renderEmptyCart(w http.ResponseWriter, r *http.Request, status, message string) {
// 	ctx := r.Context()
// 	emptyCart := &models.Cart{
// 		BaseTotalPrice:  decimal.Zero,
// 		TaxAmount:       decimal.Zero,
// 		TaxPercent:      calc.GetTaxPercent(),
// 		DiscountAmount:  decimal.Zero,
// 		DiscountPercent: decimal.Zero,
// 		GrandTotal:      decimal.Zero,
// 		TotalWeight:     decimal.Zero,
// 		ShippingCost:    decimal.Zero,
// 		TotalItems:      0,
// 		CartItems:       []models.CartItem{},
// 	}

// 	userID, userOk := ctx.Value(helpers.ContextKeyUserID).(string)
// 	var userAddresses []models.Address
// 	if userOk && userID != "" {
// 		userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(ctx, userID)
// 		if err != nil {
// 			log.Printf("KomerceCartHandler.renderEmptyCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
// 		} else if userWithAddresses != nil {
// 			userAddresses = userWithAddresses.Address
// 		}
// 	} else {
// 		log.Printf("KomerceCartHandler.renderEmptyCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
// 	}

// 	supportedCouriers := []other.Courier{
// 		{Code: "jne", Name: "JNE"},
// 		{Code: "tiki", Name: "TIKI"},
// 		{Code: "pos", Name: "POS"},
// 	}

// 	originLocationID := os.Getenv("RAJAONGKIR_ORIGIN_LOCATION_ID")
// 	if originLocationID == "" {
// 		originLocationID = "25986" // Default ke Depok (contoh) jika tidak diatur
// 		log.Println("RAJAONGKIR_ORIGIN_LOCATION_ID tidak diatur, menggunakan default Depok (25986)")
// 	}

// 	pageSpecificData := map[string]interface{}{
// 		"title":                 "Keranjang Belanja",
// 		"cart":                  emptyCart,
// 		"totalWeight":           emptyCart.TotalWeight,
// 		"baseTotalPrice":        emptyCart.BaseTotalPrice,
// 		"totalDiscountAmount":   emptyCart.DiscountAmount,
// 		"taxAmount":             emptyCart.TaxAmount,
// 		"taxPercent":            emptyCart.TaxPercent,
// 		"grandTotal":            emptyCart.GrandTotal,
// 		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
// 		"MessageStatus":         status,
// 		"Message":               message,
// 		"couriers":              supportedCouriers,
// 		"OriginLocationID":      originLocationID,
// 		"finalPrice":            emptyCart.GrandTotal,
// 		"GrandTotalAmountForJS": emptyCart.GrandTotal.InexactFloat64(),
// 		"Addresses":             userAddresses,
// 	}

// 	datas := helpers.GetBaseData(r, pageSpecificData)
// 	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
// }

// func (h *KomerceCartHandler) AddItemCart(w http.ResponseWriter, r *http.Request) {
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Gagal membaca data", http.StatusBadRequest)
// 		return
// 	}

// 	productID := r.FormValue("product_id")
// 	qtyStr := r.FormValue("qty")
// 	action := r.FormValue("action")

// 	log.Println("KomerceCartHandler.AddItemCart - Product ID:", productID)
// 	log.Println("KomerceCartHandler.AddItemCart - Qty:", qtyStr)
// 	log.Println("KomerceCartHandler.AddItemCart - Action:", action)

// 	if productID == "" || qtyStr == "" {
// 		log.Printf("KomerceCartHandler.AddItemCart: Data tidak lengkap (productID: '%s', qtyStr: '%s')", productID, qtyStr)
// 		redirectBackWithError(w, r, productID, "Data produk atau kuantitas tidak lengkap.", "error", h.productRepo)
// 		return
// 	}

// 	qty, err := strconv.Atoi(qtyStr)
// 	if err != nil || qty <= 0 {
// 		log.Printf("KomerceCartHandler.AddItemCart: Jumlah tidak valid (qtyStr: '%s', error: %v)", qtyStr, err)
// 		redirectBackWithError(w, r, productID, "Jumlah tidak valid, harus lebih dari 0.", "error", h.productRepo)
// 		return
// 	}

// 	product, err := h.productRepo.GetByID(r.Context(), productID)
// 	if err != nil || product == nil {
// 		log.Printf("KomerceCartHandler.AddItemCart: Produk tidak ditemukan: %v", err)
// 		redirectBackWithError(w, r, productID, "Produk tidak ditemukan.", "error", h.productRepo)
// 		return
// 	}

// 	cartID, _ := r.Context().Value(helpers.ContextKeyCartID).(string)
// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)

// 	if !userOk || userID == "" {
// 		log.Printf("KomerceCartHandler.AddItemCart: UserID not found in context. Redirecting to login.")
// 		redirectBackWithError(w, r, productID, "Anda harus login untuk menambahkan produk ke keranjang.", "warning", h.productRepo)
// 		return
// 	}

// 	err = h.cartSvc.AddItemToCart(r.Context(), cartID, userID, productID, qty)
// 	if err != nil {
// 		log.Printf("KomerceCartHandler.AddItemCart: Gagal menambahkan item ke keranjang melalui service: %v", err)
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

// func (h *KomerceCartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
// 	productID := r.FormValue("product_id")
// 	qtyStr := r.FormValue("qty")

// 	qty, err := strconv.Atoi(qtyStr)
// 	if err != nil {
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
// 		log.Printf("KomerceCartHandler.UpdateCartItem: Gagal memperbarui item keranjang melalui service: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal memperbarui item: %v", err))), http.StatusSeeOther)
// 		return
// 	}

// 	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus atau kuantitas diubah menjadi nol!")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Kuantitas item keranjang berhasil diperbarui!")), http.StatusSeeOther)
// }

// func (h *KomerceCartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
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
// 		log.Printf("KomerceCartHandler.DeleteCartItem: Gagal menghapus item keranjang melalui service: %v", err)
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal menghapus item: %v", err))), http.StatusSeeOther)
// 		return
// 	}

// 	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
// 		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus dan keranjang kosong!")), http.StatusSeeOther)
// 		return
// 	}

// 	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item keranjang berhasil dihapus!")), http.StatusSeeOther)
// }

// // CalculateShippingCost handles the AJAX request to calculate shipping costs
// func (h *KomerceCartHandler) CalculateShippingCost(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req struct {
// 		OriginID      string  `json:"origin_id"`
// 		DestinationID string  `json:"destination_id"`
// 		Weight        float64 `json:"weight"`
// 		Courier       string  `json:"courier"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		log.Printf("CalculateShippingCost: Gagal decode request body: %v", err)
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validasi input
// 	if req.OriginID == "" || req.DestinationID == "" || req.Weight <= 0 || req.Courier == "" {
// 		log.Printf("CalculateShippingCost: Invalid input: %+v", req)
// 		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 			"success": false,
// 			"message": "Data input tidak lengkap atau tidak valid (origin_id, destination_id, weight, courier diperlukan).",
// 		})
// 		return
// 	}

// 	// Panggil service KomerceRajaOngkirClient untuk menghitung biaya
// 	// Asumsi komerceLocationSvc memiliki metode CalculateCost
// 	costs, err := h.komerceLocationSvc.CalculateCost(req.OriginID, req.DestinationID, int(req.Weight), req.Courier)
// 	if err != nil {
// 		log.Printf("CalculateShippingCost: Gagal menghitung biaya pengiriman melalui service Komerce: %v", err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"success": false,
// 			"message": fmt.Sprintf("Gagal menghitung ongkos kirim: %v", err),
// 		})
// 		return
// 	}

// 	// Kirim respons sukses
// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"success": true,
// 		"message": "Biaya pengiriman berhasil dihitung.",
// 		"costs":   costs,
// 	})
// }

// func (h *KomerceCartHandler) GetCartCount(w http.ResponseWriter, r *http.Request) {
// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !userOk || userID == "" {
// 		w.Write([]byte("0"))
// 		return
// 	}

// 	cart, err := h.cartSvc.GetUserCart(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("GetCartCount: Gagal mengambil cart untuk userID %s: %v", userID, err)
// 		w.Write([]byte("0"))
// 		return
// 	}

// 	if cart == nil {
// 		w.Write([]byte("0"))
// 		return
// 	}

// 	w.Write([]byte(strconv.Itoa(cart.TotalItems)))
// }
