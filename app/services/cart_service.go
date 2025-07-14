package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartService struct {
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	productRepo  repositories.ProductRepositoryImpl
}

func NewCartService(cartRepo repositories.CartRepositoryImpl, cartItemRepo repositories.CartItemRepositoryImpl, productRepo repositories.ProductRepositoryImpl) *CartService {
	return &CartService{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		productRepo:  productRepo,
	}
}

func (s *CartService) AddItemToCart(ctx context.Context, userID, productID string, qty int) (*models.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create cart: %w", err)
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return nil, fmt.Errorf("product not found")
	}

	if product.Stock < qty {
		return nil, fmt.Errorf("not enough stock for product %s", product.Name)
	}

	existingItem, err := s.cartItemRepo.GetCartAndProduct(ctx, cart.ID, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing cart item: %w", err)
	}

	var cartItem *models.CartItem
	if existingItem != nil {
		cartItem = existingItem
		cartItem.Qty += qty
	} else {
		cartItem = &models.CartItem{
			ID:        uuid.New().String(),
			CartID:    cart.ID,
			ProductID: productID,
			Qty:       qty,
			CreatedAt: time.Now(),
		}
	}

	productDiscountAmount := product.DiscountAmount.Mul(decimal.NewFromInt(int64(cartItem.Qty)))

	cartItem.BasePrice = product.Price
	cartItem.BaseTotal = cartItem.BasePrice.Mul(decimal.NewFromInt(int64(cartItem.Qty)))
	cartItem.TaxPercent = calc.GetTaxPercent()
	cartItem.TaxAmount = calc.CalculateTax(cartItem.BaseTotal.Sub(productDiscountAmount))
	cartItem.SubTotal = cartItem.BaseTotal.Sub(productDiscountAmount)
	cartItem.GrandTotal = cartItem.SubTotal.Add(cartItem.TaxAmount)
	cartItem.UpdatedAt = time.Now()

	if existingItem != nil {
		if err := s.cartItemRepo.Update(ctx, cartItem); err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	} else {
		if err := s.cartItemRepo.Add(ctx, cartItem); err != nil {
			return nil, fmt.Errorf("failed to add new cart item: %w", err)
		}
	}

	if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
		log.Printf("AddItemToCart: Gagal update ringkasan cart %s: %v", cart.ID, err)
		return nil, fmt.Errorf("failed to update cart summary: %w", err)
	}

	updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("AddItemToCart: Gagal mengambil cart yang diperbarui: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated cart: %w", err)
	}
	return updatedCart, nil
}

func (s *CartService) UpdateCartItemQty(ctx context.Context, userID, productID string, newQty int) (*models.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if newQty <= 0 {
		return s.RemoveItemFromCart(ctx, userID, productID)
	}

	item, err := s.cartItemRepo.GetCartAndProduct(ctx, cart.ID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cart item not found")
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return nil, fmt.Errorf("product not found for cart item")
	}

	if product.Stock < newQty {
		return nil, fmt.Errorf("not enough stock for product %s (available: %d)", product.Name, product.Stock)
	}

	productDiscountAmount := product.DiscountAmount.Mul(decimal.NewFromInt(int64(newQty)))

	item.Qty = newQty
	item.BasePrice = product.Price
	item.BaseTotal = item.BasePrice.Mul(decimal.NewFromInt(int64(newQty)))
	item.TaxPercent = calc.GetTaxPercent()
	item.TaxAmount = calc.CalculateTax(item.BaseTotal.Sub(productDiscountAmount))
	item.SubTotal = item.BaseTotal.Sub(productDiscountAmount)
	item.GrandTotal = item.SubTotal.Add(item.TaxAmount)
	item.UpdatedAt = time.Now()

	if err := s.cartItemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
		log.Printf("UpdateCartItemQty: Gagal update ringkasan cart %s: %v", cart.ID, err)
		return nil, fmt.Errorf("failed to update cart summary: %w", err)
	}

	updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("UpdateCartItemQty: Gagal mengambil cart yang diperbarui: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated cart: %w", err)
	}
	return updatedCart, nil
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, userID, productID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if err := s.cartItemRepo.Delete(ctx, cart.ID, productID); err != nil {
		return nil, fmt.Errorf("failed to remove item from cart: %w", err)
	}

	if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
		log.Printf("RemoveItemFromCart: Gagal update ringkasan cart %s: %v", cart.ID, err)
	}

	count, err := s.cartRepo.GetCartItemCount(ctx, cart.ID)
	if err != nil {
		log.Printf("RemoveItemFromCart: Gagal mendapatkan hitungan item cart: %v", err)
	}
	if count == 0 {
		log.Printf("RemoveItemFromCart: Cart %s kosong, menghapus cart.", cart.ID)
		if err := s.cartRepo.DeleteCart(ctx, cart.ID); err != nil {
			log.Printf("RemoveItemFromCart: Gagal menghapus cart kosong %s: %v", cart.ID, err)
			return nil, fmt.Errorf("failed to delete empty cart: %w", err)
		}
		return nil, nil
	}

	updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("RemoveItemFromCart: Gagal mengambil cart yang diperbarui: %v", err)
		return nil, fmt.Errorf("failed to retrieve updated cart: %w", err)
	}
	return updatedCart, nil
}

func (s *CartService) GetUserCart(ctx context.Context, userID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user cart: %w", err)
	}
	if cart == nil {
		return nil, nil
	}

	detailedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		log.Printf("GetUserCart: Gagal mengambil detailed cart untuk user %s, cart ID %s: %v", userID, cart.ID, err)
		return nil, fmt.Errorf("failed to get detailed user cart: %w", err)
	}
	if detailedCart == nil {
		return nil, nil
	}

	s.RecalculateCartItemTotals(ctx, detailedCart)
	if err := s.cartRepo.UpdateCartSummary(ctx, detailedCart.ID); err != nil {
		log.Printf("GetUserCart: Gagal update summary cart %s setelah recalculate item: %v", detailedCart.ID, err)
	}

	return detailedCart, nil
}

func (s *CartService) RecalculateCartItemTotals(ctx context.Context, cart *models.Cart) {
	if cart == nil || len(cart.CartItems) == 0 {
		return
	}

	for i := range cart.CartItems {
		item := &cart.CartItems[i]

		if item.Product.ID == "" {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil || product == nil {
				log.Printf("RecalculateCartItemTotals: Product %s not found for cart item %s. Skipping item recalculation.", item.ProductID, item.ID)
				continue
			}
			item.Product = *product
		}

		productDiscountAmount := item.Product.DiscountAmount.Mul(decimal.NewFromInt(int64(item.Qty)))

		item.BasePrice = item.Product.Price
		item.BaseTotal = item.BasePrice.Mul(decimal.NewFromInt(int64(item.Qty)))
		item.TaxPercent = calc.GetTaxPercent()
		item.TaxAmount = calc.CalculateTax(item.BaseTotal.Sub(productDiscountAmount))
		item.SubTotal = item.BaseTotal.Sub(productDiscountAmount)
		item.GrandTotal = item.SubTotal.Add(item.TaxAmount)
		item.UpdatedAt = time.Now()
		if err := s.cartItemRepo.Update(ctx, item); err != nil {
			log.Printf("RecalculateCartItemTotals: Gagal memperbarui cart item %s: %v", item.ID, err)
		}
	}
}
