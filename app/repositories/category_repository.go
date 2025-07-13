package repositories

import (
	"context"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type CategoryRepositoryImpl interface {
	GetAll(ctx context.Context) ([]models.Category, error)
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepositoryImpl {
	return &categoryRepo{db}
}

func (c *categoryRepo) GetAll(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category

	err := c.db.
		Model(&models.Category{}).
		Select("id, name, slug").
		Group("slug, id, name").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, err
}
