package repositories

import (
	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetDefaultCategories() (*[]models.Category, error) {
	var defaultCategories []models.Category

	err := r.db.Where("is_default = ? AND user_id IS NULL", true).Find(&defaultCategories).Error
	if err != nil {
		return nil, err
	}

	return &defaultCategories, nil
}

func (r *CategoryRepository) CreateMany(categories []*models.Category) error {
	return r.db.Create(categories).Error
}

func (r *CategoryRepository) GetByUserID(userID uint) (*[]models.Category, error) {
	var categories []models.Category

	err := r.db.Where("user_id = ?", userID).Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return &categories, nil
}
