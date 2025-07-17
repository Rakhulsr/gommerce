package handlers

// func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
// 	cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
// 	if !ok || cartID == "" {
// 		log.Printf("GetCart: CartID not found in context. Rendering empty cart.")
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
// 		return
// 	}

// 	cart, err := h.cartRepo.GetCartWithItems(r.Context(), cartID)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			log.Printf("GetCart: Cart with ID %s not found, potentially invalid ID in session. Rendering empty cart.", cartID)
// 			h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong atau tidak valid.")
// 			return
// 		}
// 		log.Printf("GetCart: Gagal mengambil data cart untuk ID %s: %v", cartID, err)
// 		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
// 		return
// 	}

// 	if cart == nil || len(cart.CartItems) == 0 {
// 		log.Printf("GetCart: Cart %s found but empty. Rendering empty cart.", cartID)
// 		h.renderEmptyCart(w, r, "info", "Keranjang Anda kosong.")
// 		return
// 	}

// 	totalWeight := 0
// 	grandTotal := decimal.NewFromFloat(0)

// 	for _, cartItem := range cart.CartItems {
// 		if cartItem.Product.ID == "" {
// 			product, err := h.productRepo.GetByID(r.Context(), cartItem.ProductID)
// 			if err != nil || product == nil {
// 				log.Printf("GetCart: Product %s not found for cart item %s. Skipping item recalculation.", cartItem.ProductID, cartItem.ID)
// 				continue
// 			}
// 			cartItem.Product = *product
// 		}

// 		cartItem.BasePrice = cartItem.Product.Price
// 		cartItem.BaseTotal = cartItem.BasePrice.Mul(decimal.NewFromInt(int64(cartItem.Qty)))

// 		itemDiscountAmount := cartItem.Product.DiscountAmount.Mul(decimal.NewFromInt(int64(cartItem.Qty)))

// 		cartItem.TaxPercent = calc.GetTaxPercent()
// 		cartItem.TaxAmount = calc.CalculateTax(cartItem.BaseTotal.Sub(itemDiscountAmount))

// 		cartItem.SubTotal = cartItem.BaseTotal.Sub(itemDiscountAmount)
// 		cartItem.GrandTotal = cartItem.SubTotal.Add(cartItem.TaxAmount)

// 		if cartItem.Product.ID != "" {
// 			productWeigth := cartItem.Product.Weight.InexactFloat64()
// 			ceilWeight := math.Ceil(productWeigth)
// 			itemWeight := cartItem.Qty * int(ceilWeight)
// 			totalWeight += itemWeight

// 			grandTotal = grandTotal.Add(cartItem.GrandTotal)
// 		}
// 	}

// 	cart.TotalWeight = totalWeight
// 	cart.GrandTotal = grandTotal

// 	if err := h.cartRepo.UpdateCartSummary(r.Context(), cart.ID); err != nil {
// 		log.Printf("GetCart: Gagal update ringkasan cart %s setelah recalculate item: %v", cart.ID, err)
// 	}

// 	status := r.URL.Query().Get("status")
// 	message := r.URL.Query().Get("message")

// 	provinces, err := h.locationSvc.GetProvincesFromAPI() // Gunakan h.locationSvc
// 	if err != nil {
// 		log.Printf("GetCart: Gagal mengambil daftar provinsi dari RajaOngkir API: %v", err)
// 		status = "error"
// 		message = "Gagal memuat daftar provinsi untuk pengiriman. Coba lagi nanti."
// 		provinces = []other.Province{}
// 	}

// 	// NEW: Ambil userID dari konteks untuk memuat alamat
// 	userID, userOk := r.Context().Value(helpers.ContextKeyUserID).(string)
// 	var userAddresses []models.Address
// 	if userOk && userID != "" {
// 		userWithAddresses, err := h.userRepo.GetUserByIDWithAddresses(r.Context(), userID)
// 		if err != nil {
// 			log.Printf("GetCart: Gagal mengambil user dengan alamat untuk user %s: %v", userID, err)
// 			// Lanjutkan saja, alamat akan kosong
// 		} else if userWithAddresses != nil {
// 			userAddresses = userWithAddresses.Address
// 		}
// 	} else {
// 		log.Printf("GetCart: UserID tidak ditemukan di konteks untuk memuat alamat.")
// 	}

// 	supportedCouriers := []other.Courier{
// 		{Code: "jne", Name: "JNE"},
// 		{Code: "tiki", Name: "TIKI"},
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
// 		"GrandTotalAmountForJS": cart.GrandTotal.InexactFloat64(), // Pastikan ini float64 untuk JS
// 		"Addresses":             userAddresses,                    // NEW: Tambahkan data alamat pengguna
// 	}

// 	datas := helpers.GetBaseData(r, pageSpecificData)
// 	_ = h.render.HTML(w, http.StatusOK, "carts", datas)
// }
