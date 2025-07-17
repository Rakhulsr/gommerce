package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

type CalculateShippingCostRequest struct {
	Origin      int    `json:"origin"`
	Destination int    `json:"destination"`
	Weight      int    `json:"weight"`
	Courier     string `json:"courier"`
}
type KomerceCartHandler struct {
	productRepo        repositories.ProductRepositoryImpl
	cartRepo           repositories.CartRepositoryImpl
	cartItemRepo       repositories.CartItemRepositoryImpl
	render             *render.Render
	komerceLocationSvc services.KomerceRajaOngkirClient
	userRepo           repositories.UserRepositoryImpl
	addressRepo        repositories.AddressRepository
	cartSvc            *services.CartService
	merchantOriginID   int
}

func NewKomerceCartHandler(
	productRepo repositories.ProductRepositoryImpl,
	cartRepo repositories.CartRepositoryImpl,
	render *render.Render,
	cartItemRepo repositories.CartItemRepositoryImpl,
	komerceLocationSvc services.KomerceRajaOngkirClient,
	userRepo repositories.UserRepositoryImpl,
	addressRepo repositories.AddressRepository,
	cartSvc *services.CartService,
	merchantOriginID int,
) *KomerceCartHandler {
	return &KomerceCartHandler{
		productRepo:        productRepo,
		cartRepo:           cartRepo,
		cartItemRepo:       cartItemRepo,
		render:             render,
		komerceLocationSvc: komerceLocationSvc,
		userRepo:           userRepo,
		addressRepo:        addressRepo,
		cartSvc:            cartSvc,
		merchantOriginID:   merchantOriginID,
	}
}

func (h *KomerceCartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, userOk := ctx.Value(helpers.ContextKeyUserID).(string)
	if !userOk || userID == "" {
		log.Printf("KomerceCartHandler.GetCart: UserID not found in context. Rendering empty cart for non-logged in user.")
		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong. Mohon login untuk menyimpan keranjang Anda.")
		return
	}

	cart, err := h.cartSvc.GetUserCart(ctx, userID)
	if err != nil {
		log.Printf("KomerceCartHandler.GetCart: Gagal mengambil data cart untuk user %s: %v", userID, err)
		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
		return
	}

	if cart == nil || len(cart.CartItems) == 0 {
		log.Printf("KomerceCartHandler.GetCart: Cart for user %s is empty or not found after service call. Rendering empty cart.", userID)
		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
		return
	}

	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	var userAddresses []models.Address
	userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(ctx, userID)
	if err != nil {
		log.Printf("KomerceCartHandler.GetCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
	} else if userWithAddresses != nil {
		userAddresses = userWithAddresses.Address
	} else {
		log.Printf("KomerceCartHandler.GetCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
	}

	supportedCouriers := []other.Courier{
		{Code: "jne", Name: "JNE"},
		{Code: "pos", Name: "POS"},
		{Code: "jnt", Name: "J&T Express"},
	}

	originLocationIDStr := strconv.Itoa(h.merchantOriginID)
	if h.merchantOriginID == 0 {
		log.Println("KomerceCartHandler.GetCart: Merchant Origin ID is 0, using default Depok (25986). Check .env config.")
		originLocationIDStr = "25986"
	}

	pageSpecificData := map[string]interface{}{
		"title":                 "Keranjang Belanja",
		"cart":                  cart,
		"totalWeight":           cart.TotalWeight,
		"baseTotalPrice":        cart.BaseTotalPrice,
		"totalDiscountAmount":   cart.DiscountAmount,
		"taxAmount":             cart.TaxAmount,
		"taxPercent":            cart.TaxPercent,
		"grandTotal":            cart.GrandTotal,
		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Carts", URL: "/carts"}},
		"MessageStatus":         status,
		"Message":               message,
		"couriers":              supportedCouriers,
		"OriginLocationID":      originLocationIDStr,
		"finalPrice":            cart.GrandTotal,
		"GrandTotalAmountForJS": cart.GrandTotal.InexactFloat64(),
		"Addresses":             userAddresses,
	}

	datas := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
}

func (h *KomerceCartHandler) renderEmptyCart(w http.ResponseWriter, r *http.Request, status, message string) {
	ctx := r.Context()
	emptyCart := &models.Cart{
		BaseTotalPrice:  decimal.Zero,
		TaxAmount:       decimal.Zero,
		TaxPercent:      calc.GetTaxPercent(),
		DiscountAmount:  decimal.Zero,
		DiscountPercent: decimal.Zero,
		GrandTotal:      decimal.Zero,
		TotalWeight:     decimal.Zero,
		ShippingCost:    decimal.Zero,
		TotalItems:      0,
		CartItems:       []models.CartItem{},
	}

	userID, userOk := ctx.Value(helpers.ContextKeyUserID).(string)
	var userAddresses []models.Address
	if userOk && userID != "" {
		userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(ctx, userID)
		if err != nil {
			log.Printf("KomerceCartHandler.renderEmptyCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
		} else if userWithAddresses != nil {
			userAddresses = userWithAddresses.Address
		}
	} else {
		log.Printf("KomerceCartHandler.renderEmptyCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
	}

	supportedCouriers := []other.Courier{
		{Code: "jne", Name: "JNE"},
		{Code: "tiki", Name: "TIKI"},
		{Code: "pos", Name: "POS"},
		{Code: "jnt", Name: "J&T Express"},
		{Code: "sicepat", Name: "SiCepat"},
		{Code: "anteraja", Name: "AnterAja"},
	}

	originLocationIDStr := strconv.Itoa(h.merchantOriginID)
	if h.merchantOriginID == 0 {
		log.Println("KomerceCartHandler.renderEmptyCart: Merchant Origin ID is 0, using default Depok (25986). Check .env config.")
		originLocationIDStr = "25986"
	}

	pageSpecificData := map[string]interface{}{
		"title":                 "Keranjang Belanja",
		"cart":                  emptyCart,
		"totalWeight":           emptyCart.TotalWeight,
		"baseTotalPrice":        emptyCart.BaseTotalPrice,
		"totalDiscountAmount":   emptyCart.DiscountAmount,
		"taxAmount":             emptyCart.TaxAmount,
		"taxPercent":            emptyCart.TaxPercent,
		"grandTotal":            emptyCart.GrandTotal,
		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
		"MessageStatus":         status,
		"Message":               message,
		"couriers":              supportedCouriers,
		"OriginLocationID":      originLocationIDStr,
		"finalPrice":            emptyCart.GrandTotal,
		"GrandTotalAmountForJS": emptyCart.GrandTotal.InexactFloat64(),
		"Addresses":             userAddresses,
	}

	datas := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
}

func (h *KomerceCartHandler) AddItemCart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal membaca data", http.StatusBadRequest)
		return
	}

	productID := r.FormValue("product_id")
	qtyStr := r.FormValue("qty")
	action := r.FormValue("action")

	log.Println("KomerceCartHandler.AddItemCart - Product ID:", productID)
	log.Println("KomerceCartHandler.AddItemCart - Qty:", qtyStr)
	log.Println("KomerceCartHandler.AddItemCart - Action:", action)

	if productID == "" || qtyStr == "" {
		log.Printf("KomerceCartHandler.AddItemCart: Data tidak lengkap (productID: '%s', qtyStr: '%s')", productID, qtyStr)
		redirectBackWithError(w, r, productID, "Data produk atau kuantitas tidak lengkap.", "error", h.productRepo)
		return
	}

	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		log.Printf("KomerceCartHandler.AddItemCart: Jumlah tidak valid (qtyStr: '%s', error: %v)", qtyStr, err)
		redirectBackWithError(w, r, productID, "Jumlah tidak valid, harus lebih dari 0.", "error", h.productRepo)
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil || product == nil {
		log.Printf("KomerceCartHandler.AddItemCart: Produk tidak ditemukan: %v", err)
		redirectBackWithError(w, r, productID, "Produk tidak ditemukan.", "error", h.productRepo)
		return
	}

	cartID, _ := r.Context().Value(helpers.ContextKeyCartID).(string)
	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)

	if !userOk || userID == "" {
		log.Printf("KomerceCartHandler.AddItemCart: UserID not found in context. Redirecting to login.")
		redirectBackWithError(w, r, productID, "Anda harus login untuk menambahkan produk ke keranjang.", "warning", h.productRepo)
		return
	}

	err = h.cartSvc.AddItemToCart(r.Context(), cartID, userID, productID, qty)
	if err != nil {
		log.Printf("KomerceCartHandler.AddItemCart: Gagal menambahkan item ke keranjang melalui service: %v", err)
		redirectBackWithError(w, r, productID, fmt.Sprintf("Gagal menambahkan produk ke keranjang: %v", err), "error", h.productRepo)
		return
	}

	switch action {
	case "buy":
		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
		return
	default:
		http.Redirect(w, r, fmt.Sprintf("/products/%s?status=success&message=%s", product.Slug, url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
	}
}

func (h *KomerceCartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	qtyStr := r.FormValue("qty")

	qty, err := strconv.Atoi(qtyStr)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Kuantitas tidak valid!")), http.StatusSeeOther)
		return
	}

	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !userOk || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updatedCart, err := h.cartSvc.UpdateCartItemQty(r.Context(), userID, productID, qty)
	if err != nil {
		log.Printf("KomerceCartHandler.UpdateCartItem: Gagal memperbarui item keranjang melalui service: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal memperbarui item: %v", err))), http.StatusSeeOther)
		return
	}

	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus atau kuantitas diubah menjadi nol!")), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Kuantitas item keranjang berhasil diperbarui!")), http.StatusSeeOther)
}

func (h *KomerceCartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	if productID == "" {
		http.Error(w, "Produk tidak valid", http.StatusBadRequest)
		return
	}

	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !userOk || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updatedCart, err := h.cartSvc.RemoveItemFromCart(r.Context(), userID, productID)
	if err != nil {
		log.Printf("KomerceCartHandler.DeleteCartItem: Gagal menghapus item keranjang melalui service: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal menghapus item: %v", err))), http.StatusSeeOther)
		return
	}

	if updatedCart == nil || len(updatedCart.CartItems) == 0 {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil dihapus dan keranjang kosong!")), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item keranjang berhasil dihapus!")), http.StatusSeeOther)
}

func (h *KomerceCartHandler) CalculateShippingCost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Println("CalculateShippingCost: UserID not found in context.")
		h.render.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("CalculateShippingCost: Gagal membaca raw request body: %v", err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Failed to read request body.",
		})
		return
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Printf("CalculateShippingCost: Raw Request Body: %s", string(bodyBytes))

	var reqBody CalculateShippingCostRequest
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Printf("CalculateShippingCost: Gagal decode request body JSON: %v", err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body format.",
		})
		return
	}

	log.Printf("CalculateShippingCost: Decoded Origin: %d", reqBody.Origin)
	log.Printf("CalculateShippingCost: Decoded Destination: %d", reqBody.Destination)
	log.Printf("CalculateShippingCost: Decoded Weight: %d", reqBody.Weight)
	log.Printf("CalculateShippingCost: Decoded Courier: %s", reqBody.Courier)

	originID := h.merchantOriginID
	destinationID := reqBody.Destination
	weight := reqBody.Weight
	courier := reqBody.Courier

	if originID == 0 {
		log.Println("CalculateShippingCost: Merchant Origin ID is not configured (0).")
		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Merchant origin ID not configured.",
		})
		return
	}

	if destinationID == 0 {
		log.Println("CalculateShippingCost: Destination ID is missing or invalid (0).")
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Destination ID is required and must be a valid number.",
		})
		return
	}
	if weight <= 0 {
		log.Println("CalculateShippingCost: Weight is missing or invalid (<= 0).")
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Weight must be greater than 0.",
		})
		return
	}
	if courier == "" {
		log.Println("CalculateShippingCost: Courier is empty.")
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Courier is required.",
		})
		return
	}

	shippingCosts, err := h.komerceLocationSvc.CalculateCost(ctx, originID, destinationID, weight, courier)
	if err != nil {
		log.Printf("CalculateShippingCost: Gagal menghitung biaya pengiriman melalui service Komerce: %v", err)
		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Gagal menghitung biaya pengiriman: %v", err),
		})
		return
	}

	h.render.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    shippingCosts,
	})
}

func (h *KomerceCartHandler) GetCartCount(w http.ResponseWriter, r *http.Request) {
	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !userOk || userID == "" {
		w.Write([]byte("0"))
		return
	}

	cart, err := h.cartSvc.GetUserCart(r.Context(), userID)
	if err != nil {
		log.Printf("GetCartCount: Gagal mengambil cart untuk userID %s: %v", userID, err)
		w.Write([]byte("0"))
		return
	}

	if cart == nil {
		w.Write([]byte("0"))
		return
	}

	w.Write([]byte(strconv.Itoa(cart.TotalItems)))
}

func redirectBackWithError(w http.ResponseWriter, r *http.Request, productID string, msg string, status string, productRepo repositories.ProductRepositoryImpl) {
	if productID != "" {
		product, err := productRepo.GetByID(r.Context(), productID)
		if err == nil && product != nil {
			http.Redirect(w, r, fmt.Sprintf("/products/%s?status=%s&message=%s", product.Slug, status, url.QueryEscape(msg)), http.StatusSeeOther)
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/?status=%s&message=%s", status, url.QueryEscape(msg)), http.StatusSeeOther)
}
