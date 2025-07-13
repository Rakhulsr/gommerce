package repositories

import (
	"context"
	"strings"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
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
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepositoryImpl {
	return &productRepository{db}
}

func (p *productRepository) GetProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := p.db.WithContext(ctx).Model(&models.Product{}).Limit(20).Find(&products).Error; err != nil {
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
		Find(&products).Error
	return products, err
}

func (p *productRepository) GetPaginated(ctx context.Context, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	if err := p.db.WithContext(ctx).Model(&models.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := p.db.WithContext(ctx).
		Preload("Categories").
		Preload("ProductImages").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

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
	return products, err
}

func (p *productRepository) SearchProductsPaginated(ctx context.Context, keyword string, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	searchKeyword := "%" + strings.ToLower(keyword) + "%"

	if err := p.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("LOWER(name) LIKE ? OR LOWER(short_description) LIKE ?", searchKeyword, searchKeyword).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := p.db.WithContext(ctx).
		Preload("ProductImages").
		Preload("Categories").
		Where("LOWER(name) LIKE ? OR LOWER(short_description) LIKE ?", searchKeyword, searchKeyword).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	return products, total, err
}
