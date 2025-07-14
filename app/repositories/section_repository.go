package repositories

import (
	"context"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"gorm.io/gorm"
)

type SectionRepositoryImpl interface {
	GetByID(ctx context.Context, id string) (*models.Section, error)
	GetBySlug(ctx context.Context, slug string) (*models.Section, error)
	GetOrCreateDefaultSection(ctx context.Context) (*models.Section, error)
	GetAll(ctx context.Context) ([]models.Section, error)
}

type sectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) SectionRepositoryImpl {
	return &sectionRepository{db}
}

func (r *sectionRepository) GetByID(ctx context.Context, id string) (*models.Section, error) {
	var section models.Section
	err := r.db.WithContext(ctx).First(&section, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &section, nil
}

func (r *sectionRepository) GetBySlug(ctx context.Context, slug string) (*models.Section, error) {
	var section models.Section
	err := r.db.WithContext(ctx).First(&section, "slug = ?", slug).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &section, nil
}

func (r *sectionRepository) GetOrCreateDefaultSection(ctx context.Context) (*models.Section, error) {
	defaultSlug := "produk-utama"
	var section models.Section
	err := r.db.WithContext(ctx).Where("slug = ?", defaultSlug).First(&section).Error

	if err == gorm.ErrRecordNotFound {

		section = models.Section{
			ID:        "default-section-id",
			Name:      "Produk Utama",
			Slug:      defaultSlug,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if section.ID == "default-section-id" {
			section.ID = "default-section-id"
		}

		if err := r.db.WithContext(ctx).Create(&section).Error; err != nil {
			return nil, err
		}
		return &section, nil
	} else if err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *sectionRepository) GetAll(ctx context.Context) ([]models.Section, error) {
	var sections []models.Section
	err := r.db.WithContext(ctx).Find(&sections).Error
	if err != nil {
		return nil, err
	}
	return sections, nil
}
