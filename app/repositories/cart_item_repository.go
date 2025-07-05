package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type CartItemRepository struct {
	DB *gorm.DB
}

func NewCartItemRepository(db *gorm.DB) CartItemRepository {
	return CartItemRepository{DB: db}
}

func (r *CartItemRepository) Add(ctx context.Context, item *models.CartItem) error {
	return r.DB.WithContext(ctx).Create(item).Error
}

func (r *CartItemRepository) Update(ctx context.Context, item *models.CartItem) error {
	return r.DB.WithContext(ctx).Save(item).Error
}

func (r *CartItemRepository) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Delete(&models.CartItem{}, "id = ?", id).Error
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
