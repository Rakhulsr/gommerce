package repositories

import (
	"strings"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProducts() ([]models.Product, error)
	GetByCategorySlug(slug string) ([]models.Product, error)
	GetPaginated(limit, offset int) ([]models.Product, int64, error)
	GetByCategorySlugPaginated(slug string, limit, offset int) ([]models.Product, int64, error)
	GetBySlug(slug string) (*models.Product, error)
	GetFeaturedProducts(limit int) ([]models.Product, error)
	SearchProductsPaginated(keyword string, limit, offset int) ([]models.Product, int64, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

func (p *productRepository) GetProducts() ([]models.Product, error) {

	var products []models.Product

	if err := p.db.Debug().Model(&models.Product{}).Limit(20).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (p *productRepository) GetBySlug(slug string) (*models.Product, error) {
	var product models.Product

	if err := p.db.Debug().Model(&models.Product{}).Preload("Categories").Preload("ProductImages").Where("slug = ?", slug).First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *productRepository) GetByCategorySlug(slug string) ([]models.Product, error) {
	var products []models.Product
	err := p.db.
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Preload("ProductImages").
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetPaginated(limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	if err := r.db.Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("Categories").
		Preload("ProductImages").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) GetByCategorySlugPaginated(slug string, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	err := r.db.
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Model(&models.Product{}).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	err = r.db.
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Joins("JOIN categories c ON c.id = pc.category_id").
		Where("c.slug = ?", slug).
		Preload("Categories").
		Preload("ProductImages").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) GetFeaturedProducts(limit int) ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Preload("ProductImages").
		Preload("Categories").
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error

	return products, err
}

func (r *productRepository) SearchProductsPaginated(keyword string, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	searchKeyword := "%" + keyword + "%"

	if err := r.db.Model(&models.Product{}).
		Where("LOWER(name) LIKE ? OR LOWER(short_description) LIKE ?", strings.ToLower(searchKeyword), strings.ToLower(searchKeyword)).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("ProductImages").
		Preload("Categories").
		Where("LOWER(name) LIKE ? OR LOWER(short_description) LIKE ?", strings.ToLower(searchKeyword), strings.ToLower(searchKeyword)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	return products, total, err
}
