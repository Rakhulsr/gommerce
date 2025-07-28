package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

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
	paymentSvc         services.PaymentService
	cartItemRepo       repositories.CartItemRepositoryImpl
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
	paymentSvc services.PaymentService,
	cartItemRepo repositories.CartItemRepositoryImpl,
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
		paymentSvc:         paymentSvc,
		cartItemRepo:       cartItemRepo,
	}
}

type CheckoutPageDataKomerce struct {
	other.BasePageData
	Cart                        *models.Cart
	SelectedAddress             *models.Address
	ShippingCost                decimal.Decimal
	ShippingServiceCode         string
	ShippingServiceName         string
	FinalTotalPrice             decimal.Decimal
	FinalTotalPriceForJS        float64
	Errors                      map[string]string
	Addresses                   []models.Address
	SelectedAddressID           string
	SelectedShippingServiceCode string
}

func (h *KomerceCheckoutHandler) DisplayCheckoutSelection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	cartID := helpers.GetCartIDFromContext(r)
	cart, err := h.cartRepo.GetCartWithItems(ctx, cartID)
	if err != nil || cart == nil || len(cart.CartItems) == 0 {
		log.Printf("DisplayCheckoutSelection: Keranjang kosong atau tidak ditemukan untuk user %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Keranjang belanja Anda kosong.")), http.StatusSeeOther)
		return
	}

	addresses, err := h.addressRepo.FindAddressesByUserID(ctx, userID)
	if err != nil {
		log.Printf("DisplayCheckoutSelection: Gagal mengambil alamat untuk user %s: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Gagal memuat alamat pengiriman.")), http.StatusSeeOther)
		return
	}

	status := r.URL.Query().Get("status")
	message := r.URL.Query().Get("message")

	selectedAddressID := r.URL.Query().Get("selected_address_id")
	shippingCostStr := r.URL.Query().Get("shipping_cost")
	shippingServiceCode := r.URL.Query().Get("shipping_service_code")
	shippingServiceName := r.URL.Query().Get("shipping_service_name")
	finalTotalPriceStr := r.URL.Query().Get("final_total_price")

	var selectedAddress *models.Address
	if selectedAddressID != "" {
		addr, err := h.addressRepo.FindAddressByID(ctx, selectedAddressID)
		if err == nil && addr.UserID == userID {
			selectedAddress = addr
		}
	}

	var shippingCost decimal.Decimal
	if sc, err := decimal.NewFromString(shippingCostStr); err == nil {
		shippingCost = sc
	}

	var finalTotalPrice decimal.Decimal
	if ft, err := decimal.NewFromString(finalTotalPriceStr); err == nil {
		finalTotalPrice = ft
	}

	pageData := CheckoutPageDataKomerce{
		Cart:                        cart,
		Addresses:                   addresses,
		SelectedAddress:             selectedAddress,
		SelectedAddressID:           selectedAddressID,
		ShippingCost:                shippingCost,
		ShippingServiceCode:         shippingServiceCode,
		ShippingServiceName:         shippingServiceName,
		FinalTotalPrice:             finalTotalPrice,
		FinalTotalPriceForJS:        finalTotalPrice.InexactFloat64(),
		Errors:                      make(map[string]string),
		SelectedShippingServiceCode: shippingServiceCode,
	}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData.BasePageData, baseDataMap)
	pageData.Title = "Checkout"
	pageData.Message = message
	pageData.MessageStatus = status

	h.render.HTML(w, http.StatusOK, "checkot/process", pageData)
}

func (h *KomerceCheckoutHandler) DisplayCheckoutConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(helpers.ContextKeyUserID).(string)

	if err := r.ParseForm(); err != nil {
		log.Printf("DisplayCheckoutConfirmation: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s", url.QueryEscape("Gagal memproses checkout: Kesalahan form.")), http.StatusSeeOther)
		return
	}

	addressID := r.PostFormValue("selected_address_id")
	shippingCostStr := r.PostFormValue("shipping_cost")
	shippingServiceCode := r.PostFormValue("shipping_service_code")
	shippingServiceName := r.PostFormValue("shipping_service_name")
	finalTotalPriceStr := r.PostFormValue("final_total_price")

	if addressID == "" || shippingCostStr == "" || shippingServiceCode == "" || shippingServiceName == "" || finalTotalPriceStr == "" {
		log.Printf("DisplayCheckoutConfirmation: Data checkout tidak lengkap. AddressID: '%s', ShippingCost: '%s', ServiceCode: '%s', ServiceName: '%s', FinalTotalPrice: '%s'",
			addressID, shippingCostStr, shippingServiceCode, shippingServiceName, finalTotalPriceStr)

		http.Redirect(w, r, fmt.Sprintf("/carts?status=error&message=%s&selected_address_id=%s&shipping_cost=%s&shipping_service_code=%s&shipping_service_name=%s&final_total_price=%s",
			url.QueryEscape("Data checkout tidak lengkap. Mohon pilih alamat dan opsi pengiriman."),
			url.QueryEscape(addressID), url.QueryEscape(shippingCostStr), url.QueryEscape(shippingServiceCode), url.QueryEscape(shippingServiceName), url.QueryEscape(finalTotalPriceStr)), http.StatusSeeOther)
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

	h.render.HTML(w, http.StatusOK, "checkout/process", pageData)
}

func (h *KomerceCheckoutHandler) CalculateShippingCostKomerce(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CalculateShippingCostRequest
	if err := helpers.DecodeJSONBody(w, r, &req); err != nil {
		log.Printf("CalculateShippingCostKomerce: Error decoding JSON body: %v", err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request payload.",
		})
		return
	}

	originID := req.Origin
	destinationID := req.Destination
	weight := req.Weight
	courier := req.Courier

	if originID == 0 || destinationID == 0 || weight == 0 || courier == "" {
		log.Printf("CalculateShippingCostKomerce: Data tidak lengkap. OriginID: %d, DestinationID: %d, Weight: %d, Courier: %s", originID, destinationID, weight, courier)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Data pengiriman tidak lengkap. Mohon isi semua field yang wajib.",
		})
		return
	}

	if weight <= 0 {
		log.Printf("CalculateShippingCostKomerce: Berat tidak valid: %d", weight)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Berat tidak valid, harus berupa angka positif.",
		})
		return
	}

	costs, err := h.komerceLocationSvc.CalculateCost(ctx, originID, destinationID, weight, courier)
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

func (h *KomerceCheckoutHandler) InitiateMidtransTransactionPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(helpers.ContextKeyUserID).(string)
	cartID := helpers.GetCartIDFromContext(r)

	var reqBody struct {
		OrderID             string  `json:"order_id"`
		GrossAmount         float64 `json:"gross_amount"`
		AddressID           string  `json:"address_id"`
		ShippingCost        float64 `json:"shipping_cost"`
		ShippingServiceCode string  `json:"shipping_service_code"`
		ShippingServiceName string  `json:"shipping_service_name"`
	}

	if err := helpers.DecodeJSONBody(w, r, &reqBody); err != nil {
		log.Printf("InitiateMidtransTransactionPost: Error decoding JSON body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	addressID := reqBody.AddressID
	shippingCost := decimal.NewFromFloat(reqBody.ShippingCost)
	shippingServiceCode := reqBody.ShippingServiceCode
	shippingServiceName := reqBody.ShippingServiceName

	if addressID == "" || shippingServiceCode == "" || shippingServiceName == "" || shippingCost.IsZero() {
		log.Printf("InitiateMidtransTransactionPost: Data pembayaran tidak lengkap.")
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Data pembayaran tidak lengkap. Mohon lengkapi alamat dan opsi pengiriman.",
		})
		return
	}

	cart, err := h.cartRepo.GetCartWithItems(ctx, cartID)
	if err != nil || cart == nil || len(cart.CartItems) == 0 {
		log.Printf("InitiateMidtransTransactionPost: Keranjang kosong atau tidak ditemukan untuk user %s: %v", userID, err)
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Keranjang belanja Anda kosong atau tidak valid.",
		})
		return
	}

	for _, item := range cart.CartItems {
		product, err := h.productRepo.GetByID(ctx, item.ProductID)
		if err != nil || product == nil {
			log.Printf("InitiateMidtransTransactionPost: Produk %s tidak ditemukan: %v", item.ProductID, err)
			h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Produk '%s' tidak ditemukan.", item.Product.Name),
			})
			return
		}
		if product.Stock < item.Qty {
			log.Printf("InitiateMidtransTransactionPost: Stok tidak mencukupi untuk produk %s. Stok: %d, Qty: %d", product.Name, product.Stock, item.Qty)
			h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Stok produk '%s' tidak mencukupi. Sisa stok: %d", product.Name, product.Stock),
			})
			return
		}
	}

	order, snapRedirectURL, err := h.checkoutSvc.ProcessFullCheckout(
		r.Context(),
		userID,
		cartID,
		addressID,
		shippingServiceCode,
		shippingServiceName,
		shippingCost,
	)

	if err == nil {
		helpers.ClearCartIDFromSession(w, r, h.sessionStore)
		log.Printf("InitiateMidtransTransactionPost: Berhasil menginisiasi Midtrans Snap URL: %s untuk OrderID: %s", snapRedirectURL, order.OrderCode)
		h.render.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  true,
			"token":    snapRedirectURL,
			"order_id": order.OrderCode,
		})
		return
	}

	if errors.Is(err, services.ErrInsufficientStock) {
		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Stok produk tidak mencukupi. Mohon periksa kembali keranjang Anda.",
		})
		return
	}

	h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"success": false,
		"message": fmt.Sprintf("Gagal memproses pesanan: %v", err),
	})
}

func (h *KomerceCheckoutHandler) MidtransNotificationPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var notificationPayload services.MidtransNotificationPayload
	err := json.NewDecoder(r.Body).Decode(&notificationPayload)
	if err != nil {
		log.Printf("MidtransNotificationPost: Gagal decode JSON body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	newPaymentStatus, newOrderStatus, shouldReduceStock, shouldClearCart, shouldRefundStock, order, svcErr := h.paymentSvc.ProcessMidtransNotification(ctx, notificationPayload)
	if svcErr != nil {
		log.Printf("ERROR: PaymentService failed to process Midtrans notification for OrderID %s: %v", notificationPayload.OrderID, svcErr)
		http.Error(w, svcErr.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		log.Printf("WARNING: Order %s not found in database after PaymentService processing.", notificationPayload.OrderID)
		http.Error(w, "Order not found after status processing", http.StatusNotFound)
		return
	}

	txErr := h.db.Transaction(func(tx *gorm.DB) error {

		if shouldReduceStock {

			for _, item := range order.OrderItems {
				product, err := h.productRepo.GetByID(ctx, item.ProductID)
				if err != nil {
					log.Printf("WARNING: Failed to get product %s for stock reduction: %v", item.ProductID, err)

					return fmt.Errorf("product %s not found during stock reduction: %w", item.ProductID, err)
				}
				if product != nil {
					if product.Stock < item.Qty {
						log.Printf("CRITICAL: Insufficient stock for product %s (ID: %s) during reduction. Current: %d, Ordered: %d. Rolling back transaction.", product.Name, product.ID, product.Stock, item.Qty)
						return fmt.Errorf("insufficient stock for product %s. Current: %d, Ordered: %d", product.Name, product.Stock, item.Qty)
					}
					if err := h.productRepo.UpdateStock(ctx, tx, product.ID, product.Stock-item.Qty); err != nil {
						return fmt.Errorf("failed to reduce stock for product %s: %w", product.Name, err)
					}
				}
			}
		} else if shouldRefundStock {

			for _, item := range order.OrderItems {
				product, err := h.productRepo.GetByID(ctx, item.ProductID)
				if err != nil {
					log.Printf("WARNING: Failed to get product %s for stock refund: %v", item.ProductID, err)
					continue
				}
				if product != nil {
					if err := h.productRepo.UpdateStock(ctx, tx, product.ID, product.Stock+item.Qty); err != nil {
						return fmt.Errorf("failed to refund stock for product %s: %w", product.Name, err)
					}
					log.Printf("Stock refunded for product %s (ID: %s). New stock: %d", product.Name, product.ID, product.Stock+item.Qty)
				}
			}
		}

		if shouldClearCart {

			if order.UserID != "" {
				cart, err := h.cartRepo.GetCartByUserID(ctx, order.UserID)
				if err != nil {
					log.Printf("ERROR: Failed to find cart for user %s associated with order %s for clearing: %v", order.UserID, order.ID, err)
					return fmt.Errorf("failed to find cart for clearing: %w", err)
				}
				if cart != nil {

					if err := h.cartItemRepo.DeleteAllItemsByCartID(ctx, tx, cart.ID); err != nil {
						return fmt.Errorf("failed to delete cart items for cart %s: %w", cart.ID, err)
					}

					if err := h.cartRepo.UpdateCartTotalPrice(ctx, tx, cart.ID, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, 0); err != nil {
						return fmt.Errorf("failed to reset cart totals for cart %s: %w", cart.ID, err)
					}

					helpers.ClearCartIDFromSession(w, r, h.sessionStore)
				} else {
					log.Printf("INFO: No active cart found for user %s associated with order %s to clear. Possibly a guest checkout or cart already cleared.", order.UserID, order.OrderCode)
				}
			} else {
				log.Printf("WARNING: UserID not found for Order %s. Cannot clear cart for anonymous user.", order.OrderCode)
			}
		}
		return nil
	})

	if txErr != nil {
		log.Printf("ERROR during Midtrans notification (stock/cart ops) transaction for OrderID %s: %v", order.ID, txErr)

		http.Error(w, "Internal server error during stock/cart processing", http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: Order %s and Payment updated to PaymentStatus: %s, OrderStatus: %d. Stock Reduced: %t, Stock Refunded: %t, Cart Cleared: %t", order.ID, newPaymentStatus, newOrderStatus, shouldReduceStock, shouldRefundStock, shouldClearCart)

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

	h.render.HTML(w, http.StatusOK, "checkout/finish", pageData)
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

	pageData.OrderID = orderID

	h.render.HTML(w, http.StatusOK, "checkout/unfinish", pageData)
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
