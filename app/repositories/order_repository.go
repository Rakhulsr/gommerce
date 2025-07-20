package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, tx *gorm.DB, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	FindByCode(ctx context.Context, orderCode string) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID string, status int) error
	UpdatePaymentStatus(ctx context.Context, db *gorm.DB, orderID, paymentStatus string) error
	UpdatePaymentStatusAndOrderStatus(ctx context.Context, db *gorm.DB, orderID, paymentStatus string, orderStatus int) error
	GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error)

	GetAllOrders(ctx context.Context) ([]models.Order, error)
	UpdateMidtransDetails(ctx context.Context, db *gorm.DB, orderID, transactionToken, paymentURL string) error
	GetOrderByIDWithRelations(ctx context.Context, orderID string) (*models.Order, error)
	FindByCodeWithDetails(ctx context.Context, orderCode string) (*models.Order, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Order, error)
}

type gormOrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &gormOrderRepository{db: db}
}

func (r *gormOrderRepository) Create(ctx context.Context, tx *gorm.DB, order *models.Order) error {
	return tx.WithContext(ctx).Create(order).Error
}

func (r *gormOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).Preload("OrderItems.Product.ProductImages").Preload("Address").First(&order, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) FindByCode(ctx context.Context, orderCode string) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).Preload("OrderItems.Product.ProductImages").Preload("Address").First(&order, "order_code = ?", orderCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) Update(ctx context.Context, order *models.Order) error {
	order.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *gormOrderRepository) UpdateStatus(ctx context.Context, orderID string, status int) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (r *gormOrderRepository) UpdatePaymentStatus(ctx context.Context, db *gorm.DB, orderID, paymentStatus string) error {
	return db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", paymentStatus).Error
}

func (r *gormOrderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	var orders []models.Order

	err := r.db.WithContext(ctx).Preload("OrderItems.Product.ProductImages").Preload("Address").Where("user_id = ?", userID).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *gormOrderRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	err := r.db.WithContext(ctx).Preload("OrderItems.Product.ProductImages").Preload("Address").Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *gormOrderRepository) UpdateMidtransDetails(ctx context.Context, db *gorm.DB, orderID, transactionToken, paymentURL string) error {
	return db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"midtrans_transaction_id": transactionToken,
		"midtrans_payment_url":    paymentURL,
		"updated_at":              time.Now(),
	}).Error
}
func (r *gormOrderRepository) GetOrderByIDWithRelations(ctx context.Context, orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.Product.ProductImages").
		Preload("Address").
		First(&order, "id = ?", orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order with relations: %w", err)
	}
	return &order, nil
}

func (r *gormOrderRepository) UpdatePaymentStatusAndOrderStatus(ctx context.Context, db *gorm.DB, orderID, paymentStatus string, orderStatus int) error {
	return db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"payment_status": paymentStatus,
		"status":         orderStatus,
		"updated_at":     time.Now(),
	}).Error
}

func (r *gormOrderRepository) FindByCodeWithDetails(ctx context.Context, orderCode string) (*models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.Product.ProductImages").
		Preload("Address").
		Where("order_code = ?", orderCode).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *gormOrderRepository) FindByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	var orders []models.Order

	err := r.db.WithContext(ctx).
		Preload("OrderItems.Product.ProductImages").
		Preload("Address").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}
	return orders, nil
}
