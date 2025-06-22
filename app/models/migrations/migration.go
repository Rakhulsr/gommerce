package migrations

import (
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Address{}, &models.Product{}, &models.ProductImage{}, &models.Section{}, &models.Category{}, &models.Order{}, &models.OrderItem{}, &models.OrderCustomer{}, &models.Payment{}, &models.Shipment{})
}
