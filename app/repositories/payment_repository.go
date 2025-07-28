package repositories

import (
	"context"
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, tx *gorm.DB, payment *models.Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	UpdatePaymentStatusTx(ctx context.Context, tx *gorm.DB, paymentID string, status string) error
	UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error
}

type PaymentRepositoryImpl struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepositoryImpl {
	return PaymentRepositoryImpl{db}
}

func (r *PaymentRepositoryImpl) Create(ctx context.Context, tx *gorm.DB, payment *models.Payment) error {

	dbInstance := r.DB
	if tx != nil {
		dbInstance = tx
		log.Printf("DEBUG: Using transactional DB instance in PaymentRepository.Create.")
	} else {
		log.Printf("DEBUG: Using direct DB instance in PaymentRepository.Create.")
	}

	result := dbInstance.WithContext(ctx).Create(payment)
	if result.Error != nil {

		return result.Error
	}

	return nil
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

func (r *PaymentRepositoryImpl) UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error {
	return r.DB.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", paymentID).Update("status", status).Error
}

func (r *PaymentRepositoryImpl) UpdatePaymentStatusTx(ctx context.Context, tx *gorm.DB, paymentID string, status string) error {
	return tx.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", paymentID).Update("status", status).Error
}
