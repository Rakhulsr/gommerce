package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type OrderCustomerRepository interface {
	Create(ctx context.Context, db *gorm.DB, customer *models.OrderCustomer) error
}

type OrderCustomerRepositoryImpl struct {
	DB *gorm.DB
}

func NewOrderCustomerRepository(db *gorm.DB) OrderCustomerRepository {
	return &OrderCustomerRepositoryImpl{DB: db}
}

func (r *OrderCustomerRepositoryImpl) Create(ctx context.Context, db *gorm.DB, customer *models.OrderCustomer) error {
	return db.WithContext(ctx).Create(customer).Error
}
