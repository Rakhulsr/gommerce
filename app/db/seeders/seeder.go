package seeders

import (
	"github.com/Rakhulsr/go-ecommerce/app/db/fakers"
	"gorm.io/gorm"
)

type Seeder struct {
	Seeder interface{}
}

func SeedersRegister(db *gorm.DB) []Seeder {
	return []Seeder{
		{Seeder: fakers.UserFaker(db)},
		{Seeder: fakers.ProductFaker(db)},
	}

}

func DBSeed(db *gorm.DB) error {
	for _, seeder := range SeedersRegister(db) {
		if err := db.Debug().Create(seeder.Seeder).Error; err != nil {
			return err
		}
	}
	return nil
}
