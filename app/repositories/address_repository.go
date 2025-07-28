package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, address *models.Address) error
	FindAddressByID(ctx context.Context, id string) (*models.Address, error)
	FindAddressesByUserID(ctx context.Context, userID string) ([]models.Address, error)
	UpdateAddress(ctx context.Context, address *models.Address) error
	DeleteAddress(ctx context.Context, id string) error
	SetPrimaryAddress(ctx context.Context, userID, addressID string) error
	GetPrimaryAddressByUserID(ctx context.Context, userID string) (*models.Address, error)
	SetAllAddressesNonPrimary(ctx context.Context, userID string) error // <-- Metode baru
}

type GormAddressRepository struct {
	db *gorm.DB
}

func NewGormAddressRepository(db *gorm.DB) *GormAddressRepository {
	return &GormAddressRepository{db: db}
}

func (r *GormAddressRepository) CreateAddress(ctx context.Context, address *models.Address) error {
	address.ID = uuid.New().String()
	if address.IsPrimary {

		err := r.db.WithContext(ctx).Model(&models.Address{}).
			Where("user_id = ? AND id != ?", address.UserID, address.ID).
			Update("is_primary", false).Error
		if err != nil {
			log.Printf("GormAddressRepository: Failed to unset primary status for other addresses for user %s: %v", address.UserID, err)
			return fmt.Errorf("failed to unset old primary address: %w", err)
		}
	} else {

		var count int64
		r.db.WithContext(ctx).Model(&models.Address{}).Where("user_id = ?", address.UserID).Count(&count)
		if count == 0 {
			address.IsPrimary = true
		}
	}

	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		log.Printf("GormAddressRepository: Failed to create address for user %s: %v", address.UserID, err)
		return fmt.Errorf("failed to create address: %w", err)
	}
	return nil
}

func (r *GormAddressRepository) FindAddressByID(ctx context.Context, id string) (*models.Address, error) {
	var address models.Address
	if err := r.db.WithContext(ctx).First(&address, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("GormAddressRepository: Failed to find address by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to find address by ID: %w", err)
	}
	return &address, nil
}

func (r *GormAddressRepository) FindAddressesByUserID(ctx context.Context, userID string) ([]models.Address, error) {
	var addresses []models.Address
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_primary DESC, created_at DESC").Find(&addresses).Error; err != nil {
		log.Printf("GormAddressRepository: Failed to find addresses for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to find addresses by user ID: %w", err)
	}
	return addresses, nil
}

func (r *GormAddressRepository) UpdateAddress(ctx context.Context, address *models.Address) error {
	if address.IsPrimary {

		err := r.db.WithContext(ctx).Model(&models.Address{}).
			Where("user_id = ? AND id != ?", address.UserID, address.ID).
			Update("is_primary", false).Error
		if err != nil {
			log.Printf("GormAddressRepository: Failed to unset primary status for other addresses during update for user %s: %v", address.UserID, err)
			return fmt.Errorf("failed to unset old primary address during update: %w", err)
		}
	}
	if err := r.db.WithContext(ctx).Save(address).Error; err != nil {
		log.Printf("GormAddressRepository: Failed to update address %s: %v", address.ID, err)
		return fmt.Errorf("failed to update address: %w", err)
	}
	return nil
}

func (r *GormAddressRepository) DeleteAddress(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.Address{}, "id = ?", id).Error; err != nil {
		log.Printf("GormAddressRepository: Failed to delete address %s: %v", id, err)
		return fmt.Errorf("failed to delete address: %w", err)
	}
	return nil
}

func (r *GormAddressRepository) SetPrimaryAddress(ctx context.Context, userID, addressID string) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.Address{}).Where("user_id = ?", userID).Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		log.Printf("GormAddressRepository: Failed to unset all primary addresses for user %s: %v", userID, err)
		return fmt.Errorf("failed to unset existing primary addresses: %w", err)
	}

	if err := tx.Model(&models.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		log.Printf("GormAddressRepository: Failed to set address %s as primary for user %s: %v", addressID, userID, err)
		return fmt.Errorf("failed to set new primary address: %w", err)
	}

	return tx.Commit().Error
}

func (r *GormAddressRepository) GetPrimaryAddressByUserID(ctx context.Context, userID string) (*models.Address, error) {
	var address models.Address
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_primary = ?", userID, true).First(&address).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("GormAddressRepository: Failed to get primary address for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get primary address: %w", err)
	}
	return &address, nil
}

func (r *GormAddressRepository) SetAllAddressesNonPrimary(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Model(&models.Address{}).Where("user_id = ?", userID).Update("is_primary", false)
	if result.Error != nil {
		log.Printf("GormAddressRepository.SetAllAddressesNonPrimary: Gagal mengatur alamat user %s menjadi non-utama: %v", userID, result.Error)
		return result.Error
	}

	return nil
}
