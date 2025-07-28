package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	_ "gorm.io/gorm/clause"
)

type ProductRepositoryImpl interface {
	GetProducts(ctx context.Context) ([]models.Product, error)
	GetByCategorySlug(ctx context.Context, slug string) ([]models.Product, error)
	GetPaginated(ctx context.Context, limit, offset int) ([]models.Product, int64, error)
	GetByCategorySlugPaginated(ctx context.Context, slug string, limit, offset int) ([]models.Product, int64, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetFeaturedProducts(ctx context.Context, limit int) ([]models.Product, error)
	SearchProductsPaginated(ctx context.Context, keyword string, limit, offset int) ([]models.Product, int64, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetProductCount(ctx context.Context) (int64, error)

	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
	UpdateProductDiscount(ctx context.Context, productID string, discountPercent decimal.Decimal, discountAmount decimal.Decimal) error
	IsSKUExists(ctx context.Context, sku string) (bool, error)
	DecrementStock(ctx context.Context, tx *gorm.DB, productID string, quantity int) error
	UpdateProductTx(ctx context.Context, tx *gorm.DB, product *models.Product) error
	UpdateStock(ctx context.Context, tx *gorm.DB, productID string, newStock int) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepositoryImpl {
	return &productRepository{db}
}

func (p *productRepository) GetProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := p.db.WithContext(ctx).Model(&models.Product{}).Preload("Categories").Preload("ProductImages").Find(&products).Error; err != nil {
		log.Printf("ProductRepository.GetProducts: Error getting products: %v", err)
		return nil, err
	}
	return products, nil
}

func (p *productRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	var product models.Product
	if err := p.db.WithContext(ctx).
		Model(&models.Product{}).
		Preload("Categories").
		Preload("ProductImages").
		Where("id = ?", id).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ProductRepository.GetByID: Product with ID %s not found.", id)
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (p *productRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	if err := p.db.WithContext(ctx).
		Model(&models.Product{}).
		Preload("Categories").
		Preload("ProductImages").
		Where("slug = ?", slug).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ProductRepository.GetBySlug: Product with slug %s not found.", slug)
			return nil, nil
		}

		return nil, err
	}
	return &product, nil
}

func (p *productRepository) GetByCategorySlug(ctx context.Context, slug string) ([]models.Product, error) {
	var products []models.Product
	err := p.db.WithContext(ctx).
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Preload("ProductImages").
		Preload("Categories").
		Find(&products).Error
	if err != nil {
		log.Printf("ProductRepository.GetByCategorySlug: Error getting products by category slug %s: %v", slug, err)
	}
	return products, err
}

func (p *productRepository) GetPaginated(ctx context.Context, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	if err := p.db.WithContext(ctx).Model(&models.Product{}).Count(&total).Error; err != nil {
		log.Printf("ProductRepository.GetPaginated: Error counting products: %v", err)
		return nil, 0, err
	}

	err := p.db.WithContext(ctx).
		Preload("Categories").
		Preload("ProductImages").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	if err != nil {
		log.Printf("ProductRepository.GetPaginated: Error getting paginated products: %v", err)
	}

	return products, total, err
}

func (p *productRepository) GetByCategorySlugPaginated(ctx context.Context, slug string, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	err := p.db.WithContext(ctx).
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Model(&models.Product{}).
		Count(&total).Error
	if err != nil {
		log.Printf("ProductRepository.GetByCategorySlugPaginated: Error counting products by category slug %s: %v", slug, err)
		return nil, 0, err
	}

	err = p.db.WithContext(ctx).
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Preload("Categories").
		Preload("ProductImages").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	if err != nil {
		log.Printf("ProductRepository.GetByCategorySlugPaginated: Error getting paginated products by category slug %s: %v", slug, err)
	}

	return products, total, err
}

func (p *productRepository) GetFeaturedProducts(ctx context.Context, limit int) ([]models.Product, error) {
	var products []models.Product
	err := p.db.WithContext(ctx).
		Preload("ProductImages").
		Preload("Categories").
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error
	if err != nil {
		log.Printf("ProductRepository.GetFeaturedProducts: Error getting featured products: %v", err)
	}
	return products, err
}

func (p *productRepository) SearchProductsPaginated(ctx context.Context, keyword string, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	searchKeyword := "%" + strings.ToLower(keyword) + "%"

	if err := p.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("LOWER(name) LIKE ?", searchKeyword).
		Count(&total).Error; err != nil {
		log.Printf("ProductRepository.SearchProductsPaginated: Error counting search results for keyword '%s': %v", keyword, err)
		return nil, 0, err
	}

	err := p.db.WithContext(ctx).
		Preload("ProductImages").
		Preload("Categories").
		Where("LOWER(name) LIKE ?", searchKeyword).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	if err != nil {
		log.Printf("ProductRepository.SearchProductsPaginated: Error getting paginated search results for keyword '%s': %v", keyword, err)
	}

	return products, total, err
}

func (p *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(product).Error; err != nil {
			log.Printf("ProductRepository.CreateProduct: Error creating product in DB: %v", err)
			return fmt.Errorf("failed to create product: %w", err)
		}

		for i := range product.ProductImages {
			product.ProductImages[i].ProductID = product.ID

			product.ProductImages[i].ID = uuid.New().String()

			if err := tx.Create(&product.ProductImages[i]).Error; err != nil {
				log.Printf("ProductRepository.CreateProduct: Error creating product image %d for product %s (ImageID: %s): %v", i, product.ID, product.ProductImages[i].ID, err)
				return fmt.Errorf("failed to create product image: %w", err)
			}
			log.Printf("ProductRepository.CreateProduct: Product image %s created for product %s.", product.ProductImages[i].ID, product.ID)
		}

		if len(product.Categories) > 0 {
			if err := tx.Model(product).Association("Categories").Append(product.Categories); err != nil {
				log.Printf("ProductRepository.CreateProduct: Error appending categories for product %s: %v", product.ID, err)
				return fmt.Errorf("failed to associate categories: %w", err)
			}
			log.Printf("ProductRepository.CreateProduct: Categories appended for product %s.", product.ID)
		}

		return nil
	})

	if err != nil {
		log.Printf("ProductRepository.CreateProduct: Transaction failed for product %s: %v", product.ID, err)
	} else {
		log.Printf("ProductRepository.CreateProduct: Transaction committed successfully for product %s.", product.ID)
	}
	return err
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	if product.ID == "" {
		log.Printf("ProductRepository.UpdateProduct: ERROR: Product ID is empty.")
		return fmt.Errorf("product ID is empty for update")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Save(product).Error; err != nil {
			log.Printf("ProductRepository.UpdateProduct: Error saving product %s in DB: %v", product.ID, err)
			return fmt.Errorf("failed to save product: %w", err)
		}
		log.Printf("ProductRepository.UpdateProduct: Product %s saved successfully (basic fields).", product.ID)

		if err := tx.Model(product).Association("Categories").Replace(product.Categories); err != nil {
			log.Printf("ProductRepository.UpdateProduct: Error replacing categories for product %s: %v", product.ID, err)
			return fmt.Errorf("failed to replace categories: %w", err)
		}

		if len(product.ProductImages) > 0 {
			log.Printf("ProductRepository.UpdateProduct: New image path provided in form.")
			newImage := product.ProductImages[0]

			var existingImage models.ProductImage
			result := tx.Where("product_id = ?", product.ID).First(&existingImage)

			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Printf("ProductRepository.UpdateProduct: Error checking for existing image for product %s: %v", product.ID, result.Error)
				return fmt.Errorf("failed to check existing product image: %w", result.Error)
			}

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Printf("ProductRepository.UpdateProduct: No existing image found for product %s, creating new one.", product.ID)
				newImage.ID = uuid.New().String()
				newImage.ProductID = product.ID
				if err := tx.Create(&newImage).Error; err != nil {
					log.Printf("ProductRepository.UpdateProduct: Error creating new product image %s for product %s: %v", newImage.ID, product.ID, err)
					return fmt.Errorf("failed to create new product image: %w", err)
				}
				log.Printf("ProductRepository.UpdateProduct: New product image %s created for product %s.", newImage.ID, product.ID)
			} else {
				log.Printf("ProductRepository.UpdateProduct: Existing image found (%s) for product %s, updating it.", existingImage.ID, product.ID)
				existingImage.Path = newImage.Path
				existingImage.ExtraLarge = newImage.ExtraLarge
				existingImage.Large = newImage.Large
				existingImage.Medium = newImage.Medium
				existingImage.Small = newImage.Small
				existingImage.UpdatedAt = time.Now()

				if err := tx.Save(&existingImage).Error; err != nil {
					log.Printf("ProductRepository.UpdateProduct: Error updating existing product image %s for product %s: %v", existingImage.ID, product.ID, err)
					return fmt.Errorf("failed to update existing product image: %w", err)
				}
				log.Printf("ProductRepository.UpdateProduct: Existing product image %s updated for product %s.", existingImage.ID, product.ID)
			}
		} else {
			log.Printf("ProductRepository.UpdateProduct: No new image path provided in form. Existing images (if any) will be retained.")
		}

		return nil
	})
}

func (p *productRepository) DeleteProduct(ctx context.Context, id string) error {
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("product_id = ?", id).Delete(&models.ProductImage{}).Error; err != nil {
			log.Printf("ProductRepository.DeleteProduct: Error deleting product images for product %s: %v", id, err)
			return fmt.Errorf("failed to delete product images: %w", err)
		}

		var product models.Product
		product.ID = id
		if err := tx.Model(&product).Association("Categories").Clear(); err != nil {
			log.Printf("ProductRepository.DeleteProduct: Error clearing categories association for product %s: %v", id, err)
			return fmt.Errorf("failed to clear categories association: %w", err)
		}

		if err := tx.Delete(&models.Product{}, "id = ?", id).Error; err != nil {
			log.Printf("ProductRepository.DeleteProduct: Error deleting product %s from DB: %v", id, err)
			return fmt.Errorf("failed to delete product: %w", err)
		}
		return nil
	})

	if err != nil {
		log.Printf("ProductRepository.DeleteProduct: Transaction failed for product %s: %v", id, err)
	}
	return err
}

func (r *productRepository) UpdateProductDiscount(ctx context.Context, productID string, discountPercent decimal.Decimal, discountAmount decimal.Decimal) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"discount_percent": discountPercent,
			"discount_amount":  discountAmount,
			"updated_at":       time.Now(),
		}).Error
}

func (p *productRepository) IsSKUExists(ctx context.Context, sku string) (bool, error) {
	var count int64
	err := p.db.WithContext(ctx).Model(&models.Product{}).Where("sku = ?", sku).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *productRepository) DecrementStock(ctx context.Context, tx *gorm.DB, productID string, quantity int) error {

	result := tx.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.Error != nil {
		return fmt.Errorf("failed to decrement stock for product %s: %w", productID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no product found with ID %s to decrement stock", productID)
	}
	return nil
}

func (r *productRepository) UpdateProductTx(ctx context.Context, tx *gorm.DB, product *models.Product) error {
	return tx.WithContext(ctx).Save(product).Error
}

func (r *productRepository) UpdateStock(ctx context.Context, tx *gorm.DB, productID string, newStock int) error {
	return tx.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("stock", newStock).Error
}

func (r *productRepository) GetProductCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Count(&count).Error; err != nil {
		log.Printf("ProductRepository.GetProductCount: Failed to count products: %v", err)
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}
