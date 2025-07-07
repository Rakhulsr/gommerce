package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetCartWithItems(ctx context.Context, cartID string) (models.Cart, error)
	AddItemToCart(ctx context.Context, cartID string, item models.CartItem) error
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	CreateCart(ctx context.Context, cart *models.Cart) error
	UpdateCartSummary(ctx context.Context, cartID string) error
	GetCartItemCount(ctx context.Context, cartID string) (int, error)
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
		Preload("CartItems.Product.ProductImages").
		Preload("CartItems.Product").
		Preload("CartItems").
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

func (r *cartRepository) GetCartItemCount(ctx context.Context, cartID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("cart_items").
		Where("cart_id = ?", cartID).
		Count(&count).Error

	return int(count), err
}
