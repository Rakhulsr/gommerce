package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	UpdateStatus(ctx context.Context, orderID, status string) error
}

type PaymentRepositoryImpl struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{DB: db}
}

func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *models.Payment) error {
	return r.DB.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepositoryImpl) FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.DB.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepositoryImpl) UpdateStatus(ctx context.Context, orderID, status string) error {
	return r.DB.WithContext(ctx).
		Model(&models.Payment{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}
