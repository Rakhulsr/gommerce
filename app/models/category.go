package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        string    `gorm:"size:36;not null;uniqueIndex;primary_key"`
	Name      string    `gorm:"size:100;not null;uniqueIndex"`
	Slug      string    `gorm:"size:100;not null;uniqueIndex"`
	ParentID  *string   `gorm:"size:36;index"`
	Parent    *Category `gorm:"foreignKey:ParentID"`
	SectionID string    `gorm:"size:36;index"`
	Section   Section
	Products  []Product `gorm:"many2many:product_categories;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
