package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
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
		if err == gorm.ErrRecordNotFound {
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

	cart.CalculateTotals()

	err = r.UpdateCart(ctx, cart)
	if err != nil {
		return fmt.Errorf("gagal memperbarui ringkasan keranjang: %w", err)
	}
	log.Printf("UpdateCartSummary: Ringkasan keranjang untuk ID %s berhasil diperbarui. TotalItems: %d", cartID, cart.TotalItems)
	return nil
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
	log.Printf("CartRepository: Successfully created cart with ID: %s for user %s", newCart.ID, userID)
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
			TotalWeight:     0,
			TotalItems:      0,
		}
		createdCart, createErr := r.AddCart(ctx, newCart)
		if createErr != nil {
			return nil, fmt.Errorf("gagal membuat keranjang baru: %w", createErr)
		}
		cart = createdCart
		log.Printf("CartRepository: Keranjang baru dibuat untuk UserID: %s, ID: %s", userID, cart.ID)
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
