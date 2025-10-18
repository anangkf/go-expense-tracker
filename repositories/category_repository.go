package repositories

import (
	"go-expense-tracker-api/middleware"
	"go-expense-tracker-api/models"
	"strings"

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

func (r *CategoryRepository) GetByUserID(userID uint, queryParams middleware.QueryParams) (*[]models.Category, int64, int64, error) {
	var categories []models.Category
	var total int64

	query := r.db.Model(&models.Category{}).Where("user_id = ?", userID)

	// APPLY FILTERS
	for key, value := range queryParams.Filters {
		if value != "" {
			query = query.Where(key+" LIKE ?", "%"+value+"%")
		}
	}

	// COUNT TOTAL RECORDS
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// CALCULATE TOTAL PAGES
	totalPages := int64(total) / int64(queryParams.Limit)
	if int64(total)%int64(queryParams.Limit) != 0 {
		totalPages++
	}

	// APPLY SORTING
	if queryParams.SortBy != "" {
		order := "asc"
		if strings.ToLower(queryParams.Order) == "desc" {
			order = "desc"
		}
		query = query.Order(queryParams.SortBy + " " + order)
	}

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	if err := query.Limit(queryParams.Limit).Offset(offset).Find(&categories).Error; err != nil {
		return nil, 0, 0, err
	}

	return &categories, total, totalPages, nil
}

func (r *CategoryRepository) GetByID(categoryID uint) (*models.Category, error) {
	var category models.Category

	err := r.db.Where("id = ?", categoryID).First(&category).Error
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(categoryID uint) error {
	return r.db.Delete(&models.Category{}, categoryID).Error
}
