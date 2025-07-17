package services

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/Rakhulsr/go-ecommerce/app/configs"
// 	"github.com/Rakhulsr/go-ecommerce/app/helpers"

// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/Rakhulsr/go-ecommerce/app/repositories"
// 	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
// 	"github.com/google/uuid"
// 	"github.com/midtrans/midtrans-go"
// 	_ "github.com/midtrans/midtrans-go/coreapi"
// 	_ "github.com/midtrans/midtrans-go/iris"
// 	"github.com/midtrans/midtrans-go/snap"
// 	"github.com/shopspring/decimal"
// 	"gorm.io/gorm"
// )

// var ErrInsufficientStock = errors.New("stok produk tidak mencukupi")
// var ErrOrderNotFound = errors.New("order tidak ditemukan")

// type CheckoutService struct {
// 	db                 *gorm.DB
// 	cartRepo           repositories.CartRepositoryImpl
// 	cartItemRepo       repositories.CartItemRepositoryImpl
// 	productRepo        repositories.ProductRepositoryImpl
// 	userRepo           repositories.UserRepositoryImpl
// 	addressRepo        repositories.AddressRepository
// 	orderRepo          repositories.OrderRepository
// 	orderItemRepo      repositories.OrderItemRepository
// 	orderCustomerRepo  repositories.OrderCustomerRepository
// 	midtransSnapClient snap.Client
// }

// func NewCheckoutService(
// 	db *gorm.DB,
// 	cartRepo repositories.CartRepositoryImpl,
// 	cartItemRepo repositories.CartItemRepositoryImpl,
// 	productRepo repositories.ProductRepositoryImpl,
// 	userRepo repositories.UserRepositoryImpl,
// 	addressRepo repositories.AddressRepository,
// 	orderRepo repositories.OrderRepository,
// 	orderItemRepo repositories.OrderItemRepository,
// 	orderCustomerRepo repositories.OrderCustomerRepository,
// ) *CheckoutService {
// 	return &CheckoutService{
// 		db:                 db,
// 		cartRepo:           cartRepo,
// 		cartItemRepo:       cartItemRepo,
// 		productRepo:        productRepo,
// 		userRepo:           userRepo,
// 		addressRepo:        addressRepo,
// 		orderRepo:          orderRepo,
// 		orderItemRepo:      orderItemRepo,
// 		orderCustomerRepo:  orderCustomerRepo,
// 		midtransSnapClient: configs.MidtransClient,
// 	}
// }

// func (s *CheckoutService) CreateOrder(ctx context.Context, userID, cartID, addressID, shippingServiceCode, shippingServiceName string, shippingCost decimal.Decimal) (*models.Order, error) {
// 	tx := s.db.WithContext(ctx).Begin()
// 	if tx.Error != nil {
// 		return nil, fmt.Errorf("gagal memulai transaksi database: %w", tx.Error)
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			panic(r)
// 		}
// 	}()

// 	cart, err := s.cartRepo.GetCartWithItems(ctx, cartID)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal mengambil keranjang: %w", err)
// 	}
// 	if cart == nil || len(cart.CartItems) == 0 {
// 		tx.Rollback()
// 		return nil, errors.New("keranjang kosong atau tidak ditemukan")
// 	}

// 	user, err := s.userRepo.FindByID(ctx, userID)
// 	if err != nil || user == nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("pengguna tidak ditemukan: %w", err)
// 	}

// 	address, err := s.addressRepo.FindAddressByID(ctx, addressID)
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal mengambil alamat pengiriman: %w", err)
// 	}
// 	if address == nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("alamat pengiriman dengan ID '%s' tidak ditemukan", addressID)
// 	}

// 	var totalOrderItemsGrandTotal decimal.Decimal = decimal.Zero
// 	var orderItems []models.OrderItem
// 	for _, cartItem := range cart.CartItems {
// 		product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
// 		if err != nil || product == nil {
// 			tx.Rollback()
// 			return nil, fmt.Errorf("produk '%s' tidak ditemukan", cartItem.ProductID)
// 		}
// 		if product.Stock < cartItem.Qty {
// 			tx.Rollback()
// 			return nil, ErrInsufficientStock
// 		}
// 		if err := s.productRepo.DecrementStock(ctx, tx, product.ID, cartItem.Qty); err != nil {
// 			tx.Rollback()
// 			return nil, fmt.Errorf("gagal mengurangi stok produk %s: %w", product.Name, err)
// 		}

// 		itemBasePrice := product.Price
// 		itemBaseTotal := itemBasePrice.Mul(decimal.NewFromInt(int64(cartItem.Qty)))
// 		itemDiscountAmount := product.DiscountAmount.Mul(decimal.NewFromInt(int64(cartItem.Qty)))
// 		itemTaxPercent := calc.GetTaxPercent()
// 		itemTaxAmount := calc.CalculateTax(itemBaseTotal.Sub(itemDiscountAmount))
// 		itemSubTotal := itemBaseTotal.Sub(itemDiscountAmount)
// 		itemGrandTotal := itemSubTotal.Add(itemTaxAmount).Round(0)

// 		orderItem := models.OrderItem{
// 			ID:              uuid.New().String(),
// 			OrderID:         "",
// 			ProductID:       product.ID,
// 			ProductName:     product.Name,
// 			ProductSku:      product.Sku,
// 			Qty:             cartItem.Qty,
// 			BasePrice:       itemBasePrice,
// 			BaseTotal:       itemBaseTotal,
// 			TaxAmount:       itemTaxAmount,
// 			TaxPercent:      itemTaxPercent,
// 			DiscountAmount:  itemDiscountAmount,
// 			DiscountPercent: product.DiscountPercent,
// 			GrandTotal:      itemGrandTotal,
// 			CreatedAt:       time.Now(),
// 			UpdatedAt:       time.Now(),
// 		}
// 		orderItems = append(orderItems, orderItem)
// 		totalOrderItemsGrandTotal = totalOrderItemsGrandTotal.Add(itemGrandTotal)
// 	}

// 	orderCustomer := &models.OrderCustomer{
// 		ID:         uuid.New().String(),
// 		FirstName:  user.FirstName,
// 		LastName:   user.LastName,
// 		Email:      user.Email,
// 		Phone:      address.Phone,
// 		Address1:   address.Address1,
// 		Address2:   address.Address2,
// 		CityID:     address.CityID,
// 		ProvinceID: address.ProvinceID,
// 		PostCode:   address.PostCode,
// 		CreatedAt:  time.Now(),
// 		UpdatedAt:  time.Now(),
// 	}
// 	if err := s.orderCustomerRepo.Create(ctx, tx, orderCustomer); err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal membuat order customer: %w", err)
// 	}

// 	orderCode := helpers.GenerateOrderCode()
// 	roundedShippingCost := shippingCost.Round(0)
// 	calculatedGrandTotal := totalOrderItemsGrandTotal.Add(roundedShippingCost)

// 	order := &models.Order{
// 		ID:                  uuid.New().String(),
// 		UserID:              userID,
// 		OrderCode:           orderCode,
// 		OrderDate:           time.Now(),
// 		OrderCustomerID:     orderCustomer.ID,
// 		BaseTotalPrice:      cart.BaseTotalPrice,
// 		TaxAmount:           cart.TaxAmount,
// 		TaxPercent:          cart.TaxPercent,
// 		DiscountAmount:      cart.DiscountAmount,
// 		DiscountPercent:     cart.DiscountPercent,
// 		ShippingCost:        roundedShippingCost,
// 		GrandTotal:          calculatedGrandTotal,
// 		ShippingAddress:     fmt.Sprintf("%s, %s, %s, %s, %s", address.Address1, address.Address2, address.CityID, address.ProvinceID, address.PostCode),
// 		ShippingService:     fmt.Sprintf("%s - %s", shippingServiceCode, shippingServiceName),
// 		ShippingServiceCode: shippingServiceCode,
// 		ShippingServiceName: shippingServiceName,
// 		PaymentStatus:       "pending",
// 		Status:              models.OrderStatusPending,
// 		AddressID:           addressID,
// 		CreatedAt:           time.Now(),
// 		UpdatedAt:           time.Now(),
// 	}

// 	if err := s.orderRepo.Create(ctx, tx, order); err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal membuat order: %w", err)
// 	}

// 	for i := range orderItems {
// 		orderItems[i].OrderID = order.ID
// 	}
// 	if err := s.orderItemRepo.BulkCreate(ctx, tx, orderItems); err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal membuat order items: %w", err)
// 	}

// 	if err := s.cartItemRepo.ClearCartItems(ctx, tx, cartID); err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal menghapus item keranjang setelah checkout: %w", err)
// 	}
// 	if err := s.cartRepo.DeleteCart(ctx, tx, cartID); err != nil {
// 		tx.Rollback()
// 		return nil, fmt.Errorf("gagal menghapus keranjang setelah checkout: %w", err)
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		return nil, fmt.Errorf("gagal melakukan commit transaksi database: %w", err)
// 	}

// 	log.Printf("Order berhasil dibuat dengan kode: %s", order.OrderCode)
// 	return order, nil
// }

// func (s *CheckoutService) InitiateMidtransSnapTransaction(ctx context.Context, order *models.Order, user *models.User) (string, error) {
// 	if order == nil || user == nil {
// 		return "", errors.New("data order atau user tidak boleh kosong")
// 	}
// 	if order.OrderCustomer.ID == "" {
// 		log.Printf("InitiateMidtransSnapTransaction: OrderCustomer tidak terisi untuk order %s. Pastikan OrderCustomer di-preload.", order.OrderCode)
// 		return "", errors.New("data order customer tidak lengkap (pastikan OrderCustomer di-preload)")
// 	}
// 	if len(order.OrderItems) == 0 {
// 		log.Printf("InitiateMidtransSnapTransaction: OrderItems kosong untuk order %s. Pastikan OrderItems di-preload.", order.OrderCode)
// 		return "", errors.New("data order items tidak lengkap (pastikan OrderItems di-preload)")
// 	}

// 	var midtransItems []midtrans.ItemDetails
// 	for _, item := range order.OrderItems {
// 		unitPrice := item.GrandTotal.Div(decimal.NewFromInt(int64(item.Qty))).IntPart()

// 		midtransItems = append(midtransItems, midtrans.ItemDetails{
// 			ID:    item.ProductID,
// 			Name:  item.ProductName,
// 			Price: unitPrice,
// 			Qty:   int32(item.Qty),
// 		})
// 	}

// 	if order.ShippingCost.GreaterThan(decimal.Zero) {
// 		midtransItems = append(midtransItems, midtrans.ItemDetails{
// 			ID:    "SHIPPING_FEE",
// 			Name:  "Biaya Pengiriman (" + order.ShippingService + ")",
// 			Price: order.ShippingCost.IntPart(),
// 			Qty:   1,
// 		})
// 	}

// 	// --- Debug Logging Tambahan ---
// 	log.Printf("Midtrans Request Debug - OrderCode: %s", order.OrderCode)
// 	log.Printf("Midtrans Request Debug - GrossAmt (dari Order.GrandTotal decimal): %s, GrossAmt (sent to Midtrans - int part): %d", order.GrandTotal.String(), order.GrandTotal.IntPart())

// 	var debugSumOfItemPrices int64 = 0
// 	for i, item := range midtransItems {
// 		log.Printf("Midtrans Request Debug - Item %d: Name=%s, ID=%s, Price=%d, Qty=%d", i, item.Name, item.ID, item.Price, item.Qty)
// 		debugSumOfItemPrices += item.Price * int64(item.Qty)
// 	}
// 	log.Printf("Midtrans Request Debug - Sum of ALL ItemDetails.Price (calculated by backend): %d", debugSumOfItemPrices)

// 	if order.GrandTotal.IntPart() != debugSumOfItemPrices {
// 		log.Printf("Midtrans Request Debug - !!! WARNING: GrossAmt (%d) TIDAK SAMA dengan Sum of ItemDetails.Price (%d). Ini kemungkinan besar akan menyebabkan error Midtrans.", order.GrandTotal.IntPart(), debugSumOfItemPrices)
// 	}
// 	// --- End Debug Logging ---

// 	snapReq := &snap.Request{
// 		TransactionDetails: midtrans.TransactionDetails{
// 			OrderID:  order.OrderCode,
// 			GrossAmt: order.GrandTotal.IntPart(),
// 		},
// 		CreditCard: &snap.CreditCardDetails{
// 			Secure: true,
// 		},
// 		CustomerDetail: &midtrans.CustomerDetails{
// 			FName: user.FirstName,
// 			LName: user.LastName,
// 			Email: user.Email,
// 			Phone: order.OrderCustomer.Phone,
// 			BillAddr: &midtrans.CustomerAddress{
// 				FName:       user.FirstName,
// 				LName:       user.LastName,
// 				Address:     order.OrderCustomer.Address1,
// 				City:        order.OrderCustomer.CityID,
// 				Postcode:    order.OrderCustomer.PostCode,
// 				Phone:       order.OrderCustomer.Phone,
// 				CountryCode: "IDN",
// 			},
// 			ShipAddr: &midtrans.CustomerAddress{
// 				FName:       user.FirstName,
// 				LName:       user.LastName,
// 				Address:     order.OrderCustomer.Address1,
// 				City:        order.OrderCustomer.CityID,
// 				Postcode:    order.OrderCustomer.PostCode,
// 				Phone:       order.OrderCustomer.Phone,
// 				CountryCode: "IDN",
// 			},
// 		},
// 		Items: &midtransItems,
// 		Callbacks: &snap.Callbacks{
// 			Finish: fmt.Sprintf("%s/checkout/finish?order_id=%s", configs.LoadENV.APP_URL, order.OrderCode),
// 		},
// 		EnabledPayments: snap.AllSnapPaymentType,
// 		CustomField1:    order.ID,
// 		CustomField2:    user.ID,
// 	}

// 	snapResp, err := s.midtransSnapClient.CreateTransaction(snapReq)
// 	if err != nil {
// 		log.Printf("InitiateMidtransSnapTransaction: Gagal memanggil Midtrans CreateTransaction untuk OrderCode %s: %v", order.OrderCode, err)
// 		return "", fmt.Errorf("gagal menginisiasi transaksi Midtrans: %w", err)
// 	}

// 	// KOREKSI: Tingkatkan delay menjadi 2 detik
// 	time.Sleep(5 * time.Second) // Delay 2 detik

// 	if err := s.orderRepo.UpdateMidtransDetails(ctx, s.db, order.ID, snapResp.Token, snapResp.RedirectURL); err != nil {
// 		log.Printf("InitiateMidtransSnapTransaction: Gagal memperbarui detail Midtrans di order %s: %v", order.ID, err)
// 	}

// 	log.Printf("Midtrans Snap Transaction berhasil diinisiasi untuk OrderCode: %s, Snap URL: %s", order.OrderCode, snapResp.RedirectURL)
// 	return snapResp.RedirectURL, nil
// }
