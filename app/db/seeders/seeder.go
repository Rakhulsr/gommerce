package seeders

// import (
// 	"log"

// 	"github.com/Rakhulsr/go-ecommerce/app/db/fakers"
// 	"github.com/Rakhulsr/go-ecommerce/app/models"
// 	"github.com/google/uuid"
// 	"github.com/gosimple/slug"
// 	"gorm.io/gorm"
// )

// type Seeder struct {
// 	Seeder interface{}
// }

// func SeedersRegister(db *gorm.DB) []Seeder {
// 	return []Seeder{
// 		{Seeder: fakers.UserFaker(db)},
// 	}

// }

// func DBSeed(db *gorm.DB) error {

// 	for _, seeder := range SeedersRegister(db) {
// 		if err := db.Debug().Create(seeder.Seeder).Error; err != nil {
// 			return err
// 		}
// 	}

// 	SeedCategories(db)

// 	var categories []models.Category
// 	if err := db.Find(&categories).Error; err != nil {
// 		log.Fatal("Failed to fetch categories:", err)
// 	}

// 	for _, category := range categories {
// 		for i := 0; i < 5; i++ {
// 			product := fakers.ProductFaker(db, &category)
// 			if err := db.Create(product).Error; err != nil {
// 				log.Fatalf("Failed to seed product for category %s: %v", category.Name, err)
// 			}
// 		}
// 	}

// 	return nil
// }

// func SeedCategories(db *gorm.DB) {
// 	section := models.Section{
// 		Name: "Produk Utama",
// 		Slug: slug.Make("Produk Utama"),
// 	}

// 	if err := db.
// 		Where("slug = ?", section.Slug).
// 		FirstOrCreate(&section).Error; err != nil {
// 		log.Fatalf("Failed to create/find section: %v", err)
// 	}

// 	categories := []string{
// 		"Pakan Hewan Ternak",
// 		"Pakan Hewan Peliharaan",
// 		"Keperluan Memancing",
// 		"Vitamin dan Obat Hewan",
// 		"Aksesoris Hewan",
// 	}

// 	for _, name := range categories {
// 		slugged := slug.Make(name)

// 		var category models.Category
// 		if err := db.
// 			Where("slug = ?", slugged).
// 			First(&category).Error; err != nil {

// 			category = models.Category{
// 				ID:        uuid.NewString(),
// 				Name:      name,
// 				Slug:      slugged,
// 				SectionID: section.ID,
// 			}

// 			if err := db.Create(&category).Error; err != nil {
// 				log.Fatalf("Failed to create category %s: %v", name, err)
// 			}
// 		}
// 	}

// }
