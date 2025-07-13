package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartRepositoryImpl interface {
	GetCartWithItems(ctx context.Context, cartID string) (*models.Cart, error)
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	CreateCart(ctx context.Context, cart *models.Cart) error
	UpdateCartSummary(ctx context.Context, cartID string) error
	GetCartItemCount(ctx context.Context, cartID string) (int, error)

	GetOrCreateCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	DeleteCart(ctx context.Context, cartID string) error
	CreateCartForUser(ctx context.Context, userID string) (*models.Cart, error)
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

// func (r *cartRepository) AddItemToCart(ctx context.Context, cartID string, item models.CartItem) error {
// 	item.CartID = cartID
// 	return r.db.WithContext(ctx).Create(&item).Error
// }

func (r *cartRepository) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).First(&cart, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Tidak ditemukan, bukan error
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(ctx context.Context, cart *models.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *cartRepository) UpdateCartSummary(ctx context.Context, cartID string) error {
	var items []models.CartItem

	if err := r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Find(&items).Error; err != nil {
		return err
	}

	var baseTotal, taxTotal, discountTotal, grandTotal decimal.Decimal

	for _, item := range items {
		baseTotal = baseTotal.Add(item.BaseTotal)
		taxTotal = taxTotal.Add(item.TaxAmount)
		discountTotal = discountTotal.Add(item.DiscountAmount)
		grandTotal = grandTotal.Add(item.GrandTotal)
	}

	var taxPercent, discountPercent decimal.Decimal
	if len(items) > 0 {
		taxPercent = items[0].TaxPercent
		discountPercent = items[0].DiscountPercent
	}

	return r.db.WithContext(ctx).
		Model(&models.Cart{}).
		Where("id = ?", cartID).
		Updates(models.Cart{
			BaseTotalPrice:  baseTotal,
			TaxAmount:       taxTotal,
			TaxPercent:      taxPercent,
			DiscountAmount:  discountTotal,
			DiscountPercent: discountPercent,
			GrandTotal:      grandTotal,
		}).Error
}

func (r *cartRepository) DeleteCart(ctx context.Context, cartID string) error {

	items, err := r.cartItemRepo.GetByCartID(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to get cart items for deletion of cart %s: %w", cartID, err)
	}

	for _, item := range items {
		if err := r.cartItemRepo.Delete(ctx, item.CartID, item.ProductID); err != nil {
			log.Printf("Warning: Failed to delete cart item %s from cart %s during cart deletion: %v", item.ID, cartID, err)

		}
	}

	if err := r.db.WithContext(ctx).Delete(&models.Cart{}, "id = ?", cartID).Error; err != nil {
		return fmt.Errorf("failed to delete cart %s: %w", cartID, err)
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
		UserID: userID, // <-- SET UserID
		// CreatedAt: time.Now(),
		// UpdatedAt: time.Now(),
	}

	result := r.db.WithContext(ctx).Create(newCart)
	if result.Error != nil {
		log.Printf("CartRepository: Failed to create cart for user %s: %v", userID, result.Error)
		return nil, result.Error
	}
	log.Printf("CartRepository: Successfully created cart with ID: %s for user %s", newCart.ID, userID)
	return newCart, nil
}

func (r *cartRepository) GetOrCreateCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Cart not found, create a new one
			newCart := &models.Cart{
				ID:     uuid.New().String(),
				UserID: userID,
				// CreatedAt: time.Now(),
				// UpdatedAt: time.Now(),
			}
			if createErr := r.db.WithContext(ctx).Create(newCart).Error; createErr != nil {
				log.Printf("CartRepository: Failed to create cart for user %s: %v", userID, createErr)
				return nil, createErr
			}
			log.Printf("CartRepository: Created new cart %s for user %s", newCart.ID, userID)
			return newCart, nil
		}
		log.Printf("CartRepository: Error finding cart by user ID %s: %v", userID, err)
		return nil, err
	}
	log.Printf("CartRepository: Found existing cart %s for user %s", cart.ID, userID)
	return &cart, nil
}
