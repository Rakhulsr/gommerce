package fakers

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func ProductFaker(db *gorm.DB) *models.Product {
	name := faker.Name()
	user := UserFaker(db)

	if err := db.Debug().Create(user).Error; err != nil {
		log.Fatal("Failed to create user for product faker:", err)

	}

	return &models.Product{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		Sku:              slug.Make(name),
		Name:             name,
		Slug:             slug.Make(name),
		Price:            decimal.NewFromFloat(fakePrice()),
		Weight:           decimal.NewFromFloat(rand.Float64()),
		Stock:            rand.Intn(10),
		ShortDescription: faker.Sentence(),
		Description:      faker.Paragraph(),
		Status:           1,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        gorm.DeletedAt{},
	}
}

func fakePrice() float64 {
	return precision(rand.Float64()*math.Pow10(rand.Intn(8)), rand.Intn(2)+1)
}

func precision(val float64, pre int) float64 {
	a := math.Pow10(pre)
	return float64(int(val*a)) / a

}
