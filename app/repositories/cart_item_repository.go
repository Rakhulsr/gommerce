package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type CartItemRepository struct {
	DB *gorm.DB
}

type CartItemRepositoryImpl interface {
	Add(ctx context.Context, item *models.CartItem) error
	Update(ctx context.Context, item *models.CartItem) error
	Delete(ctx context.Context, cartItemID string) error
	GetByID(ctx context.Context, id string) (*models.CartItem, error)
	GetByCartID(ctx context.Context, cartID string) ([]models.CartItem, error)
	GetCartAndProduct(ctx context.Context, cartID, productID string) (*models.CartItem, error)
	ClearCartItems(ctx context.Context, tx *gorm.DB, cartID string) error
	GetByCartIDAndProductID(ctx context.Context, cartID, productID string) (*models.CartItem, error)
	DeleteAllItemsByCartID(ctx context.Context, tx *gorm.DB, cartID string) error
}

func NewCartItemRepository(db *gorm.DB) CartItemRepositoryImpl {
	return &CartItemRepository{db}
}

func (r *CartItemRepository) Add(ctx context.Context, item *models.CartItem) error {
	return r.DB.WithContext(ctx).Create(item).Error
}

func (r *CartItemRepository) Update(ctx context.Context, item *models.CartItem) error {
	return r.DB.WithContext(ctx).Save(item).Error
}

func (r *CartItemRepository) Delete(ctx context.Context, cartItemID string) error {

	return r.DB.WithContext(ctx).Delete(&models.CartItem{}, "id = ?", cartItemID).Error
}

func (r *CartItemRepository) GetByID(ctx context.Context, id string) (*models.CartItem, error) {
	var item models.CartItem
	if err := r.DB.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartItemRepository) GetByCartID(ctx context.Context, cartID string) ([]models.CartItem, error) {
	var items []models.CartItem
	if err := r.DB.WithContext(ctx).Where("cart_id = ?", cartID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CartItemRepository) GetCartAndProduct(ctx context.Context, cartID, productID string) (*models.CartItem, error) {
	var item models.CartItem

	err := r.DB.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *CartItemRepository) ClearCartItems(ctx context.Context, tx *gorm.DB, cartID string) error {

	return tx.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error
}

func (r *CartItemRepository) GetByCartIDAndProductID(ctx context.Context, cartID, productID string) (*models.CartItem, error) {
	var cartItem models.CartItem
	if err := r.DB.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cart item by cart ID and product ID: %w", err)
	}
	return &cartItem, nil
}

func (r *CartItemRepository) DeleteAllItemsByCartID(ctx context.Context, tx *gorm.DB, cartID string) error {
	dbInstance := r.DB
	if tx != nil {
		dbInstance = tx
	}

	result := dbInstance.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&models.CartItem{})
	if result.Error != nil {
		log.Printf("ERROR: Failed to delete cart items for CartID %s: %v", cartID, result.Error)
		return fmt.Errorf("failed to delete cart items: %w", result.Error)
	}

	return nil
}
