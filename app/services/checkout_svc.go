package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrInsufficientStock = errors.New("insufficient product stock")

type CheckoutService struct {
	db                *gorm.DB
	cartRepo          repositories.CartRepositoryImpl
	cartItemRepo      repositories.CartItemRepositoryImpl
	productRepo       repositories.ProductRepositoryImpl
	userRepo          repositories.UserRepositoryImpl
	addressRepo       repositories.AddressRepository
	orderRepo         repositories.OrderRepository
	orderItemRepo     repositories.OrderItemRepository
	orderCustomerRepo repositories.OrderCustomerRepository
	paymentRepo       repositories.PaymentRepositoryImpl
}

func NewCheckoutService(
	db *gorm.DB,
	cartRepo repositories.CartRepositoryImpl,
	cartItemRepo repositories.CartItemRepositoryImpl,
	productRepo repositories.ProductRepositoryImpl,
	userRepo repositories.UserRepositoryImpl,
	addressRepo repositories.AddressRepository,
	orderRepo repositories.OrderRepository,
	orderItemRepo repositories.OrderItemRepository,
	orderCustomerRepo repositories.OrderCustomerRepository,
	paymentRepo repositories.PaymentRepositoryImpl,
) *CheckoutService {
	return &CheckoutService{
		db:                db,
		cartRepo:          cartRepo,
		cartItemRepo:      cartItemRepo,
		productRepo:       productRepo,
		userRepo:          userRepo,
		addressRepo:       addressRepo,
		orderRepo:         orderRepo,
		orderItemRepo:     orderItemRepo,
		orderCustomerRepo: orderCustomerRepo,
		paymentRepo:       paymentRepo,
	}
}

func (s *CheckoutService) ProcessFullCheckout(ctx context.Context, userID, cartID, addressID, shippingServiceCode, shippingServiceName string, shippingCost decimal.Decimal) (*models.Order, string, error) {

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, "", fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC: Rolling back transaction due to panic: %v", r)
			tx.Rollback()
		} else if tx.Error != nil {
			log.Printf("ERROR: Rolling back transaction due to unhandled error in tx: %v", tx.Error)
			tx.Rollback()
		}
	}()

	cart, err := s.cartRepo.GetCartWithItems(ctx, cartID)
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to get cart with items: %w", err)
	}
	if cart == nil || len(cart.CartItems) == 0 {
		tx.Rollback()
		return nil, "", errors.New("cart is empty or not found")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		tx.Rollback()
		return nil, "", errors.New("user not found")
	}

	address, err := s.addressRepo.FindAddressByID(ctx, addressID)
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to get address: %w", err)
	}
	if address == nil {
		tx.Rollback()
		return nil, "", errors.New("address not found")
	}

	orderItems := []models.OrderItem{}

	for _, cartItem := range cart.CartItems {
		product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, "", fmt.Errorf("failed to get product %s: %w", cartItem.ProductID, err)
		}
		if product == nil {
			tx.Rollback()
			return nil, "", fmt.Errorf("product %s not found", cartItem.ProductID)
		}

		if product.Stock < cartItem.Qty {
			tx.Rollback()
			return nil, "", fmt.Errorf("%w: product '%s' has insufficient stock. Available: %d, Requested: %d", ErrInsufficientStock, product.Name, product.Stock, cartItem.Qty)
		}

		orderItems = append(orderItems, models.OrderItem{
			ProductID:       product.ID,
			ProductName:     product.Name,
			Qty:             cartItem.Qty,
			Price:           cartItem.Price,
			BaseTotal:       cartItem.Subtotal,
			TaxAmount:       decimal.Zero,
			TaxPercent:      decimal.Zero,
			DiscountAmount:  cartItem.DiscountAmount,
			DiscountPercent: cartItem.DiscountPercent,
			GrandTotal:      cartItem.Subtotal,
		})
	}

	orderCode := fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	order := &models.Order{
		UserID:              userID,
		OrderCode:           orderCode,
		BaseTotalPrice:      cart.BaseTotalPrice,
		DiscountAmount:      cart.DiscountAmount,
		TaxPercent:          cart.TaxPercent,
		TaxAmount:           cart.TaxAmount,
		ShippingCost:        shippingCost,
		GrandTotal:          cart.GrandTotal.Add(shippingCost).Round(2),
		OrderDate:           time.Now(),
		Status:              models.OrderStatusPending,
		PaymentStatus:       "Pending",
		ShippingServiceCode: shippingServiceCode,
		ShippingServiceName: shippingServiceName,
		AddressID:           address.ID,
		ShippingAddress:     address.Address1,
		ShippingService:     shippingServiceName,
	}

	if err := s.orderRepo.Create(ctx, tx, order); err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to create order: %w", err)
	}
	log.Printf("DEBUG: Order created with ID: %s", order.ID)

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	if err := s.orderItemRepo.BulkCreate(ctx, tx, orderItems); err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to create order items: %w", err)
	}
	log.Printf("DEBUG: Order items created for Order ID: %s", order.ID)

	nameParts := strings.Fields(user.FirstName + " " + user.LastName)
	firstName := ""
	lastName := ""
	if len(nameParts) > 0 {
		firstName = nameParts[0]
		if len(nameParts) > 1 {
			lastName = strings.Join(nameParts[1:], " ")
		}
	}
	orderCustomer := &models.OrderCustomer{
		OrderID:      order.ID,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        user.Email,
		Phone:        user.Phone,
		Address1:     address.Address1,
		Address2:     address.Address2,
		LocationID:   address.LocationID,
		LocationName: address.LocationName,
		PostCode:     address.PostCode,
	}
	if err := s.orderCustomerRepo.Create(ctx, tx, orderCustomer); err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to create order customer: %w", err)
	}
	log.Printf("DEBUG: Order customer created for Order ID: %s", order.ID)

	newPayment := &models.Payment{
		OrderID:     order.ID,
		Number:      order.OrderCode,
		Amount:      order.GrandTotal,
		Method:      "Midtrans Snap",
		Status:      "Pending",
		PaymentType: "Snap",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	snapClient := configs.GetMidtransSnapClient()
	var midtransItemDetails []midtrans.ItemDetails

	for _, item := range orderItems {
		itemName := item.ProductName
		if len(itemName) > 50 {
			itemName = itemName[:50]
		}
		priceForMidtrans := item.GrandTotal.Round(0).IntPart()
		midtransItemDetails = append(midtransItemDetails, midtrans.ItemDetails{
			ID:    item.ProductID,
			Name:  itemName,
			Price: int64(priceForMidtrans),
			Qty:   int32(item.Qty),
		})
	}

	shippingItemName := fmt.Sprintf("Biaya Pengiriman (%s - %s)", order.ShippingServiceCode, order.ShippingServiceName)
	if len(shippingItemName) > 50 {
		shippingItemName = shippingItemName[:50]
	}
	shippingCostForMidtrans := order.ShippingCost.Round(0).IntPart()
	midtransItemDetails = append(midtransItemDetails, midtrans.ItemDetails{
		ID:    "SHIPPING_FEE",
		Name:  shippingItemName,
		Price: int64(shippingCostForMidtrans),
		Qty:   1,
	})

	initialItemsTotal := decimal.Zero
	for _, item := range midtransItemDetails {
		initialItemsTotal = initialItemsTotal.Add(decimal.NewFromInt(item.Price).Mul(decimal.NewFromInt32(item.Qty)))
	}
	targetGrossAmount := order.GrandTotal.Round(0)
	difference := targetGrossAmount.Sub(initialItemsTotal)

	if difference.Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
		midtransItemDetails = append(midtransItemDetails, midtrans.ItemDetails{
			ID:    "ADJUSTMENT",
			Name:  "Penyesuaian Total Harga",
			Price: difference.IntPart(),
			Qty:   1,
		})
	}
	grossAmountForMidtrans := order.GrandTotal.Round(0).IntPart()

	custDetails := &midtrans.CustomerDetails{
		FName: user.FirstName,
		LName: user.LastName,
		Email: user.Email,
		Phone: user.Phone,
		BillAddr: &midtrans.CustomerAddress{
			FName:       address.Name,
			Address:     address.Address1,
			City:        address.LocationName,
			Postcode:    address.PostCode,
			Phone:       address.Phone,
			CountryCode: "IDN",
		},
		ShipAddr: &midtrans.CustomerAddress{
			FName:       address.Name,
			Address:     address.Address1,
			City:        address.LocationName,
			Postcode:    address.PostCode,
			Phone:       address.Phone,
			CountryCode: "IDN",
		},
	}

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  order.OrderCode,
			GrossAmt: int64(grossAmountForMidtrans),
		},
		Items:           &midtransItemDetails,
		CustomerDetail:  custDetails,
		EnabledPayments: snap.AllSnapPaymentType,
		Callbacks: &snap.Callbacks{
			Finish: configs.GetAppBaseURL() + "/checkout/finish?order_code=" + order.OrderCode,
		},
	}

	snapResp, errMidtrans := snapClient.CreateTransaction(snapReq)

	if errMidtrans != nil {
		log.Printf("Midtrans CreateTransaction Error: %v", errMidtrans)
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to initiate Midtrans transaction: %w", errMidtrans)
	}

	if snapResp == nil || snapResp.RedirectURL == "" || snapResp.Token == "" {
		log.Printf("Midtrans CreateTransaction returned empty or invalid response for OrderCode: %s. Response: %+v", order.OrderCode, snapResp)
		tx.Rollback()
		return nil, "", errors.New("midtrans transaction initiated but returned invalid response (missing redirect URL or token)")
	}

	log.Printf("DEBUG: Midtrans transaction created successfully. Snap Response: %+v", snapResp)

	newPayment.Token = snapResp.Token

	if err := s.paymentRepo.Create(ctx, tx, newPayment); err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to create payment record for OrderID %s: %v", order.ID, err)
		return nil, "", fmt.Errorf("failed to create payment record: %w", err)
	}
	log.Printf("DEBUG: Payment record created for Payment ID: %s", newPayment.ID)

	log.Printf("DEBUG: Attempting to update Order status for OrderID: %s", order.ID)

	err = s.orderRepo.UpdatePaymentStatusAndOrderStatus(ctx, tx, order.ID, "Pending", models.OrderStatusPending)
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to update order status for OrderID %s: %v", order.ID, err)
		return nil, "", fmt.Errorf("failed to update order status: %w", err)
	}
	log.Printf("DEBUG: Order status updated to Pending for OrderID: %s", order.ID)

	err = tx.Commit().Error
	if err != nil {
		log.Printf("ERROR: Failed to commit database transaction after payment/order status update: %v", err)
		return nil, "", fmt.Errorf("failed to commit database transaction: %w", err)
	}
	log.Printf("DEBUG: Database transaction committed successfully for OrderID: %s", order.ID)

	log.Printf("DEBUG: Successfully processed full checkout for OrderID: %s. Returning result.", order.OrderCode)
	log.Printf("SUCCESS: Order %s created, Payment record created, and Midtrans Snap initiated. Redirect URL: %s", order.OrderCode, snapResp.RedirectURL)
	return order, snapResp.RedirectURL, nil
}
