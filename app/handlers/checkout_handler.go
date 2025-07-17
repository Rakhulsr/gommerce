package handlers

// import (
// 	"crypto/sha512"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"strconv"
// 	"strings"

// 	"github.com/Rakhulsr/go-ecommerce/app/helpers"
// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/Rakhulsr/go-ecommerce/app/models/other"
// 	"github.com/Rakhulsr/go-ecommerce/app/repositories"
// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
// 	"github.com/shopspring/decimal"
// 	"gorm.io/gorm"

// 	"github.com/go-playground/validator/v10"
// 	"github.com/midtrans/midtrans-go"
// 	"github.com/midtrans/midtrans-go/coreapi"
// 	"github.com/unrolled/render"
// )

// type CheckoutHandler struct {
// 	render          *render.Render
// 	validator       *validator.Validate
// 	checkoutSvc     *services.CheckoutService
// 	cartRepo        repositories.CartRepositoryImpl
// 	userRepo        repositories.UserRepositoryImpl
// 	orderRepo       repositories.OrderRepository
// 	productRepo     repositories.ProductRepositoryImpl
// 	midtransCoreAPI coreapi.Client
// 	db              *gorm.DB
// 	rajaOngkirSvc   services.RajaOngkirClient
// 	addressRepo     repositories.AddressRepository
// }

// func NewCheckoutHandler(
// 	render *render.Render,
// 	validator *validator.Validate,
// 	checkoutSvc *services.CheckoutService,
// 	cartRepo repositories.CartRepositoryImpl,
// 	userRepo repositories.UserRepositoryImpl,
// 	orderRepo repositories.OrderRepository,
// 	productRepo repositories.ProductRepositoryImpl,
// 	db *gorm.DB,
// 	rajaOngkirSvc services.RajaOngkirClient,
// 	addressRepo repositories.AddressRepository,
// ) *CheckoutHandler {
// 	return &CheckoutHandler{
// 		render:      render,
// 		validator:   validator,
// 		checkoutSvc: checkoutSvc,
// 		cartRepo:    cartRepo,
// 		userRepo:    userRepo,
// 		orderRepo:   orderRepo,
// 		productRepo: productRepo,
// 		midtransCoreAPI: coreapi.Client{
// 			ServerKey: midtrans.ServerKey,
// 			Env:       midtrans.Sandbox,
// 		},
// 		db:            db,
// 		rajaOngkirSvc: rajaOngkirSvc,
// 		addressRepo:   addressRepo,
// 	}
// }

// type CheckoutForm struct {
// 	AddressID           string `form:"addressid" validate:"required"`
// 	ShippingServiceCode string `form:"shippingservicecode" validate:"required"`
// 	ShippingServiceName string `form:"shippingservicename" validate:"required"`
// 	ShippingCost        string `form:"shipping_cost" validate:"required,numeric,min=0"`
// 	FinalTotalPrice     string `form:"final_total_price" validate:"required,numeric,min=0"`
// }

// type OrderConfirmationPageData struct {
// 	other.BasePageData
// 	Order           *models.Order
// 	Message         string
// 	IsSuccess       bool
// 	OrderStatusText string // NEW: Field untuk representasi string dari status pesanan
// }

// type CheckoutPageData struct {
// 	other.BasePageData
// 	Cart                *models.Cart
// 	Address             *models.Address
// 	ShippingCost        decimal.Decimal
// 	ShippingServiceCode string
// 	ShippingServiceName string
// 	FinalTotalPrice     decimal.Decimal
// 	Order               *models.Order
// 	MidtransClientKey   string
// }

// // NEW: Fungsi helper untuk mengonversi OrderStatus (int) menjadi string
// func getOrderStatusString(status int) string {
// 	switch status {
// 	case models.OrderStatusPending:
// 		return "Pending"
// 	case models.OrderStatusProcessing:
// 		return "Processing"
// 	case models.OrderStatusShipped:
// 		return "Dikirim"
// 	case models.OrderStatusCompleted:
// 		return "Selesai"
// 	case models.OrderStatusCancelled:
// 		return "Dibatalkan"
// 	case models.OrderStatusFailed:
// 		return "Gagal"
// 	default:
// 		return "Tidak Diketahui"
// 	}
// }

// func (h *CheckoutHandler) DisplayCheckoutConfirmation(w http.ResponseWriter, r *http.Request) {
// 	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !ok || userID == "" {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	cart, err := h.cartRepo.GetOrCreateCartByUserID(r.Context(), "", userID)
// 	if err != nil || cart == nil || len(cart.CartItems) == 0 {
// 		log.Printf("DisplayCheckoutConfirmation: Gagal mengambil keranjang atau keranjang kosong untuk user %s: %v", userID, err)
// 		h.setFlashMessage(w, r, "Keranjang Anda kosong atau sesi kadaluarsa.", "warning")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}

// 	if err := r.ParseForm(); err != nil {
// 		log.Printf("DisplayCheckoutConfirmation: Kesalahan parsing form: %v", err)
// 		h.setFlashMessage(w, r, "Terjadi kesalahan saat memproses permintaan Anda.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}

// 	var form CheckoutForm
// 	form.AddressID = r.PostFormValue("addressid")
// 	form.ShippingServiceCode = r.PostFormValue("shippingservicecode")
// 	form.ShippingServiceName = r.PostFormValue("shippingservicename")
// 	form.ShippingCost = r.PostFormValue("shipping_cost")
// 	form.FinalTotalPrice = r.PostFormValue("final_total_price")

// 	log.Printf("DisplayCheckoutConfirmation: Received form values - AddressID: '%s', ShippingCost: '%s', ServiceCode: '%s', ServiceName: '%s', FinalTotal: '%s'",
// 		form.AddressID, form.ShippingCost, form.ShippingServiceCode, form.ShippingServiceName, form.FinalTotalPrice)

// 	if err := h.validator.Struct(&form); err != nil {
// 		validationErrors := err.(validator.ValidationErrors)
// 		formattedErrors := helpers.FormatValidationErrors(validationErrors)
// 		log.Printf("DisplayCheckoutConfirmation: Validasi form gagal: %+v", formattedErrors)
// 		h.setFlashMessage(w, r, "Mohon lengkapi semua detail pengiriman dan alamat dengan benar.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}

// 	shippingCostFloat, err := strconv.ParseFloat(form.ShippingCost, 64)
// 	if err != nil {
// 		log.Printf("DisplayCheckoutConfirmation: Kesalahan konversi biaya pengiriman '%s': %v", form.ShippingCost, err)
// 		h.setFlashMessage(w, r, "Biaya pengiriman tidak valid.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}
// 	shippingCostDecimal := decimal.NewFromFloat(shippingCostFloat)

// 	finalTotalPriceFloat, err := strconv.ParseFloat(form.FinalTotalPrice, 64)
// 	if err != nil {
// 		log.Printf("DisplayCheckoutConfirmation: Kesalahan konversi total harga akhir '%s': %v", form.FinalTotalPrice, err)
// 		h.setFlashMessage(w, r, "Total harga akhir tidak valid.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}
// 	finalTotalPriceDecimal := decimal.NewFromFloat(finalTotalPriceFloat)

// 	user, err := h.userRepo.FindByID(r.Context(), userID)
// 	if err != nil || user == nil {
// 		log.Printf("DisplayCheckoutConfirmation: Gagal mengambil user %s: %v", userID, err)
// 		h.setFlashMessage(w, r, "User tidak ditemukan.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}

// 	selectedAddress, err := h.addressRepo.FindAddressByID(r.Context(), form.AddressID)
// 	if err != nil {
// 		log.Printf("DisplayCheckoutConfirmation: Gagal mengambil alamat dengan ID %s: %v", form.AddressID, err)
// 		h.setFlashMessage(w, r, "Alamat pengiriman tidak ditemukan.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}
// 	if selectedAddress == nil || selectedAddress.UserID != userID {
// 		log.Printf("DisplayCheckoutConfirmation: Alamat dengan ID %s tidak ditemukan atau bukan milik user %s.", form.AddressID, userID)
// 		h.setFlashMessage(w, r, "Alamat pengiriman tidak valid.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}

// 	if selectedAddress.ProvinceID != "" {
// 		province, err := h.rajaOngkirSvc.GetProvinceByID(selectedAddress.ProvinceID)
// 		if err == nil && province != nil {
// 			selectedAddress.ProvinceName = province.Name
// 		} else {
// 			log.Printf("DisplayCheckoutConfirmation: Failed to get province name for ID %s: %v", selectedAddress.ProvinceID, err)
// 			selectedAddress.ProvinceName = "Provinsi Tidak Dikenal"
// 		}
// 	}
// 	if selectedAddress.CityID != "" {
// 		city, err := h.rajaOngkirSvc.GetCityByID(selectedAddress.CityID)
// 		if err == nil && city != nil {
// 			selectedAddress.CityName = fmt.Sprintf("%s %s", city.Type, city.Name)
// 		} else {
// 			log.Printf("DisplayCheckoutConfirmation: Failed to get city name for ID %s: %v", selectedAddress.CityID, err)
// 			selectedAddress.CityName = "Kota Tidak Dikenal"
// 		}
// 	}

// 	data := h.newTemplateData(r)
// 	data.Title = "Konfirmasi Pesanan"
// 	data.Cart = cart
// 	data.Address = selectedAddress
// 	data.ShippingCost = shippingCostDecimal
// 	data.ShippingServiceCode = form.ShippingServiceCode
// 	data.ShippingServiceName = form.ShippingServiceName
// 	data.FinalTotalPrice = finalTotalPriceDecimal
// 	data.MidtransClientKey = midtrans.ClientKey

// 	order, err := h.checkoutSvc.CreateOrder(r.Context(), userID, cart.ID, form.AddressID, form.ShippingServiceCode, form.ShippingServiceName, shippingCostDecimal)
// 	if err != nil {
// 		log.Printf("DisplayCheckoutConfirmation: Gagal membuat order baru: %v", err)
// 		h.setFlashMessage(w, r, "Gagal membuat pesanan. Mohon coba lagi.", "error")
// 		http.Redirect(w, r, "/carts", http.StatusSeeOther)
// 		return
// 	}
// 	data.Order = order

// 	h.render.HTML(w, http.StatusOK, "checkout/process", data)
// }

// func (h *CheckoutHandler) InitiateMidtransTransactionPost(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		h.render.JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan"})
// 		return
// 	}

// 	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	if !ok || userID == "" {
// 		h.render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
// 		return
// 	}

// 	var requestPayload struct {
// 		OrderID     string  `json:"order_id"`
// 		GrossAmount float64 `json:"gross_amount"`
// 	}

// 	if err := helpers.DecodeJSONBody(w, r, &requestPayload); err != nil {
// 		log.Printf("InitiateMidtransTransactionPost: Gagal decode JSON body: %v", err)
// 		h.render.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
// 		return
// 	}

// 	if requestPayload.OrderID == "" || requestPayload.GrossAmount <= 0 {
// 		h.render.JSON(w, http.StatusBadRequest, map[string]string{"error": "Order ID atau Gross Amount tidak valid"})
// 		return
// 	}

// 	order, err := h.orderRepo.FindByCode(r.Context(), requestPayload.OrderID)
// 	if err != nil || order == nil {
// 		log.Printf("InitiateMidtransTransactionPost: Order dengan kode %s tidak ditemukan: %v", requestPayload.OrderID, err)
// 		// KOREKSI: Kembalikan JSON error
// 		h.render.JSON(w, http.StatusNotFound, map[string]string{"error": "Order not found"})
// 		return
// 	}

// 	if order.UserID != userID {
// 		log.Printf("InitiateMidtransTransactionPost: User %s mencoba mengakses order %s milik user lain %s", userID, order.OrderCode, order.UserID)
// 		h.render.JSON(w, http.StatusForbidden, map[string]string{"error": "Unauthorized access to order"})
// 		return
// 	}

// 	user, err := h.userRepo.FindByID(r.Context(), order.UserID)
// 	if err != nil || user == nil {
// 		log.Printf("InitiateMidtransTransactionPost: Gagal mengambil user %s untuk order %s: %v", order.UserID, order.OrderCode, err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": "User not found for order"})
// 		return
// 	}

// 	midtransSnapURL, err := h.checkoutSvc.InitiateMidtransSnapTransaction(
// 		r.Context(),
// 		order,
// 		user,
// 	)

// 	if err != nil {
// 		log.Printf("InitiateMidtransTransactionPost: Gagal inisiasi Midtrans Snap: %v", err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to initiate Midtrans transaction"})
// 		return
// 	}

// 	h.render.JSON(w, http.StatusOK, map[string]string{"token": midtransSnapURL})
// }

// func (h *CheckoutHandler) CheckoutFinishGet(w http.ResponseWriter, r *http.Request) {
// 	orderCode := r.URL.Query().Get("order_id")
// 	transactionStatus := r.URL.Query().Get("transaction_status")
// 	log.Printf("CheckoutFinishGet: Callback FINISH diterima untuk Order ID: %s, Status: %s", orderCode, transactionStatus)

// 	data := &OrderConfirmationPageData{}
// 	h.populateBaseDataForOrderConfirmation(r, data)
// 	data.Title = "Konfirmasi Pesanan"
// 	data.IsAuthPage = true

// 	order, err := h.orderRepo.FindByCode(r.Context(), orderCode)
// 	if err != nil || order == nil {
// 		log.Printf("CheckoutFinishGet: Pesanan dengan kode %s tidak ditemukan: %v", orderCode, err)
// 		data.Message = "Pesanan tidak ditemukan."
// 		data.IsSuccess = false
// 		h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// 		return
// 	}

// 	data.Order = order
// 	data.IsSuccess = true
// 	data.Message = "Pembayaran berhasil diproses! Pesanan Anda sedang kami siapkan."
// 	data.OrderStatusText = getOrderStatusString(order.Status) // NEW: Set status text
// 	h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// }

// func (h *CheckoutHandler) CheckoutErrorGet(w http.ResponseWriter, r *http.Request) {
// 	orderCode := r.URL.Query().Get("order_id")
// 	transactionStatus := r.URL.Query().Get("transaction_status")
// 	log.Printf("CheckoutErrorGet: Callback ERROR diterima untuk Order ID: %s, Status: %s", orderCode, transactionStatus)

// 	data := &OrderConfirmationPageData{}
// 	h.populateBaseDataForOrderConfirmation(r, data)
// 	data.Title = "Pembayaran Gagal"
// 	data.IsAuthPage = true

// 	order, err := h.orderRepo.FindByCode(r.Context(), orderCode)
// 	if err != nil || order == nil {
// 		log.Printf("CheckoutErrorGet: Pesanan dengan kode %s tidak ditemukan: %v", orderCode, err)
// 		data.Message = "Pesanan tidak ditemukan."
// 		data.IsSuccess = false
// 		h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// 		return
// 	}

// 	data.Order = order
// 	data.IsSuccess = false
// 	data.Message = "Pembayaran Anda gagal. Mohon coba lagi atau hubungi kami."
// 	data.OrderStatusText = getOrderStatusString(order.Status) // NEW: Set status text
// 	h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// }

// func (h *CheckoutHandler) CheckoutUnfinishGet(w http.ResponseWriter, r *http.Request) {
// 	orderCode := r.URL.Query().Get("order_id")
// 	transactionStatus := r.URL.Query().Get("transaction_status")
// 	log.Printf("CheckoutUnfinishGet: Callback UNFINISH diterima untuk Order ID: %s, Status: %s", orderCode, transactionStatus)

// 	data := &OrderConfirmationPageData{}
// 	h.populateBaseDataForOrderConfirmation(r, data)
// 	data.Title = "Pembayaran Belum Selesai"
// 	data.IsAuthPage = true

// 	order, err := h.orderRepo.FindByCode(r.Context(), orderCode)
// 	if err != nil || order == nil {
// 		log.Printf("CheckoutUnfinishGet: Pesanan dengan kode %s tidak ditemukan: %v", orderCode, err)
// 		data.Message = "Pesanan tidak ditemukan."
// 		data.IsSuccess = false
// 		h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// 		return
// 	}

// 	data.Order = order
// 	data.IsSuccess = false
// 	data.Message = "Pembayaran Anda belum selesai. Silakan lanjutkan pembayaran atau coba lagi."
// 	data.OrderStatusText = getOrderStatusString(order.Status) // NEW: Set status text
// 	h.render.HTML(w, http.StatusOK, "order_confirmation", data)
// }

// func (h *CheckoutHandler) MidtransNotificationPost(w http.ResponseWriter, r *http.Request) {
// 	var notificationPayload map[string]interface{}
// 	if err := helpers.DecodeJSONBody(w, r, &notificationPayload); err != nil {
// 		log.Printf("MidtransNotificationPost: Gagal decode JSON body: %v", err)
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	orderID, ok := notificationPayload["order_id"].(string)
// 	if !ok || orderID == "" {
// 		log.Println("MidtransNotificationPost: order_id tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid order_id", http.StatusBadRequest)
// 		return
// 	}

// 	transactionStatus, ok := notificationPayload["transaction_status"].(string)
// 	if !ok {
// 		log.Println("MidtransNotificationPost: transaction_status tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid transaction_status", http.StatusBadRequest)
// 		return
// 	}

// 	fraudStatus, ok := notificationPayload["fraud_status"].(string)
// 	if !ok {
// 		log.Println("MidtransNotificationPost: fraud_status tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid fraud_status", http.StatusBadRequest)
// 		return
// 	}

// 	grossAmountStr, ok := notificationPayload["gross_amount"].(string)
// 	if !ok {
// 		log.Println("MidtransNotificationPost: gross_amount tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid gross_amount", http.StatusBadRequest)
// 		return
// 	}

// 	grossAmountStr = strings.ReplaceAll(grossAmountStr, ".00", "")
// 	grossAmount, err := strconv.ParseFloat(grossAmountStr, 64)
// 	if err != nil {
// 		log.Printf("MidtransNotificationPost: Gagal parse gross_amount: %v", err)
// 		http.Error(w, "Invalid gross_amount format", http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("MidtransNotificationPost: Notifikasi diterima untuk Order ID: %s, Status: %s, Fraud: %s, Gross Amount: %.2f",
// 		orderID, transactionStatus, fraudStatus, grossAmount)

// 	serverKey := h.midtransCoreAPI.ServerKey
// 	if serverKey == "" {
// 		log.Println("MidtransNotificationPost: MIDTRANS_SERVER_KEY tidak ditemukan.")
// 		http.Error(w, "Server key not configured", http.StatusInternalServerError)
// 		return
// 	}

// 	statusCodeStr, ok := notificationPayload["status_code"].(string)
// 	if !ok {
// 		log.Println("MidtransNotificationPost: status_code tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid status_code", http.StatusBadRequest)
// 		return
// 	}

// 	midtransSignatureKey, ok := notificationPayload["signature_key"].(string)
// 	if !ok {
// 		log.Println("MidtransNotificationPost: signature_key tidak ditemukan di payload notifikasi.")
// 		http.Error(w, "Invalid signature_key", http.StatusBadRequest)
// 		return
// 	}

// 	isSignatureValid := isSignatureKeyValid(orderID, statusCodeStr, grossAmountStr, serverKey, midtransSignatureKey)
// 	if !isSignatureValid {
// 		log.Printf("MidtransNotificationPost: Signature key tidak valid untuk Order ID: %s", orderID)
// 		http.Error(w, "Invalid signature key", http.StatusUnauthorized)
// 		return
// 	}
// 	log.Printf("MidtransNotificationPost: Signature key valid untuk Order ID: %s", orderID)

// 	order, err := h.orderRepo.FindByCode(r.Context(), orderID)
// 	if err != nil || order == nil {
// 		log.Printf("MidtransNotificationPost: Pesanan dengan kode %s tidak ditemukan di database: %v", orderID, err)
// 		http.Error(w, "Order not found", http.StatusNotFound)
// 		return
// 	}

// 	var newPaymentStatus string
// 	var newOrderStatus int
// 	shouldDecreaseStock := false

// 	switch transactionStatus {
// 	case "capture":
// 		if fraudStatus == "accept" {
// 			newPaymentStatus = "settlement"
// 			newOrderStatus = models.OrderStatusProcessing
// 			shouldDecreaseStock = true
// 		} else if fraudStatus == "challenge" {
// 			newPaymentStatus = "challenge"
// 			newOrderStatus = models.OrderStatusPending
// 		}
// 	case "settlement":
// 		newPaymentStatus = "settlement"
// 		newOrderStatus = models.OrderStatusProcessing
// 		shouldDecreaseStock = true
// 	case "pending":
// 		newPaymentStatus = "pending"
// 		newOrderStatus = models.OrderStatusPending
// 	case "deny":
// 		newPaymentStatus = "deny"
// 		newOrderStatus = models.OrderStatusFailed
// 	case "expire":
// 		newPaymentStatus = "expire"
// 		newOrderStatus = models.OrderStatusCancelled
// 	case "cancel":
// 		newPaymentStatus = "cancel"
// 		newOrderStatus = models.OrderStatusCancelled
// 	default:
// 		log.Printf("MidtransNotificationPost: Status transaksi tidak dikenal: %s", transactionStatus)
// 		http.Error(w, "Unknown transaction status", http.StatusBadRequest)
// 		return
// 	}

// 	if order.PaymentStatus != newPaymentStatus || order.Status != newOrderStatus {

// 		if err := h.orderRepo.UpdatePaymentStatus(r.Context(), h.db, order.ID, newPaymentStatus); err != nil {
// 			log.Printf("MidtransNotificationPost: Gagal memperbarui payment status order %s ke %s: %v", order.ID, newPaymentStatus, err)
// 			http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
// 			return
// 		}
// 		log.Printf("MidtransNotificationPost: Payment status order %s diperbarui ke %s", order.ID, newPaymentStatus)

// 		if err := h.orderRepo.UpdateStatus(r.Context(), order.ID, newOrderStatus); err != nil {
// 			log.Printf("MidtransNotificationPost: Gagal memperbarui order status order %s ke %d: %v", order.ID, newOrderStatus, err)
// 			http.Error(w, "Failed to update order status", http.StatusInternalServerError)
// 			return
// 		}
// 		log.Printf("MidtransNotificationPost: Order status order %s diperbarui ke %d", order.ID, newOrderStatus)

// 		if shouldDecreaseStock && order.Status != models.OrderStatusProcessing {
// 			for _, item := range order.OrderItems {
// 				product, err := h.productRepo.GetByID(r.Context(), item.ProductID)
// 				if err != nil || product == nil {
// 					log.Printf("MidtransNotificationPost: Produk %s untuk order item %s tidak ditemukan saat pengurangan stok: %v", item.ProductID, item.ID, err)
// 					continue
// 				}
// 				if product.Stock >= item.Qty {
// 					product.Stock -= item.Qty
// 					if err := h.productRepo.UpdateProduct(r.Context(), product); err != nil {
// 						log.Printf("MidtransNotificationPost: Gagal mengurangi stok produk %s: %v", product.ID, err)

// 					} else {
// 						log.Printf("MidtransNotificationPost: Stok produk %s dikurangi sebanyak %d.", product.ID, item.Qty)
// 					}
// 				} else {
// 					log.Printf("MidtransNotificationPost: Stok produk %s tidak cukup untuk order item %s (stok: %d, qty: %d).", product.ID, item.ID, product.Stock, item.Qty)

// 				}
// 			}
// 		}
// 	} else {
// 		log.Printf("MidtransNotificationPost: Status order %s tidak berubah (saat ini %s, target %s).", order.ID, order.PaymentStatus, newPaymentStatus)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprint(w, "Notification processed successfully")
// }

// func (h *CheckoutHandler) populateBaseDataForOrderConfirmation(r *http.Request, pageData *OrderConfirmationPageData) {
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
// 	if hideAdminWelcomeMessage, ok := baseDataMap["HideAdminWelcomeMessage"].(bool); ok {
// 		pageData.HideAdminWelcomeMessage = hideAdminWelcomeMessage
// 	}
// 	pageData.CurrentPath = r.URL.Path
// }

// func isSignatureKeyValid(orderID, statusCode, grossAmount, serverKey, signatureKey string) bool {
// 	toHash := orderID + statusCode + grossAmount + serverKey
// 	hash := sha512.Sum512([]byte(toHash))
// 	expectedSignature := hex.EncodeToString(hash[:])
// 	return expectedSignature == signatureKey
// }

// func (h *CheckoutHandler) newTemplateData(r *http.Request) *CheckoutPageData {
// 	data := &CheckoutPageData{}
// 	baseDataMap := helpers.GetBaseData(r, nil)

// 	if title, ok := baseDataMap["Title"].(string); ok {
// 		data.Title = title
// 	}
// 	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
// 		data.IsLoggedIn = isLoggedIn
// 	}
// 	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok {
// 		data.User = user
// 	}
// 	if userID, ok := baseDataMap["UserID"].(string); ok {
// 		data.UserID = userID
// 	}
// 	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
// 		data.CartCount = cartCount
// 	}
// 	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
// 		data.CSRFToken = csrfToken
// 	}
// 	if message, ok := baseDataMap["Message"].(string); ok {
// 		data.Message = message
// 	}
// 	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
// 		data.MessageStatus = messageStatus
// 	}
// 	if query, ok := baseDataMap["Query"].(url.Values); ok {

// 		data.Query = query
// 	}
// 	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
// 		data.Breadcrumbs = breadcrumbs
// 	}
// 	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
// 		data.IsAuthPage = isAuthPage
// 	}
// 	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
// 		data.IsAdminPage = isAdminPage
// 	}
// 	if hideAdminWelcomeMessage, ok := baseDataMap["HideAdminWelcomeMessage"].(bool); ok {
// 		data.HideAdminWelcomeMessage = hideAdminWelcomeMessage
// 	}
// 	data.CurrentPath = r.URL.Path
// 	return data
// }

// func (h *CheckoutHandler) setFlashMessage(w http.ResponseWriter, r *http.Request, message, status string) {

// 	log.Printf("Flash Message: %s (Status: %s)", message, status)
// }
