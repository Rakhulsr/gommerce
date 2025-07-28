package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartRepositoryImpl interface {
	GetCartWithItems(ctx context.Context, cartID string) (*models.Cart, error)
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	GetOrCreateCartByUserID(ctx context.Context, cartID, userID string) (*models.Cart, error)
	GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	UpdateCartSummary(ctx context.Context, cartID string) error
	GetCartItemCount(ctx context.Context, cartID string) (int, error)
	UpdateCart(ctx context.Context, cart *models.Cart) error
	DeleteCart(ctx context.Context, db *gorm.DB, cartID string) error
	CreateCartForUser(ctx context.Context, userID string) (*models.Cart, error)
	GetAllCarts(ctx context.Context) ([]models.Cart, error)
	AddCart(ctx context.Context, cart *models.Cart) (*models.Cart, error)
	GetByUserIDWithItems(ctx context.Context, userID string) (*models.Cart, error)
	UpdateCartTotalPrice(ctx context.Context, tx *gorm.DB, cartID string, baseTotalPrice, taxAmount, taxPercent, discountAmount, discountPercent decimal.Decimal, totalItems int) error
	ResetCartTotals(ctx context.Context, tx *gorm.DB, cartID string) error
}

type cartRepository struct {
	cartItemRepo CartItemRepositoryImpl
	db           *gorm.DB
}

func NewCartRepository(db *gorm.DB, cItemRepo CartItemRepositoryImpl) CartRepositoryImpl {
	return &cartRepository{db: db, cartItemRepo: cItemRepo}
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, cartID string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Preload("CartItems.Product.ProductImages").
		Preload("CartItems.Product").
		Preload("CartItems").
		Where("id = ?", cartID).
		First(&cart).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).First(&cart, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) AddCart(ctx context.Context, cart *models.Cart) (*models.Cart, error) {
	result := r.db.WithContext(ctx).Create(cart)
	if result.Error != nil {
		return nil, result.Error
	}
	return cart, nil
}

func (r *cartRepository) UpdateCartSummary(ctx context.Context, cartID string) error {
	cart, err := r.GetCartWithItems(ctx, cartID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("UpdateCartSummary: Keranjang dengan ID %s tidak ditemukan. Mungkin sudah dihapus.", cartID)
			return nil
		}
		return fmt.Errorf("gagal mengambil keranjang dengan item untuk ringkasan: %w", err)
	}
	if cart == nil {
		log.Printf("UpdateCartSummary: Keranjang dengan ID %s kosong setelah diambil. Mungkin sudah dihapus.", cartID)
		return nil
	}

	cart.CalculateTotals(calc.GetTaxPercent())

	err = r.UpdateCart(ctx, cart)
	if err != nil {
		return fmt.Errorf("gagal memperbarui ringkasan keranjang: %w", err)
	}

	return nil
}

func (r *cartRepository) UpdateCartTotalPrice(ctx context.Context, tx *gorm.DB, cartID string, baseTotalPrice, taxAmount, taxPercent, discountAmount, discountPercent decimal.Decimal, totalItems int) error {

	grandTotal := baseTotalPrice.Add(taxAmount).Sub(discountAmount)

	return tx.WithContext(ctx).Model(&models.Cart{}).Where("id = ?", cartID).Updates(map[string]interface{}{
		"base_total_price": baseTotalPrice,
		"tax_amount":       taxAmount,
		"tax_percent":      taxPercent,
		"discount_amount":  discountAmount,
		"discount_percent": discountPercent,
		"grand_total":      grandTotal,
		"total_items":      totalItems,
		"shipping_cost":    decimal.Zero,
		"updated_at":       time.Now(),
	}).Error
}

func (r *cartRepository) ResetCartTotals(ctx context.Context, tx *gorm.DB, cartID string) error {
	return tx.WithContext(ctx).Model(&models.Cart{}).Where("id = ?", cartID).Updates(map[string]interface{}{
		"base_total_price":      decimal.Zero,
		"tax_amount":            decimal.Zero,
		"tax_percent":           decimal.Zero,
		"discount_amount":       decimal.Zero,
		"discount_percent":      decimal.Zero,
		"shipping_cost":         decimal.Zero,
		"grand_total":           decimal.Zero,
		"total_weight":          decimal.Zero,
		"total_items":           0,
		"shipping_service":      "",
		"shipping_service_code": "",
		"shipping_service_name": "",
		"updated_at":            time.Now(),
	}).Error
}

func (r *cartRepository) DeleteCart(ctx context.Context, tx *gorm.DB, cartID string) error {

	result := tx.WithContext(ctx).Where("id = ?", cartID).Delete(&models.Cart{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cart not found or already deleted")
	}
	return nil
}

func (r *cartRepository) GetCartItemCount(ctx context.Context, cartID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("cart_items").
		Where("cart_id = ?", cartID).
		Count(&count).Error

	return int(count), err
}

func (r *cartRepository) CreateCartForUser(ctx context.Context, userID string) (*models.Cart, error) {
	newCart := &models.Cart{
		ID:     uuid.New().String(),
		UserID: userID,
	}

	result := r.db.WithContext(ctx).Create(newCart)
	if result.Error != nil {
		log.Printf("CartRepository: Failed to create cart for user %s: %v", userID, result.Error)
		return nil, result.Error
	}

	return newCart, nil
}

func (r *cartRepository) UpdateCart(ctx context.Context, cart *models.Cart) error {
	cart.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(cart).Error
	if err != nil {
		log.Printf("CartRepository.UpdateCart: Error updating cart %s: %v", cart.ID, err)
		return fmt.Errorf("failed to update cart: %w", err)
	}
	return nil
}

func (r *cartRepository) GetAllCarts(ctx context.Context) ([]models.Cart, error) {
	var carts []models.Cart

	if err := r.db.WithContext(ctx).Preload("CartItems.Product.ProductImages").Preload("CartItems.Product").Find(&carts).Error; err != nil {
		log.Printf("CartRepository.GetAllCarts: Error getting all carts: %v", err)
		return nil, fmt.Errorf("failed to get all carts: %w", err)
	}
	return carts, nil
}

func (r *cartRepository) GetOrCreateCartByUserID(ctx context.Context, cartID, userID string) (*models.Cart, error) {
	var cart *models.Cart
	var err error

	if cartID != "" {
		cart, err = r.GetByID(ctx, cartID)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan keranjang berdasarkan ID: %w", err)
		}
		if cart != nil && cart.UserID != userID {

			log.Printf("CartRepository: CartID %s di sesi bukan milik UserID %s. Mencari berdasarkan UserID.", cartID, userID)
			cart = nil
		}
	}

	if cart == nil && userID != "" {
		cart, err = r.GetCartByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan keranjang berdasarkan UserID: %w", err)
		}
	}

	if cart == nil {

		newCart := &models.Cart{
			ID:              uuid.New().String(),
			UserID:          userID,
			BaseTotalPrice:  decimal.Zero,
			TaxAmount:       decimal.Zero,
			TaxPercent:      decimal.Zero,
			DiscountAmount:  decimal.Zero,
			DiscountPercent: decimal.Zero,
			GrandTotal:      decimal.Zero,
			TotalWeight:     decimal.Zero,
			TotalItems:      0,
		}
		createdCart, createErr := r.AddCart(ctx, newCart)
		if createErr != nil {
			return nil, fmt.Errorf("gagal membuat keranjang baru: %w", createErr)
		}
		cart = createdCart

	}

	return cart, nil
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &cart, nil
}

func (r *cartRepository) GetByUserIDWithItems(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart
	if err := r.db.WithContext(ctx).
		Preload("CartItems.Product.ProductImages").
		Preload("CartItems.Product").
		Where("user_id = ?", userID).
		First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cart by user ID with items: %w", err)
	}
	return &cart, nil
}
