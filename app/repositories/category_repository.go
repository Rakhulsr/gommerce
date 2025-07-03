package repositories

import (
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	GetAll() ([]models.Category, error)
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db}
}

func (c *categoryRepo) GetAll() ([]models.Category, error) {
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
