package migrations

import (
	"log"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.Product{},
		&models.ProductImage{},
		&models.Section{},
		&models.Category{},
		&models.Payment{},
		&models.Shipment{},
		&models.Cart{},
		&models.CartItem{},
		&models.ProductCategory{},
	)
	if err != nil {
		log.Printf("Error during initial AutoMigrate: %v", err)
		return err
	}

	err = db.AutoMigrate(&models.Order{})
	if err != nil {
		log.Printf("Error during Order AutoMigrate: %v", err)
		return err
	}

	err = db.AutoMigrate(&models.OrderCustomer{})
	if err != nil {
		log.Printf("Error during OrderCustomer AutoMigrate: %v", err)
		return err
	}

	err = db.AutoMigrate(&models.OrderItem{})
	if err != nil {
		log.Printf("Error during OrderCustomer AutoMigrate: %v", err)
		return err
	}
	log.Println("✅ All models migrated successfully.")
	return nil
}
