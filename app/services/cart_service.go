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

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrCartNotFound     = errors.New("cart not found")
	ErrCartItemNotFound = errors.New("cart item not found")
)

type CartService struct {
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	productRepo  repositories.ProductRepositoryImpl
	db           *gorm.DB
}

func NewCartService(cartRepo repositories.CartRepositoryImpl, cartItemRepo repositories.CartItemRepositoryImpl, productRepo repositories.ProductRepositoryImpl, db *gorm.DB) *CartService {
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
		return nil, nil
	}

	if err := s.cartRepo.UpdateCartSummary(ctx, detailedCart.ID); err != nil {
		log.Printf("GetUserCart: Gagal update summary cart %s: %v", detailedCart.ID, err)
	}

	return detailedCart, nil
}

func (s *CartService) AddItemToCart(ctx context.Context, cartID, userID, productID string, qty int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return ErrProductNotFound
		}

		if product.Stock < qty {
			return ErrInsufficientStock
		}

		cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, cartID, userID)
		if err != nil {
			return fmt.Errorf("gagal mendapatkan atau membuat keranjang: %w", err)
		}

		cartItem, err := s.cartItemRepo.GetCartAndProduct(ctx, cart.ID, productID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("gagal memeriksa item keranjang: %w", err)
		}

		if cartItem != nil {

			newQty := cartItem.Qty + qty
			if product.Stock < newQty {
				return ErrInsufficientStock
			}
			cartItem.Qty = newQty

			cartItem.BasePrice = product.Price
			cartItem.TotalPrice = product.Price.Mul(decimal.NewFromInt(int64(newQty)))
			cartItem.BaseTotal = cartItem.BasePrice.Mul(decimal.NewFromInt(int64(newQty)))
			cartItem.TaxPercent = calc.GetTaxPercent()
			cartItem.TaxAmount = calc.CalculateTax(cartItem.BaseTotal.Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(newQty)))))
			cartItem.SubTotal = cartItem.BaseTotal.Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(newQty))))
			cartItem.GrandTotal = cartItem.SubTotal.Add(cartItem.TaxAmount)
			cartItem.UpdatedAt = time.Now()

			if err := s.cartItemRepo.Update(ctx, cartItem); err != nil {
				return err
			}
		} else {

			newCartItem := &models.CartItem{
				ID:         uuid.New().String(),
				CartID:     cart.ID,
				ProductID:  product.ID,
				Qty:        qty,
				BasePrice:  product.Price,
				BaseTotal:  product.Price.Mul(decimal.NewFromInt(int64(qty))),
				TaxPercent: calc.GetTaxPercent(),
				TaxAmount:  calc.CalculateTax(product.Price.Mul(decimal.NewFromInt(int64(qty))).Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(qty))))),
				SubTotal:   product.Price.Mul(decimal.NewFromInt(int64(qty))).Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(qty)))),
				GrandTotal: product.Price.Mul(decimal.NewFromInt(int64(qty))).Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(qty)))).Add(calc.CalculateTax(product.Price.Mul(decimal.NewFromInt(int64(qty))).Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(qty)))))),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := s.cartItemRepo.Add(ctx, newCartItem); err != nil {
				return err
			}
		}

		if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
			log.Printf("AddItemToCart: Gagal update ringkasan cart %s: %v", cart.ID, err)
			return fmt.Errorf("gagal memperbarui ringkasan keranjang: %w", err)
		}

		return nil
	})
}

func (s *CartService) UpdateCartItemQty(ctx context.Context, userID, productID string, newQty int) (*models.Cart, error) {
	var updatedCartResult *models.Cart
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, "", userID)
		if err != nil {
			return fmt.Errorf("failed to get cart: %w", err)
		}
		if cart == nil {
			return ErrCartNotFound
		}

		if newQty <= 0 {

			removedCart, removeErr := s.RemoveItemFromCart(ctx, userID, productID)
			if removeErr != nil {
				return removeErr
			}
			updatedCartResult = removedCart
			return nil
		}

		item, err := s.cartItemRepo.GetCartAndProduct(ctx, cart.ID, productID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCartItemNotFound
			}
			return fmt.Errorf("failed to get cart item: %w", err)
		}
		if item == nil {
			return ErrCartItemNotFound
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return ErrProductNotFound
		}

		if product.Stock < newQty {
			return fmt.Errorf("not enough stock for product %s (available: %d)", product.Name, product.Stock)
		}

		item.Qty = newQty
		item.BasePrice = product.Price
		item.TotalPrice = product.Price.Mul(decimal.NewFromInt(int64(newQty)))
		item.BaseTotal = item.BasePrice.Mul(decimal.NewFromInt(int64(newQty)))
		item.TaxPercent = calc.GetTaxPercent()
		item.TaxAmount = calc.CalculateTax(item.BaseTotal.Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(newQty)))))
		item.SubTotal = item.BaseTotal.Sub(product.DiscountAmount.Mul(decimal.NewFromInt(int64(newQty))))
		item.GrandTotal = item.SubTotal.Add(item.TaxAmount)
		item.UpdatedAt = time.Now()

		if err := s.cartItemRepo.Update(ctx, item); err != nil {
			return fmt.Errorf("failed to update cart item quantity: %w", err)
		}

		if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
			log.Printf("UpdateCartItemQty: Gagal update ringkasan cart %s: %v", cart.ID, err)
			return fmt.Errorf("failed to update cart summary: %w", err)
		}

		updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
		if err != nil {
			log.Printf("UpdateCartItemQty: Gagal mengambil cart yang diperbarui: %v", err)
			return fmt.Errorf("failed to retrieve updated cart: %w", err)
		}
		updatedCartResult = updatedCart
		return nil
	})

	return updatedCartResult, err
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, userID, productID string) (*models.Cart, error) {
	var updatedCartResult *models.Cart
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetOrCreateCartByUserID(ctx, "", userID)
		if err != nil {
			return fmt.Errorf("failed to get cart: %w", err)
		}
		if cart == nil {
			return ErrCartNotFound
		}

		cartItem, err := s.cartItemRepo.GetCartAndProduct(ctx, cart.ID, productID)
		if err != nil || cartItem == nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCartItemNotFound
			}
			return fmt.Errorf("failed to get cart item: %w", err)
		}

		err = s.cartItemRepo.Delete(ctx, cart.ID, productID)
		if err != nil {
			return fmt.Errorf("failed to remove item from cart: %w", err)
		}

		if err := s.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
			log.Printf("RemoveItemFromCart: Gagal update ringkasan cart %s: %v", cart.ID, err)

		}

		updatedCart, err := s.cartRepo.GetCartWithItems(ctx, cart.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				updatedCartResult = nil
				return nil
			}
			log.Printf("RemoveItemFromCart: Gagal mengambil cart yang diperbarui setelah penghapusan: %v", err)
			return fmt.Errorf("failed to retrieve updated cart after removal: %w", err)
		}

		if updatedCart == nil || updatedCart.TotalItems == 0 {
			log.Printf("RemoveItemFromCart: Cart %s kosong, menghapus cart.", cart.ID)
			if err := s.cartRepo.DeleteCart(ctx, tx, cart.ID); err != nil {
				log.Printf("RemoveItemFromCart: Gagal menghapus cart kosong %s: %v", cart.ID, err)
				return fmt.Errorf("failed to delete empty cart: %w", err)
			}
			updatedCartResult = nil
			return nil
		}

		updatedCartResult = updatedCart
		return nil
	})

	return updatedCartResult, err
}

func (s *CartService) ClearCart(ctx context.Context, cartID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cart, err := s.cartRepo.GetCartWithItems(ctx, cartID)
		if err != nil || cart == nil {
			return ErrCartNotFound
		}

		if err := s.cartItemRepo.ClearCartItems(ctx, tx, cartID); err != nil {
			return err
		}

		cart.CalculateTotals()
		err = s.cartRepo.UpdateCart(ctx, cart)
		return err
	})
}
