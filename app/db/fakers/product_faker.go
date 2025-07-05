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

func ProductFaker(db *gorm.DB, category *models.Category) *models.Product {
	name := faker.Name()

	user := UserFaker(db)
	if err := db.Debug().FirstOrCreate(user, "email = ?", user.Email).Error; err != nil {
		log.Fatal("Failed to create/find user:", err)
	}

	productID := uuid.New().String()
	slugText := slug.Make(name + "-" + uuid.NewString()[:6])

	imagePaths := []string{
		"/images/products/ss.jpg",
		"/images/products/ss1.jpg",
		"/images/products/ss2.jpg",
	}

	numImages := rand.Intn(3) + 1
	productImages := make([]models.ProductImage, numImages)

	for i := 0; i < numImages; i++ {
		img := imagePaths[rand.Intn(len(imagePaths))]

		productImages[i] = models.ProductImage{
			ID:         uuid.New().String(),
			ProductID:  productID,
			Path:       img,
			ExtraLarge: img,
			Large:      img,
			Medium:     img,
			Small:      img,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}

	var discountPercent decimal.Decimal
	if rand.Intn(2) == 0 {
		discountPercent = decimal.NewFromInt(0)
	} else {
		discountPercent = decimal.NewFromInt(10)
	}

	product := &models.Product{
		ID:               productID,
		UserID:           user.ID,
		Sku:              slug.Make(name),
		Name:             name,
		Slug:             slugText,
		Price:            decimal.NewFromFloat(fakePrice()),
		Weight:           decimal.NewFromFloat(rand.Float64() * 5),
		Stock:            rand.Intn(20) + 1,
		ShortDescription: faker.Sentence(),
		Description:      faker.Paragraph(),
		Status:           1,
		Categories:       []models.Category{*category},
		DiscountPercent:  discountPercent,
		ProductImages:    productImages,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return product
}

func fakePrice() float64 {

	min := 10000.0
	max := 5000000.0

	price := min + rand.Float64()*(max-min)
	return precision(price, 2)
}

func precision(val float64, pre int) float64 {
	a := math.Pow10(pre)
	return float64(int(val*a)) / a

}
