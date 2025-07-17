package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type OrderItemRepository interface {
	BulkCreate(ctx context.Context, db *gorm.DB, items []models.OrderItem) error
}

type OrderItemRepositoryImpl struct {
	DB *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &OrderItemRepositoryImpl{DB: db}
}

func (r *OrderItemRepositoryImpl) BulkCreate(ctx context.Context, db *gorm.DB, items []models.OrderItem) error {
	return db.WithContext(ctx).Create(&items).Error
}
