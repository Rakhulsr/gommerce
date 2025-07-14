package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
	"gorm.io/gorm"
)

type CartHandler struct {
	productRepo  repositories.ProductRepositoryImpl
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	render       *render.Render
	locationSvc  *services.RajaOngkirService
}

func NewCartHandler(
	productRepo repositories.ProductRepositoryImpl,
	cartRepo repositories.CartRepositoryImpl,
	render *render.Render,
	cartItemRepo repositories.CartItemRepositoryImpl,
	locationSvc *services.RajaOngkirService,
) *CartHandler {
	return &CartHandler{
		productRepo:  productRepo,
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		render:       render,
		locationSvc:  locationSvc,
	}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
	if !ok || cartID == "" {
		log.Printf("GetCart: CartID not found in context. Rendering empty cart.")
		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
		return
	}

	cart, err := h.cartRepo.GetCartWithItems(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("GetCart: Cart with ID %s not found, potentially invalid ID in session. Rendering empty cart.", cartID)
			h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong atau tidak valid.")
			return
		}
		log.Printf("GetCart: Gagal mengambil data cart untuk ID %s: %v", cartID, err)
		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
		return
	}

	if cart == nil || len(cart.CartItems) == 0 {
		log.Printf("GetCart: Cart %s found but empty. Rendering empty cart.", cartID)
		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
		return
	}

	totalWeight := 0
	grandTotal := decimal.NewFromFloat(0)

	for _, cartItem := range cart.CartItems {

		if cartItem.Product.ID == "" {
			product, err := h.productRepo.GetByID(r.Context(), cartItem.ProductID)
			if err != nil || product == nil {
				log.Printf("GetCart: Product %s not found for cart item %s. Skipping item recalculation.", cartItem.ProductID, cartItem.ID)
				continue
			}
			cartItem.Product = *product
		}

		cartItem.BasePrice = cartItem.Product.Price
		cartItem.BaseTotal = cartItem.BasePrice.Mul(decimal.NewFromInt(int64(cartItem.Qty)))

		itemDiscountAmount := cartItem.Product.DiscountAmount.Mul(decimal.NewFromInt(int64(cartItem.Qty)))

		cartItem.TaxPercent = calc.GetTaxPercent()
		cartItem.TaxAmount = calc.CalculateTax(cartItem.BaseTotal.Sub(itemDiscountAmount))

		cartItem.SubTotal = cartItem.BaseTotal.Sub(itemDiscountAmount)
		cartItem.GrandTotal = cartItem.SubTotal.Add(cartItem.TaxAmount)

		if cartItem.Product.ID != "" {
			productWeigth := cartItem.Product.Weight.InexactFloat64()
			ceilWeight := math.Ceil(productWeigth)
			itemWeight := cartItem.Qty * int(ceilWeight)
			totalWeight += itemWeight

			grandTotal = grandTotal.Add(cartItem.GrandTotal)
		}
	}

	cart.TotalWeight = totalWeight
	cart.GrandTotal = grandTotal

	if err := h.cartRepo.UpdateCartSummary(r.Context(), cart.ID); err != nil {
		log.Printf("GetCart: Gagal update ringkasan cart %s setelah recalculate item: %v", cart.ID, err)
	}

	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	provinces, err := h.locationSvc.GetProvincesFromAPI()
	if err != nil {
		log.Printf("GetCart: Gagal mengambil daftar provinsi dari RajaOngkir API: %v", err)
		status = "error"
		message = "Gagal memuat daftar provinsi untuk pengiriman. Coba lagi nanti."
		provinces = []other.Province{}
	}

	supportedCouriers := []other.Courier{
		{Code: "jne", Name: "JNE"},
		{Code: "tiki", Name: "TIKI"},
	}

	originCity := configs.LoadENV.API_ONGKIR_ORIGIN

	pageSpecificData := map[string]interface{}{
		"title":                 "Keranjang Belanja",
		"cart":                  cart,
		"totalWeight":           cart.TotalWeight,
		"grandTotal":            cart.GrandTotal,
		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
		"MessageStatus":         status,
		"Message":               message,
		"provinces":             provinces,
		"couriers":              supportedCouriers,
		"OriginCityID":          originCity,
		"finalPrice":            cart.GrandTotal,
		"GrandTotalAmountForJS": cart.GrandTotal.IntPart(),
	}

	datas := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
}

func (h *CartHandler) renderEmptyCart(w http.ResponseWriter, r *http.Request, status, message string) {
	emptyCart := &models.Cart{
		BaseTotalPrice:  decimal.Zero,
		TaxAmount:       decimal.Zero,
		TaxPercent:      decimal.Zero,
		DiscountAmount:  decimal.Zero,
		DiscountPercent: decimal.Zero,
		GrandTotal:      decimal.Zero,
		TotalWeight:     0,
		CartItems:       []models.CartItem{},
	}

	provinces, err := h.locationSvc.GetProvincesFromAPI()
	if err != nil {
		log.Printf("renderEmptyCart: Gagal mengambil daftar provinsi dari RajaOngkir API: %v", err)
		status = "error"
		message = "Gagal memuat daftar provinsi untuk pengiriman. Coba lagi nanti."
		provinces = []other.Province{}
	}

	supportedCouriers := []other.Courier{
		{Code: "jne", Name: "JNE"},
		{Code: "tiki", Name: "TIKI"},
	}

	originCity := configs.LoadENV.API_ONGKIR_ORIGIN

	pageSpecificData := map[string]interface{}{
		"title":                 "Keranjang Belanja",
		"cart":                  emptyCart,
		"totalWeight":           emptyCart.TotalWeight,
		"grandTotal":            emptyCart.GrandTotal,
		"Breadcrumbs":           []breadcrumb.Breadcrumb{{Name: "Home", URL: "/"}, {Name: "Keranjang Belanja", URL: "/carts"}},
		"MessageStatus":         status,
		"Message":               message,
		"provinces":             provinces,
		"couriers":              supportedCouriers,
		"OriginCityID":          originCity,
		"finalPrice":            emptyCart.GrandTotal,
		"GrandTotalAmountForJS": emptyCart.GrandTotal.IntPart(),
	}

	datas := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
}

func (h *CartHandler) AddItemCart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal membaca data", http.StatusBadRequest)
		return
	}

	productID := r.FormValue("product_id")
	qtyStr := r.FormValue("qty")
	action := r.FormValue("action")

	log.Println("AddItemCart - Product ID:", productID)
	log.Println("AddItemCart - Qty:", qtyStr)
	log.Println("AddItemCart - Action:", action)

	if productID == "" || qtyStr == "" {
		log.Printf("AddItemCart: Data tidak lengkap (productID: '%s', qtyStr: '%s')", productID, qtyStr)
		redirectBackWithError(w, r, productID, "Data produk atau kuantitas tidak lengkap.", "error", h.productRepo)
		return
	}

	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		log.Printf("AddItemCart: Jumlah tidak valid (qtyStr: '%s', error: %v)", qtyStr, err)
		redirectBackWithError(w, r, productID, "Jumlah tidak valid, harus lebih dari 0.", "error", h.productRepo)
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil || product == nil {
		log.Printf("AddItemCart: Produk tidak ditemukan: %v", err)
		redirectBackWithError(w, r, productID, "Produk tidak ditemukan.", "error", h.productRepo)
		return
	}

	cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
	if !ok || cartID == "" {
		log.Printf("AddItemCart: CartID tidak ditemukan di konteks. Ini seharusnya sudah disiapkan oleh middleware.")
		redirectBackWithError(w, r, productID, "Gagal mendapatkan sesi keranjang (ID keranjang tidak tersedia).", "error", h.productRepo)
		return
	}

	cart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("AddItemCart: Gagal mengambil cart dengan ID %s: %v", cartID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if cart == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		cart = &models.Cart{
			ID:              cartID,
			BaseTotalPrice:  decimal.Decimal{},
			TaxAmount:       decimal.Decimal{},
			TaxPercent:      decimal.Decimal{},
			DiscountAmount:  decimal.Decimal{},
			DiscountPercent: decimal.Decimal{},
			GrandTotal:      decimal.Decimal{},
		}
		if err := h.cartRepo.CreateCart(r.Context(), cart); err != nil {
			log.Printf("AddItemCart: Gagal membuat cart baru dengan ID %s: %v", cartID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("AddItemCart: Cart baru berhasil dibuat dengan ID: %s", cartID)
	}

	existingItem, err := h.cartItemRepo.GetCartAndProduct(r.Context(), cartID, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("AddItemCart: Gagal mengecek item existing: %v", err)
		redirectBackWithError(w, r, productID, "Gagal memproses permintaan keranjang.", "error", h.productRepo)
		return
	}

	var newTotalQtyInCart int
	if existingItem != nil {
		newTotalQtyInCart = existingItem.Qty + qty
	} else {
		newTotalQtyInCart = qty
	}

	if newTotalQtyInCart > product.Stock {
		log.Printf("AddItemCart: Stok tidak mencukupi untuk product %s. Diminta: %d, Tersedia: %d", product.Name, newTotalQtyInCart, product.Stock)
		redirectBackWithError(w, r, productID, fmt.Sprintf("Stok '%s' tidak mencukupi. Hanya tersedia %d item.", product.Name, product.Stock), "warning", h.productRepo)
		return
	}

	basePrice := product.Price
	baseTotal := basePrice.Mul(decimal.NewFromInt(int64(newTotalQtyInCart)))

	productDiscountAmount := product.DiscountAmount.Mul(decimal.NewFromInt(int64(newTotalQtyInCart)))

	taxPercent := calc.GetTaxPercent()
	taxAmount := calc.CalculateTax(baseTotal.Sub(productDiscountAmount))

	grandTotal := baseTotal.Sub(productDiscountAmount).Add(taxAmount)
	subTotal := baseTotal.Sub(productDiscountAmount)

	if existingItem != nil {
		existingItem.Qty = newTotalQtyInCart
		existingItem.BasePrice = basePrice
		existingItem.BaseTotal = baseTotal
		existingItem.TaxPercent = taxPercent
		existingItem.TaxAmount = taxAmount
		existingItem.SubTotal = subTotal
		existingItem.GrandTotal = grandTotal
		existingItem.UpdatedAt = time.Now()

		if err := h.cartItemRepo.Update(r.Context(), existingItem); err != nil {
			log.Printf("AddItemCart: Gagal update item di cart: %v", err)
			redirectBackWithError(w, r, productID, "Gagal memperbarui item di keranjang.", "error", h.productRepo)
			return
		}
	} else {
		item := &models.CartItem{
			ID:         uuid.New().String(),
			CartID:     cartID,
			ProductID:  productID,
			Qty:        newTotalQtyInCart,
			BasePrice:  basePrice,
			BaseTotal:  baseTotal,
			TaxAmount:  taxAmount,
			TaxPercent: taxPercent,
			GrandTotal: grandTotal,
			SubTotal:   subTotal,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := h.cartItemRepo.Add(r.Context(), item); err != nil {
			log.Printf("Gagal menambahkan item baru: %v", err)
			redirectBackWithError(w, r, productID, "Gagal menambahkan item baru ke keranjang.", "error", h.productRepo)
			return
		}
	}

	if err := h.cartRepo.UpdateCartSummary(r.Context(), cartID); err != nil {
		log.Printf("Gagal update ringkasan cart: %v", err)
	}

	switch action {
	case "buy":
		http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
		return
	default:
		http.Redirect(w, r, fmt.Sprintf("/products/%s?status=success&message=%s", product.Slug, url.QueryEscape("Item berhasil ditambahkan ke keranjang!")), http.StatusSeeOther)
	}
}

func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	qtyStr := r.FormValue("qty")

	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Kuantitas tidak valid!")), http.StatusSeeOther)
		return
	}

	cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
	if !ok || cartID == "" {
		http.Error(w, "Gagal mendapatkan cart session: CartID tidak tersedia di konteks.", http.StatusInternalServerError)
		return
	}

	item, err := h.cartItemRepo.GetCartAndProduct(r.Context(), cartID, productID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Item keranjang tidak ditemukan!")), http.StatusSeeOther)
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil || product == nil {
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Produk terkait tidak ditemukan!")), http.StatusSeeOther)
		return
	}

	if qty > product.Stock {
		log.Printf("UpdateCartItem: Stok tidak mencukupi untuk product %s. Diminta: %d, Tersedia: %d", product.Name, qty, product.Stock)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=warning&message=%s", url.QueryEscape(fmt.Sprintf("Stok '%s' tidak mencukupi. Hanya tersedia %d item.", product.Name, product.Stock))), http.StatusSeeOther)
		return
	}

	item.Qty = qty
	item.BasePrice = product.Price
	item.BaseTotal = item.BasePrice.Mul(decimal.NewFromInt(int64(qty)))

	productDiscountAmount := product.DiscountAmount.Mul(decimal.NewFromInt(int64(qty)))

	item.TaxAmount = calc.CalculateTax(item.BaseTotal.Sub(productDiscountAmount))
	item.GrandTotal = item.BaseTotal.Sub(productDiscountAmount).Add(item.TaxAmount)
	item.SubTotal = item.BaseTotal.Sub(productDiscountAmount)
	item.UpdatedAt = time.Now()

	if err := h.cartItemRepo.Update(r.Context(), item); err != nil {
		log.Println("Gagal update item:", err)
		http.Error(w, "Gagal update item", http.StatusInternalServerError)
		return
	}

	_ = h.cartRepo.UpdateCartSummary(r.Context(), cartID)

	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Kuantitas item keranjang berhasil diperbarui!")), http.StatusSeeOther)
}

func (h *CartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
	if !ok || cartID == "" {
		http.Error(w, "Session tidak ditemukan: CartID tidak tersedia di konteks.", http.StatusInternalServerError)
		return
	}

	productID := r.FormValue("product_id")
	if productID == "" {
		http.Error(w, "Produk tidak valid", http.StatusBadRequest)
		return
	}

	if err := h.cartItemRepo.Delete(r.Context(), cartID, productID); err != nil {
		http.Error(w, "Gagal menghapus item", http.StatusInternalServerError)
		return
	}

	_ = h.cartRepo.UpdateCartSummary(r.Context(), cartID)
	http.Redirect(w, r, fmt.Sprintf("/carts?status=success&message=%s", url.QueryEscape("Item keranjang berhasil dihapus!")), http.StatusSeeOther)
}

func (h *CartHandler) GetCartCount(w http.ResponseWriter, r *http.Request) {
	if cartCountVal := r.Context().Value(helpers.CartCountKey); cartCountVal != nil {
		if count, ok := cartCountVal.(int); ok {
			w.Write([]byte(strconv.Itoa(count)))
			return
		}
	}
	w.Write([]byte("0"))
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
