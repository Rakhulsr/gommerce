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

	filteredCartItems := []models.CartItem{}
	shouldUpdateCart := false
	for _, item := range detailedCart.CartItems {
		if item.Product != nil && item.Product.ID != "" {

			filteredCartItems = append(filteredCartItems, item)
		} else {
			err := s.cartItemRepo.Delete(ctx, item.ID)
			if err != nil {
				log.Printf("GetUserCart: Gagal menghapus CartItem '%s' (produk hilang): %v", item.ID, err)
			}
			shouldUpdateCart = true
		}
	}
	detailedCart.CartItems = filteredCartItems

	s.CalculateCartTotals(detailedCart)

	if shouldUpdateCart || !detailedCart.GrandTotal.Equal(cart.GrandTotal) || detailedCart.TotalItems != cart.TotalItems {

		if err := s.cartRepo.UpdateCart(ctx, detailedCart); err != nil {
			log.Printf("GetUserCart: Gagal memperbarui cart %s di DB setelah kalkulasi ulang: %v", detailedCart.ID, err)
			return detailedCart, fmt.Errorf("gagal memperbarui keranjang setelah kalkulasi ulang: %w", err)
		}
	}

	return detailedCart, nil
}

func (s *CartService) AddItemToCart(ctx context.Context, cartID, userID, productID string, qty int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
				UserID:         userID,
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
			return fmt.Errorf("failed to reload cart for total calculation: %w", err)
		}
		if updatedCartWithItems == nil {
			return fmt.Errorf("reloaded cart is nil after item addition/update")
		}

		s.CalculateCartTotals(updatedCartWithItems)

		if err := s.cartRepo.UpdateCart(ctx, updatedCartWithItems); err != nil {
			log.Printf("AddItemToCart: Gagal memperbarui total keranjang setelah menambah item: %v", err)
			return fmt.Errorf("failed to update cart totals after adding item: %w", err)
		}
		return nil
	})
}

func (s *CartService) UpdateCartItemQty(ctx context.Context, userID, productID string, newQty int) (*models.Cart, error) {
	var finalCart *models.Cart
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user cart: %w", err)
		}
		if cart == nil {
			return fmt.Errorf("cart not found for user: %s", userID)
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return fmt.Errorf("product not found or error getting product: %w", err)
		}

		cartItem, err := s.cartItemRepo.GetByCartIDAndProductID(ctx, cart.ID, productID)
		if err != nil {
			return fmt.Errorf("failed to get cart item: %w", err)
		}
		if cartItem == nil {
			return fmt.Errorf("cart item not found in cart %s for product %s", cart.ID, productID)
		}

		if newQty <= 0 {
			if err := s.cartItemRepo.Delete(ctx, cartItem.ID); err != nil {
				return fmt.Errorf("failed to delete cart item: %w", err)
			}
		} else {
			if product.Stock < newQty {
				return fmt.Errorf("not enough stock for product '%s'. Available: %d, Requested: %d", product.Name, product.Stock, newQty)
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
				return fmt.Errorf("failed to update cart item: %w", err)
			}
		}

		updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
		if err != nil {
			log.Printf("UpdateCartItemQty: Gagal memuat ulang cart setelah mengubah item: %v", err)
			return fmt.Errorf("failed to reload cart for total calculation: %w", err)
		}
		if updatedCart == nil {

			updatedCart = &models.Cart{
				ID: cart.ID, UserID: userID, CartItems: []models.CartItem{},
				BaseTotalPrice: decimal.Zero, TaxAmount: decimal.Zero, TaxPercent: calc.GetTaxPercent(),
				DiscountAmount: decimal.Zero, GrandTotal: decimal.Zero, TotalWeight: decimal.Zero,
				ShippingCost: decimal.Zero, TotalItems: 0,
			}
		}

		s.CalculateCartTotals(updatedCart)
		if err := s.cartRepo.UpdateCart(ctx, updatedCart); err != nil {
			log.Printf("UpdateCartItemQty: Gagal memperbarui total keranjang setelah mengubah item: %v", err)
			return fmt.Errorf("failed to update cart totals: %w", err)
		}
		finalCart = updatedCart
		return nil
	})
	return finalCart, err
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, userID, productID string) (*models.Cart, error) {
	var finalCart *models.Cart
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user cart: %w", err)
		}
		if cart == nil {
			return nil
		}

		cartItem, err := s.cartItemRepo.GetByCartIDAndProductID(ctx, cart.ID, productID)
		if err != nil {
			return fmt.Errorf("failed to get cart item: %w", err)
		}
		if cartItem == nil {
			return nil
		}

		if err := s.cartItemRepo.Delete(ctx, cartItem.ID); err != nil {
			return fmt.Errorf("failed to delete cart item: %w", err)
		}

		updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
		if err != nil {
			log.Printf("RemoveItemFromCart: Gagal memuat ulang cart setelah menghapus item: %v", err)
			return fmt.Errorf("failed to reload cart for total calculation: %w", err)
		}
		if updatedCart == nil {

			updatedCart = &models.Cart{
				ID: cart.ID, UserID: userID, CartItems: []models.CartItem{},
				BaseTotalPrice: decimal.Zero, TaxAmount: decimal.Zero, TaxPercent: calc.GetTaxPercent(),
				DiscountAmount: decimal.Zero, GrandTotal: decimal.Zero, TotalWeight: decimal.Zero,
				ShippingCost: decimal.Zero, TotalItems: 0,
			}
		}

		s.CalculateCartTotals(updatedCart)
		if err := s.cartRepo.UpdateCart(ctx, updatedCart); err != nil {
			log.Printf("RemoveItemFromCart: Gagal memperbarui total keranjang setelah menghapus item: %v", err)
			return fmt.Errorf("failed to update cart totals after removing item: %w", err)
		}
		finalCart = updatedCart
		return nil
	})
	return finalCart, err
}

func (s *CartService) ClearUserCart(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan keranjang user untuk dibersihkan: %w", err)
		}
		if cart == nil {
			log.Printf("ClearUserCart: Keranjang user %s tidak ditemukan, tidak ada yang perlu dibersihkan.", userID)
			return nil
		}

		if err := s.cartItemRepo.DeleteAllItemsByCartID(ctx, tx, cart.ID); err != nil {
			return fmt.Errorf("gagal membersihkan item keranjang: %w", err)
		}

		err = s.cartRepo.UpdateCartTotalPrice(ctx, tx, cart.ID, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, 0)
		if err != nil {
			return fmt.Errorf("gagal mereset total harga keranjang: %w", err)
		}
		return nil
	})
}

func (s *CartService) CalculateCartTotals(cart *models.Cart) {
	if cart == nil {
		return
	}

	baseTotalPrice := decimal.Zero
	totalWeight := decimal.Zero
	totalItems := 0
	totalDiscountAmount := decimal.Zero

	for _, item := range cart.CartItems {

		productPrice := item.Product.Price
		productWeight := item.Product.Weight

		finalPriceUnit := productPrice
		discountAmountPerUnit := decimal.Zero
		if item.Product.DiscountPercent.GreaterThan(decimal.Zero) {
			discountAmountPerUnit = productPrice.Mul(item.Product.DiscountPercent.Div(decimal.NewFromInt(100)))
			finalPriceUnit = productPrice.Sub(discountAmountPerUnit)
		} else if item.Product.DiscountAmount.GreaterThan(decimal.Zero) {
			discountAmountPerUnit = item.Product.DiscountAmount
			finalPriceUnit = productPrice.Sub(discountAmountPerUnit)
		}

		if finalPriceUnit.LessThan(decimal.Zero) {
			finalPriceUnit = decimal.Zero
		}

		item.Price = productPrice
		item.DiscountPercent = item.Product.DiscountPercent
		item.DiscountAmount = discountAmountPerUnit
		item.FinalPriceUnit = finalPriceUnit
		item.Subtotal = finalPriceUnit.Mul(decimal.NewFromInt(int64(item.Qty)))

		baseTotalPrice = baseTotalPrice.Add(item.Subtotal)
		totalWeight = totalWeight.Add(productWeight.Mul(decimal.NewFromInt(int64(item.Qty))))
		totalItems += item.Qty
		totalDiscountAmount = totalDiscountAmount.Add(discountAmountPerUnit.Mul(decimal.NewFromInt(int64(item.Qty))))
	}

	cart.BaseTotalPrice = baseTotalPrice
	cart.DiscountAmount = totalDiscountAmount
	cart.TotalWeight = totalWeight
	cart.TotalItems = totalItems

	taxPercent := calc.GetTaxPercent()
	cart.TaxPercent = taxPercent
	cart.TaxAmount = baseTotalPrice.Mul(taxPercent.Div(decimal.NewFromInt(100)))

	cart.GrandTotal = baseTotalPrice.Add(cart.TaxAmount)

}
