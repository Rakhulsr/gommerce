package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	DefaultTaxPercent = 12.00
)

type CartService struct {
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	productRepo  repositories.ProductRepositoryImpl
	db           *gorm.DB
}

func NewCartService(
	cartRepo repositories.CartRepositoryImpl,
	cartItemRepo repositories.CartItemRepositoryImpl,
	productRepo repositories.ProductRepositoryImpl,
	db *gorm.DB,
) *CartService {
	return &CartService{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		productRepo:  productRepo,
		db:           db,
	}
}

func (s *CartService) GetUserCart(ctx context.Context, userID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, "", userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan atau membuat keranjang pengguna: %w", err)
	}
	if cart == nil {
		return nil, nil
	}

	detailedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("GetUserCart: Gagal mengambil detailed cart untuk user %s, cart ID %s: %v", userID, cart.ID, err)
		return nil, fmt.Errorf("gagal mendapatkan detailed user cart: %w", err)
	}
	if detailedCart == nil {

		return &models.Cart{
			ID:             cart.ID,
			UserID:         userID,
			BaseTotalPrice: decimal.Zero,
			TaxAmount:      decimal.Zero,
			TaxPercent:     calc.GetTaxPercent(),
			DiscountAmount: decimal.Zero,
			GrandTotal:     decimal.Zero,
			TotalWeight:    decimal.Zero,
			ShippingCost:   decimal.Zero,
			TotalItems:     0,
			CartItems:      []models.CartItem{},
		}, nil
	}

	if err := s.cartRepo.UpdateCartSummary(ctx, detailedCart.ID); err != nil {
		log.Printf("GetUserCart: Gagal update summary cart %s: %v", detailedCart.ID, err)
	}

	finalCart, err := s.cartRepo.GetCartWithItems(ctx, detailedCart.ID)
	if err != nil {
		log.Printf("GetUserCart: Gagal memuat ulang cart setelah update summary %s: %v", detailedCart.ID, err)
		return detailedCart, nil
	}

	return finalCart, nil
}

func (s *CartService) AddItemToCart(ctx context.Context, cartID, userID, productID string, qty int) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return fmt.Errorf("product not found or error getting product: %w", err)
	}

	if product.Stock < qty {
		return fmt.Errorf("not enough stock for product '%s'. Available: %d, Requested: %d", product.Name, product.Stock, qty)
	}

	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user cart: %w", err)
	}

	if cart == nil {

		cart = &models.Cart{
			UserID: userID,

			BaseTotalPrice: decimal.Zero,
			TaxAmount:      decimal.Zero,
			TaxPercent:     calc.GetTaxPercent(),
			DiscountAmount: decimal.Zero,
			GrandTotal:     decimal.Zero,
			TotalWeight:    decimal.Zero,
			ShippingCost:   decimal.Zero,
			TotalItems:     0,
		}
		cart, err = s.cartRepo.AddCart(ctx, cart)
		if err != nil {
			return fmt.Errorf("failed to create new cart: %w", err)
		}
	}

	finalPriceUnit := product.Price
	discountAmountPerUnit := decimal.Zero
	if product.DiscountPercent.GreaterThan(decimal.Zero) {
		discountAmountPerUnit = product.Price.Mul(product.DiscountPercent.Div(decimal.NewFromInt(100)))
		finalPriceUnit = product.Price.Sub(discountAmountPerUnit)
	} else if product.DiscountAmount.GreaterThan(decimal.Zero) {
		discountAmountPerUnit = product.DiscountAmount
		finalPriceUnit = product.Price.Sub(discountAmountPerUnit)
	}

	if finalPriceUnit.LessThan(decimal.Zero) {
		finalPriceUnit = decimal.Zero
	}

	cartItem, err := s.cartItemRepo.GetByCartIDAndProductID(ctx, cart.ID, productID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}

	if cartItem == nil {

		cartItem = &models.CartItem{
			CartID:          cart.ID,
			ProductID:       productID,
			Qty:             qty,
			Price:           product.Price,
			DiscountPercent: product.DiscountPercent,
			DiscountAmount:  discountAmountPerUnit,
			FinalPriceUnit:  finalPriceUnit,
			Subtotal:        finalPriceUnit.Mul(decimal.NewFromInt(int64(qty))),
		}

		if err := s.cartItemRepo.Add(ctx, cartItem); err != nil {
			return fmt.Errorf("failed to create cart item: %w", err)
		}
	} else {

		newQty := cartItem.Qty + qty
		if product.Stock < newQty {
			return fmt.Errorf("not enough stock to add more for product '%s'. Max allowed: %d", product.Name, product.Stock)
		}
		cartItem.Qty = newQty
		cartItem.Price = product.Price
		cartItem.DiscountPercent = product.DiscountPercent
		cartItem.DiscountAmount = discountAmountPerUnit
		cartItem.FinalPriceUnit = finalPriceUnit
		cartItem.Subtotal = finalPriceUnit.Mul(decimal.NewFromInt(int64(newQty)))
		if err := s.cartItemRepo.Update(ctx, cartItem); err != nil {
			return fmt.Errorf("failed to update cart item: %w", err)
		}
	}

	updatedCartWithItems, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("AddItemToCart: Gagal memuat ulang cart setelah menambah/memperbarui item: %v", err)

	} else {
		cart = updatedCartWithItems
	}

	s.CalculateCartTotals(cart)

	if err := s.cartRepo.UpdateCart(ctx, cart); err != nil {
		log.Printf("AddItemToCart: Gagal memperbarui total keranjang setelah menambah item: %v", err)
		return fmt.Errorf("failed to update cart totals after adding item: %w", err)
	}

	return nil
}

func (s *CartService) UpdateCartItemQty(ctx context.Context, userID, productID string, newQty int) (*models.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart: %w", err)
	}
	if cart == nil {
		return nil, fmt.Errorf("cart not found for user: %s", userID)
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return nil, fmt.Errorf("product not found or error getting product: %w", err)
	}

	cartItem, err := s.cartItemRepo.GetByCartIDAndProductID(ctx, cart.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}
	if cartItem == nil {
		return nil, fmt.Errorf("cart item not found in cart %s for product %s", cart.ID, productID)
	}

	if newQty <= 0 {

		if err := s.cartItemRepo.Delete(ctx, cartItem.ID); err != nil {
			return nil, fmt.Errorf("failed to delete cart item: %w", err)
		}
	} else {
		if product.Stock < newQty {
			return nil, fmt.Errorf("not enough stock for product '%s'. Available: %d, Requested: %d", product.Name, product.Stock, newQty)
		}

		finalPriceUnit := product.Price
		discountAmountPerUnit := decimal.Zero
		if product.DiscountPercent.GreaterThan(decimal.Zero) {
			discountAmountPerUnit = product.Price.Mul(product.DiscountPercent.Div(decimal.NewFromInt(100)))
			finalPriceUnit = product.Price.Sub(discountAmountPerUnit)
		} else if product.DiscountAmount.GreaterThan(decimal.Zero) {
			discountAmountPerUnit = product.DiscountAmount
			finalPriceUnit = product.Price.Sub(discountAmountPerUnit)
		}
		if finalPriceUnit.LessThan(decimal.Zero) {
			finalPriceUnit = decimal.Zero
		}

		cartItem.Qty = newQty
		cartItem.Price = product.Price
		cartItem.DiscountPercent = product.DiscountPercent
		cartItem.DiscountAmount = discountAmountPerUnit
		cartItem.FinalPriceUnit = finalPriceUnit
		cartItem.Subtotal = finalPriceUnit.Mul(decimal.NewFromInt(int64(newQty)))
		if err := s.cartItemRepo.Update(ctx, cartItem); err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	}

	s.CalculateCartTotals(cart)
	if err := s.cartRepo.UpdateCart(ctx, cart); err != nil {
		log.Printf("UpdateCartItemQty: Gagal memperbarui total keranjang setelah mengubah item: %v", err)
		return nil, fmt.Errorf("failed to update cart totals: %w", err)
	}

	updatedCart, err := s.cartRepo.GetByUserIDWithItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated cart with items: %w", err)
	}
	s.CalculateCartTotals(updatedCart)

	return updatedCart, nil
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, userID, productID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart: %w", err)
	}
	if cart == nil {
		return nil, nil
	}

	cartItem, err := s.cartItemRepo.GetByCartIDAndProductID(ctx, cart.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}
	if cartItem == nil {
		return nil, nil
	}

	if err := s.cartItemRepo.Delete(ctx, cartItem.ID); err != nil {
		return nil, fmt.Errorf("failed to delete cart item: %w", err)
	}

	s.CalculateCartTotals(cart)
	if err := s.cartRepo.UpdateCart(ctx, cart); err != nil {
		log.Printf("RemoveItemFromCart: Gagal memperbarui total keranjang setelah menghapus item: %v", err)
		return nil, fmt.Errorf("failed to update cart totals after removing item: %w", err)
	}

	updatedCart, err := s.cartRepo.GetByUserIDWithItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated cart with items: %w", err)
	}
	s.CalculateCartTotals(updatedCart)

	return updatedCart, nil
}

func (s *CartService) CalculateCartTotals(cart *models.Cart) {
	if cart == nil {
		return
	}

	baseTotalPrice := decimal.Zero
	totalWeight := decimal.Zero
	totalItems := 0
	totalDiscountAmount := decimal.Zero

	if len(cart.CartItems) == 0 {
		ctx := context.Background()
		items, err := s.cartItemRepo.GetByCartID(ctx, cart.ID)
		if err != nil {
			log.Printf("CalculateCartTotals: Gagal memuat item keranjang untuk kalkulasi: %v", err)

		} else {
			cart.CartItems = items
		}
	}

	for i := range cart.CartItems {
		item := &cart.CartItems[i]

		product, err := s.productRepo.GetByID(context.Background(), item.ProductID)
		if err != nil || product == nil {
			log.Printf("CalculateCartTotals: Produk ID %s tidak ditemukan atau error saat memuat, tidak dapat menghitung item ini.", item.ProductID)
			continue
		}

		item.Price = product.Price
		item.DiscountPercent = product.DiscountPercent
		item.DiscountAmount = product.DiscountAmount

		finalPriceUnit := item.Price
		if item.DiscountPercent.GreaterThan(decimal.Zero) {

			item.DiscountAmount = item.Price.Mul(item.DiscountPercent.Div(decimal.NewFromInt(100)))
			finalPriceUnit = item.Price.Sub(item.DiscountAmount)
		} else if item.DiscountAmount.GreaterThan(decimal.Zero) {

			finalPriceUnit = item.Price.Sub(item.DiscountAmount)
		}

		if finalPriceUnit.LessThan(decimal.Zero) {
			finalPriceUnit = decimal.Zero
		}
		item.FinalPriceUnit = finalPriceUnit

		item.Subtotal = finalPriceUnit.Mul(decimal.NewFromInt(int64(item.Qty)))

		baseTotalPrice = baseTotalPrice.Add(item.Subtotal)
		totalWeight = totalWeight.Add(product.Weight.Mul(decimal.NewFromInt(int64(item.Qty))))
		totalItems += item.Qty
		totalDiscountAmount = totalDiscountAmount.Add(item.DiscountAmount.Mul(decimal.NewFromInt(int64(item.Qty))))
	}

	cart.BaseTotalPrice = baseTotalPrice
	cart.DiscountAmount = totalDiscountAmount
	cart.TotalWeight = totalWeight
	cart.TotalItems = totalItems

	taxPercent := decimal.NewFromFloat(DefaultTaxPercent)
	cart.TaxPercent = taxPercent
	cart.TaxAmount = baseTotalPrice.Mul(taxPercent.Div(decimal.NewFromInt(100)))

	cart.GrandTotal = baseTotalPrice.Add(cart.TaxAmount)

}
