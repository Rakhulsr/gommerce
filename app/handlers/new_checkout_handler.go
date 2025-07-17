package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
	"gorm.io/gorm"
)

type KomerceCheckoutHandler struct {
	render             *render.Render
	validator          *validator.Validate
	checkoutSvc        *services.CheckoutService
	cartRepo           repositories.CartRepositoryImpl
	userRepo           repositories.UserRepositoryImpl
	orderRepo          repositories.OrderRepository
	productRepo        repositories.ProductRepositoryImpl
	db                 *gorm.DB
	komerceLocationSvc services.KomerceRajaOngkirClient
	addressRepo        repositories.AddressRepository
	sessionStore       sessions.SessionStore
}

func NewKomerceCheckoutHandler(
	render *render.Render,
	validator *validator.Validate,
	checkoutSvc *services.CheckoutService,
	cartRepo repositories.CartRepositoryImpl,
	userRepo repositories.UserRepositoryImpl,
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepositoryImpl,
	db *gorm.DB,
	komerceLocationSvc services.KomerceRajaOngkirClient,
	addressRepo repositories.AddressRepository,
	sessionStore sessions.SessionStore,
) *KomerceCheckoutHandler {
	return &KomerceCheckoutHandler{
		render:             render,
		validator:          validator,
		checkoutSvc:        checkoutSvc,
		cartRepo:           cartRepo,
		userRepo:           userRepo,
		orderRepo:          orderRepo,
		productRepo:        productRepo,
		db:                 db,
		komerceLocationSvc: komerceLocationSvc,
		addressRepo:        addressRepo,
		sessionStore:       sessionStore,
	}
}

type CheckoutPageDataKomerce struct {
	other.BasePageData
	Cart                 *models.Cart
	SelectedAddress      *models.Address
	ShippingCost         decimal.Decimal
	ShippingServiceCode  string
	ShippingServiceName  string
	FinalTotalPrice      decimal.Decimal
	FinalTotalPriceForJS float64
	Errors               map[string]string
}

func (h *KomerceCheckoutHandler) DisplayCheckoutConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(helpers.ContextKeyUserID).(string)

	if err := r.ParseForm(); err != nil {
		log.Printf("DisplayCheckoutConfirmation: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Gagal memproses checkout: Kesalahan form.")), http.StatusSeeOther)
		return
	}

	// --- PERBAIKAN NAMA FIELD ---
	addressID := r.PostFormValue("selected_address_id") // Diperbaiki
	shippingCostStr := r.PostFormValue("shipping_cost")
	shippingServiceCode := r.PostFormValue("shipping_service_code") // Diperbaiki
	shippingServiceName := r.PostFormValue("shipping_service_name") // Diperbaiki
	finalTotalPriceStr := r.PostFormValue("final_total_price")
	// --- AKHIR PERBAIKAN ---

	log.Printf("DisplayCheckoutConfirmation: Received data - AddressID: %s, ShippingCost: %s, ServiceCode: %s, ServiceName: %s, FinalTotalPrice: %s",
		addressID, shippingCostStr, shippingServiceCode, shippingServiceName, finalTotalPriceStr)

	if addressID == "" || shippingCostStr == "" || shippingServiceCode == "" || shippingServiceName == "" || finalTotalPriceStr == "" {
		log.Printf("DisplayCheckoutConfirmation: Data checkout tidak lengkap. AddressID: '%s', ShippingCost: '%s', ServiceCode: '%s', ServiceName: '%s', FinalTotalPrice: '%s'",
			addressID, shippingCostStr, shippingServiceCode, shippingServiceName, finalTotalPriceStr)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Data checkout tidak lengkap. Mohon pilih alamat dan opsi pengiriman.")), http.StatusSeeOther)
		return
	}

	shippingCost, err := decimal.NewFromString(shippingCostStr)
	if err != nil {
		log.Printf("DisplayCheckoutConfirmation: Invalid shipping cost: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Biaya pengiriman tidak valid.")), http.StatusSeeOther)
		return
	}

	finalTotalPrice, err := decimal.NewFromString(finalTotalPriceStr)
	if err != nil {
		log.Printf("DisplayCheckoutConfirmation: Invalid final total price: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Total harga tidak valid.")), http.StatusSeeOther)
		return
	}

	cart, err := h.cartRepo.GetCartWithItems(ctx, helpers.GetCartIDFromContext(r))
	if err != nil || cart == nil || len(cart.CartItems) == 0 {
		log.Printf("DisplayCheckoutConfirmation: Keranjang kosong atau tidak ditemukan untuk user %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Keranjang belanja Anda kosong.")), http.StatusSeeOther)
		return
	}

	selectedAddress, err := h.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil || selectedAddress == nil || selectedAddress.UserID != userID {
		log.Printf("DisplayCheckoutConfirmation: Alamat tidak ditemukan atau tidak valid untuk user %s, addressID %s: %v", userID, addressID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Alamat pengiriman tidak ditemukan atau tidak valid.")), http.StatusSeeOther)
		return
	}

	pageData := CheckoutPageDataKomerce{
		Cart:                 cart,
		SelectedAddress:      selectedAddress,
		ShippingCost:         shippingCost,
		ShippingServiceCode:  shippingServiceCode,
		ShippingServiceName:  shippingServiceName,
		FinalTotalPrice:      finalTotalPrice,
		FinalTotalPriceForJS: finalTotalPrice.InexactFloat64(),
		Errors:               make(map[string]string),
	}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData.BasePageData, baseDataMap)
	pageData.Title = "Konfirmasi Checkout"

	h.render.HTML(w, http.StatusOK, "checkout", pageData)
}

func (h *KomerceCheckoutHandler) InitiateMidtransTransactionPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(helpers.ContextKeyUserID).(string)
	cartID := helpers.GetCartIDFromContext(r)

	if err := r.ParseForm(); err != nil {
		log.Printf("InitiateMidtransTransactionPost: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Gagal memproses pembayaran: Kesalahan form.")), http.StatusSeeOther)
		return
	}

	// --- PERBAIKAN NAMA FIELD ---
	addressID := r.PostFormValue("selected_address_id") // Diperbaiki
	shippingCostStr := r.PostFormValue("shipping_cost")
	shippingServiceCode := r.PostFormValue("shipping_service_code") // Diperbaiki
	shippingServiceName := r.PostFormValue("shipping_service_name") // Diperbaiki
	finalTotalPriceStr := r.PostFormValue("final_total_price")
	// --- AKHIR PERBAIKAN ---

	log.Printf("InitiateMidtransTransactionPost: Received data - AddressID: %s, ShippingCost: %s, ServiceCode: %s, ServiceName: %s, FinalTotalPrice: %s",
		addressID, shippingCostStr, shippingServiceCode, shippingServiceName, finalTotalPriceStr)

	if addressID == "" || shippingCostStr == "" || shippingServiceCode == "" || shippingServiceName == "" || finalTotalPriceStr == "" {
		log.Printf("InitiateMidtransTransactionPost: Data pembayaran tidak lengkap.")
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Data pembayaran tidak lengkap. Mohon lengkapi alamat dan opsi pengiriman.")), http.StatusSeeOther)
		return
	}

	shippingCost, err := decimal.NewFromString(shippingCostStr)
	if err != nil {
		log.Printf("InitiateMidtransTransactionPost: Invalid shipping cost: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Biaya pengiriman tidak valid.")), http.StatusSeeOther)
		return
	}

	cart, err := h.cartRepo.GetCartWithItems(ctx, cartID)
	if err != nil || cart == nil || len(cart.CartItems) == 0 {
		log.Printf("InitiateMidtransTransactionPost: Keranjang kosong atau tidak ditemukan untuk user %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Keranjang belanja Anda kosong atau tidak valid.")), http.StatusSeeOther)
		return
	}

	for _, item := range cart.CartItems {
		product, err := h.productRepo.GetByID(ctx, item.ProductID)
		if err != nil || product == nil {
			log.Printf("InitiateMidtransTransactionPost: Produk %s tidak ditemukan: %v", item.ProductID, err)
			http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Produk '%s' tidak ditemukan.", item.Product.Name))), http.StatusSeeOther)
			return
		}
		if product.Stock < item.Qty {
			log.Printf("InitiateMidtransTransactionPost: Stok tidak mencukupi untuk produk %s. Stok: %d, Qty: %d", product.Name, product.Stock, item.Qty)
			http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Stok produk '%s' tidak mencukupi. Sisa stok: %d", product.Name, product.Stock))), http.StatusSeeOther)
			return
		}
	}

	order, err := h.checkoutSvc.CreateOrder(ctx, userID, cartID, addressID, shippingServiceCode, shippingServiceName, shippingCost)
	if err != nil {
		log.Printf("InitiateMidtransTransactionPost: Gagal membuat order untuk user %s: %v", userID, err)
		if errors.Is(err, services.ErrInsufficientStock) {
			http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Stok produk tidak mencukupi. Mohon periksa kembali keranjang Anda.")), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape(fmt.Sprintf("Gagal membuat pesanan: %v", err))), http.StatusSeeOther)
		return
	}

	orderWithRelations, err := h.orderRepo.GetOrderByIDWithRelations(ctx, order.ID)
	if err != nil {
		log.Printf("InitiateMidtransTransactionPost: Gagal mengambil order dengan relasi untuk Midtrans: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/checkout/error?order_id=%s&message=%s", order.OrderCode, url.QueryEscape("Gagal menyiapkan transaksi pembayaran. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	snapRedirectURL, err := h.checkoutSvc.InitiateMidtransSnapTransaction(ctx, orderWithRelations, &orderWithRelations.User)
	if err != nil {
		log.Printf("InitiateMidtransTransactionPost: Gagal menginisiasi transaksi Midtrans untuk order %s: %v", order.OrderCode, err)
		http.Redirect(w, r, fmt.Sprintf("/checkout/error?order_id=%s&message=%s", order.OrderCode, url.QueryEscape("Gagal menginisiasi pembayaran. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	helpers.ClearCartIDFromSession(w, r, h.sessionStore)
	log.Printf("DisplayCheckoutConfirmation: Received data - AddressID: %s, ShippingCost: %s, ServiceCode: %s, ServiceName: %s, FinalTotalPrice: %s",
		addressID, shippingCostStr, shippingServiceCode, shippingServiceName, finalTotalPriceStr)

	log.Printf("Redirecting to Midtrans Snap URL: %s", snapRedirectURL)
	http.Redirect(w, r, snapRedirectURL, http.StatusSeeOther)
}

func (h *KomerceCheckoutHandler) MidtransNotificationPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var notificationPayload map[string]interface{}
	err := helpers.DecodeJSONBody(w, r, &notificationPayload)
	if err != nil {
		log.Printf("MidtransNotificationPost: Gagal decode JSON body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	orderID, ok := notificationPayload["order_id"].(string)
	if !ok || orderID == "" {
		log.Printf("MidtransNotificationPost: order_id tidak ditemukan atau tidak valid di payload notifikasi: %v", notificationPayload)
		http.Error(w, "Invalid order_id in notification", http.StatusBadRequest)
		return
	}

	transactionStatus, ok := notificationPayload["transaction_status"].(string)
	if !ok {
		log.Printf("MidtransNotificationPost: transaction_status tidak ditemukan atau tidak valid di payload notifikasi untuk order %s", orderID)
		http.Error(w, "Invalid transaction_status in notification", http.StatusBadRequest)
		return
	}

	fraudStatus, ok := notificationPayload["fraud_status"].(string)
	if !ok {
		log.Printf("MidtransNotificationPost: fraud_status tidak ditemukan atau tidak valid di payload notifikasi untuk order %s", orderID)
		http.Error(w, "Invalid fraud_status in notification", http.StatusBadRequest)
		return
	}

	log.Printf("MidtransNotificationPost: Notifikasi diterima untuk OrderID: %s, Status Transaksi: %s, Status Fraud: %s", orderID, transactionStatus, fraudStatus)

	order, err := h.orderRepo.FindByCode(ctx, orderID)
	if err != nil {
		log.Printf("MidtransNotificationPost: Gagal mengambil order dengan OrderCode %s: %v", orderID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if order == nil {
		log.Printf("MidtransNotificationPost: Order dengan OrderCode %s tidak ditemukan.", orderID)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	newPaymentStatus := ""
	newOrderStatus := models.OrderStatusPending

	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			newPaymentStatus = "challenge"
			newOrderStatus = models.OrderStatusPending
		} else if fraudStatus == "accept" {
			newPaymentStatus = "settlement"
			newOrderStatus = models.OrderStatusProcessing
		}
	case "settlement":
		newPaymentStatus = "settlement"
		newOrderStatus = models.OrderStatusProcessing
	case "pending":
		newPaymentStatus = "pending"
		newOrderStatus = models.OrderStatusPending
	case "deny", "expire", "cancel":
		newPaymentStatus = transactionStatus
		newOrderStatus = models.OrderStatusCancelled
	default:
		log.Printf("MidtransNotificationPost: Status transaksi tidak dikenal: %s untuk OrderCode: %s", transactionStatus, orderID)
		http.Error(w, "Unknown transaction status", http.StatusBadRequest)
		return
	}

	if newPaymentStatus != "" {
		err := h.orderRepo.UpdatePaymentStatusAndOrderStatus(ctx, h.db, order.ID, newPaymentStatus, newOrderStatus)
		if err != nil {
			log.Printf("MidtransNotificationPost: Gagal memperbarui status pembayaran dan order untuk OrderID %s: %v", order.ID, err)
			http.Error(w, "Failed to update order status", http.StatusInternalServerError)
			return
		}
		log.Printf("MidtransNotificationPost: Status pembayaran dan order untuk OrderID %s berhasil diperbarui menjadi PaymentStatus: %s, OrderStatus: %s", order.ID, newPaymentStatus, newOrderStatus)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification received and processed"))
}

func (h *KomerceCheckoutHandler) CheckoutFinishGet(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Redirect(w, r, "/carts?status=error&message=Order ID tidak ditemukan.", http.StatusSeeOther)
		return
	}

	ctx := r.Context()
	order, err := h.orderRepo.FindByCode(ctx, orderID)
	if err != nil || order == nil {
		log.Printf("CheckoutFinishGet: Order %s tidak ditemukan: %v", orderID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Pesanan tidak ditemukan.")), http.StatusSeeOther)
		return
	}

	pageData := other.BasePageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData, baseDataMap)
	pageData.Title = "Pembayaran Berhasil"
	pageData.Message = fmt.Sprintf("Pembayaran untuk pesanan Anda #%s berhasil diproses! Terima kasih telah berbelanja.", order.OrderCode)
	pageData.MessageStatus = "success"

	h.render.HTML(w, http.StatusOK, "checkout_finish", pageData)
}

func (h *KomerceCheckoutHandler) CheckoutUnfinishGet(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Redirect(w, r, "/carts?status=error&message=Order ID tidak ditemukan.", http.StatusSeeOther)
		return
	}

	pageData := other.BasePageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData, baseDataMap)
	pageData.Title = "Pembayaran Belum Selesai"
	pageData.Message = fmt.Sprintf("Pembayaran untuk pesanan Anda #%s belum selesai. Silakan coba lagi atau hubungi dukungan.", orderID)
	pageData.MessageStatus = "warning"

	h.render.HTML(w, http.StatusOK, "checkout_unfinish", pageData)
}

func (h *KomerceCheckoutHandler) CheckoutErrorGet(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	errorMessage := r.URL.Query().Get("message")

	pageData := other.BasePageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData, baseDataMap)
	pageData.Title = "Error Pembayaran"
	if orderID != "" {
		pageData.Message = fmt.Sprintf("Terjadi kesalahan saat memproses pesanan #%s. %s", orderID, errorMessage)
	} else {
		pageData.Message = fmt.Sprintf("Terjadi kesalahan saat memproses pembayaran. %s", errorMessage)
	}
	pageData.MessageStatus = "error"

	h.render.HTML(w, http.StatusOK, "checkout_error", pageData)
}

func (h *KomerceCheckoutHandler) CalculateShippingCostKomerce(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		log.Printf("CalculateShippingCostKomerce: Error parsing form: %v", err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Gagal membaca data form.",
		})
		return
	}

	originID := r.FormValue("origin_id")
	destinationID := r.FormValue("destination_id")
	weightStr := r.FormValue("weight")
	courier := r.FormValue("courier")

	if originID == "" || destinationID == "" || weightStr == "" || courier == "" {
		log.Printf("CalculateShippingCostKomerce: Data tidak lengkap. OriginID: %s, DestinationID: %s, Weight: %s, Courier: %s", originID, destinationID, weightStr, courier)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Data pengiriman tidak lengkap. Mohon isi semua field yang wajib.",
		})
		return
	}

	weight, err := strconv.Atoi(weightStr)
	if err != nil || weight <= 0 {
		log.Printf("CalculateShippingCostKomerce: Berat tidak valid: %s, error: %v", weightStr, err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Berat tidak valid, harus berupa angka positif.",
		})
		return
	}

	log.Printf("CalculateShippingCostKomerce: Menghitung biaya pengiriman dari %s ke %s untuk berat %d dengan kurir %s", originID, destinationID, weight, courier)

	originIDInt, _ := strconv.Atoi(originID)
	destinationIdInt, _ := strconv.Atoi(destinationID)

	costs, err := h.komerceLocationSvc.CalculateCost(ctx, originIDInt, destinationIdInt, weight, courier)
	if err != nil {
		log.Printf("CalculateShippingCostKomerce: Gagal menghitung biaya pengiriman dari Komerce API: %v", err)
		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gagal menghitung biaya pengiriman: %v", err),
		})
		return
	}

	if len(costs) == 0 {
		log.Printf("CalculateShippingCostKomerce: Tidak ada biaya pengiriman ditemukan untuk rute ini.")
		h.render.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Tidak ada biaya pengiriman ditemukan untuk rute ini.",
			"data":    []interface{}{},
		})
		return
	}

	log.Printf("CalculateShippingCostKomerce: Berhasil menghitung %d biaya pengiriman.", len(costs))
	h.render.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Biaya pengiriman berhasil ditemukan.",
		"data":    costs,
	})
}
