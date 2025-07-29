package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	DeleteProductImage(ctx context.Context, imageID string) error
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

		if err := tx.Omit("Categories").Create(product).Error; err != nil {
			log.Printf("ProductRepository.CreateProduct: Error creating product in DB (omitting categories): %v", err)
			return fmt.Errorf("failed to create product: %w", err)
		}
		log.Printf("ProductRepository.CreateProduct: Product created successfully (without categories): ID=%s, Name=%s", product.ID, product.Name)

		if len(product.Categories) > 0 {
			log.Printf("ProductRepository.CreateProduct: Manually appending categories for product ID: %s. Categories count: %d", product.ID, len(product.Categories))

			for i, cat := range product.Categories {
				log.Printf("ProductRepository.CreateProduct: Category to append [%d]: ID=%s, Name=%s", i, cat.ID, cat.Name)
				if cat.ID == "" {
					log.Printf("ProductRepository.CreateProduct: WARNING! Category at index %d has an empty ID.", i)
					return fmt.Errorf("kategori memiliki ID kosong")
				}

				productCategory := models.ProductCategory{
					ProductID:  product.ID,
					CategoryID: cat.ID,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				if err := tx.Create(&productCategory).Error; err != nil {
					log.Printf("ProductRepository.CreateProduct: Error creating product category entry for ProductID %s, CategoryID %s: %v", product.ID, cat.ID, err)
					return fmt.Errorf("gagal mengasosiasikan kategori secara manual: %w", err)
				}
				log.Printf("ProductRepository.CreateProduct: Category %s associated with product %s.", cat.ID, product.ID)
			}
		} else {
			log.Printf("ProductRepository.CreateProduct: No categories to append for product %s.", product.ID)
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
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("UpdateProduct [ERROR]: Recovered from panic: %v", r)
		}
	}()

	if err := tx.Omit("Categories", "ProductImages").Save(product).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal memperbarui data produk dasar: %w", err)
	}

	if product.Categories == nil {
		product.Categories = []models.Category{}

	}

	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductCategory{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus kategori lama untuk produk: %w", err)
	}

	if len(product.Categories) > 0 {
		var newProductCategories []models.ProductCategory
		for _, cat := range product.Categories {
			newProductCategories = append(newProductCategories, models.ProductCategory{
				ProductID:  product.ID,
				CategoryID: cat.ID,
			})
		}
		if err := tx.CreateInBatches(newProductCategories, 100).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal menambahkan kategori baru ke produk: %w", err)
		}

	}

	if product.ProductImages == nil {
		product.ProductImages = []models.ProductImage{}
	}

	var existingDBImages []models.ProductImage
	if err := tx.Where("product_id = ?", product.ID).Find(&existingDBImages).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal mengambil gambar produk yang sudah ada dari DB: %w", err)
	}

	retainedImageIDsMap := make(map[string]bool)
	for _, img := range product.ProductImages {
		if img.ID != "" {
			retainedImageIDsMap[img.ID] = true
		}
	}

	var imagesToDeleteFromDB []models.ProductImage
	for _, dbImg := range existingDBImages {
		if _, found := retainedImageIDsMap[dbImg.ID]; !found {
			imagesToDeleteFromDB = append(imagesToDeleteFromDB, dbImg)
		}
	}

	if len(imagesToDeleteFromDB) > 0 {
		var idsToDelete []string
		for _, img := range imagesToDeleteFromDB {
			idsToDelete = append(idsToDelete, img.ID)

			if img.Path != "" {

				fullPath := filepath.Join("static", img.Path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					log.Printf("UpdateProduct [INFO]: Gagal menghapus file fisik %s (tidak ditemukan), mungkin sudah dihapus.", fullPath)
				} else if err := os.Remove(fullPath); err != nil {
					log.Printf("UpdateProduct [ERROR]: Gagal menghapus file fisik %s: %v", fullPath, err)

				} else {
					log.Printf("UpdateProduct [INFO]: Berhasil menghapus file fisik: %s", fullPath)
				}
			} else {
				log.Printf("UpdateProduct [WARNING]: Path gambar kosong untuk gambar ID %s, tidak dapat menghapus file fisik.", img.ID)
			}
		}

		if err := tx.Where("id IN (?)", idsToDelete).Delete(&models.ProductImage{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal menghapus gambar produk lama yang tidak dipertahankan dari DB: %w", err)
		}
		log.Printf("UpdateProduct [INFO]: Berhasil menghapus %d gambar lama dari DB untuk produk ID: %s. IDs: %+v", len(idsToDelete), product.ID, idsToDelete)
	} else {
		log.Printf("UpdateProduct [INFO]: Tidak ada gambar lama yang perlu dihapus dari DB untuk produk ID: %s", product.ID)
	}

	var newImagesToCreate []models.ProductImage
	existingDBImageIDsMap := make(map[string]bool)
	for _, img := range existingDBImages {
		existingDBImageIDsMap[img.ID] = true
	}

	for _, img := range product.ProductImages {
		if img.ID == "" {
			img.ID = uuid.New().String()
			img.CreatedAt = time.Now()
			img.UpdatedAt = time.Now()
			img.ProductID = product.ID
			newImagesToCreate = append(newImagesToCreate, img)
		} else {
			if _, found := existingDBImageIDsMap[img.ID]; !found {
				img.CreatedAt = time.Now()
				img.UpdatedAt = time.Now()
				img.ProductID = product.ID
				newImagesToCreate = append(newImagesToCreate, img)

			}
		}
	}

	if len(newImagesToCreate) > 0 {
		if err := tx.CreateInBatches(newImagesToCreate, 100).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal menambahkan gambar produk baru ke DB: %w", err)
		}
		log.Printf("UpdateProduct [INFO]: Berhasil menambahkan %d gambar baru ke DB untuk produk ID: %s", len(newImagesToCreate), product.ID)
	} else {
		log.Printf("UpdateProduct [INFO]: Tidak ada gambar baru yang perlu ditambahkan ke DB untuk produk ID: %s", product.ID)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("gagal melakukan commit transaksi update produk: %w", err)
	}
	log.Printf("UpdateProduct [INFO]: Transaksi update produk berhasil di-commit untuk ID: %s", product.ID)

	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var images []models.ProductImage
	if err := tx.Where("product_id = ?", id).Find(&images).Error; err != nil {
		log.Printf("DeleteProduct: Gagal mengambil gambar produk untuk ID %s (mungkin sudah dihapus): %v", id, err)
	} else {
		for _, img := range images {
			fullPath := filepath.Join(".", img.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				log.Printf("DeleteProduct: File gambar tidak ditemukan, mungkin sudah dihapus: %s", fullPath)
			} else if err := os.Remove(fullPath); err != nil {
				log.Printf("DeleteProduct: Gagal menghapus file fisik gambar %s: %v", fullPath, err)
			} else {
				log.Printf("DeleteProduct: Berhasil menghapus file fisik gambar: %s", fullPath)
			}
		}
	}

	if err := tx.Where("product_id = ?", id).Delete(&models.ProductCategory{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus kategori produk terkait: %w", err)
	}
	log.Printf("DeleteProduct: Asosiasi kategori dihapus dari tabel join untuk produk ID: %s", id)

	if err := tx.Where("product_id = ?", id).Delete(&models.ProductImage{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus gambar produk dari database: %w", err)
	}
	log.Printf("DeleteProduct: Gambar produk dihapus dari database untuk produk ID: %s", id)

	if err := tx.Delete(&models.Product{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal menghapus produk: %w", err)
	}
	log.Printf("DeleteProduct: Produk %s berhasil dihapus dari database.", id)

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("gagal melakukan commit transaksi delete produk: %w", err)
	}
	log.Printf("DeleteProduct: Transaksi committed untuk produk ID: %s", id)

	return nil
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

func (r *productRepository) DeleteProductImage(ctx context.Context, imageID string) error {
	var productImage models.ProductImage
	if err := r.db.WithContext(ctx).Where("id = ?", imageID).First(&productImage).Error; err != nil {
		return fmt.Errorf("gambar produk tidak ditemukan: %w", err)
	}

	if err := r.db.WithContext(ctx).Delete(&productImage).Error; err != nil {
		return fmt.Errorf("gagal menghapus record gambar dari database: %w", err)
	}

	fullPath := filepath.Join(".", productImage.Path)
	if err := os.Remove(fullPath); err != nil {
		log.Printf("DeleteProductImage: Gagal menghapus file fisik %s: %v", fullPath, err)

	}
	return nil
}
