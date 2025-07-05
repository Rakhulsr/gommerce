package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetCartWithItems(ctx context.Context, cartID string) (models.Cart, error)
	AddItemToCart(ctx context.Context, cartID string, item models.CartItem) error
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	CreateCart(ctx context.Context, cart *models.Cart) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db}
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, cartID string) (models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Preload("CartItems.Product").
		FirstOrCreate(&cart, models.Cart{ID: cartID}).Error
	return cart, err
}

func (r *cartRepository) AddItemToCart(ctx context.Context, cartID string, item models.CartItem) error {
	item.CartID = cartID
	return r.db.WithContext(ctx).Create(&item).Error
}

func (r *cartRepository) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	var cart models.Cart
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(ctx context.Context, cart *models.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}
