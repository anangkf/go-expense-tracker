package repositories

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"

	"gorm.io/gorm"
)

type ExpenseTemplateRepository struct {
	db *gorm.DB
}

func NewExpenseTemplateRepository(db *gorm.DB) *ExpenseTemplateRepository {
	return &ExpenseTemplateRepository{db: db}
}

func (r *ExpenseTemplateRepository) GetByUserID(userID uint, queryParams utils.QueryParams) (*[]models.ExpenseTemplate, int64, int64, error) {
	var templates []models.ExpenseTemplate
	var total int64

	query := r.db.Model(&models.ExpenseTemplate{}).Where("expense_templates.user_id = ?", userID)
	query = query.Joins("Category")

	// APPLY FILTERS
	for key, value := range queryParams.Filters {
		if value != "" {
			switch key {
			case "category_name":
				query = query.Where(`"Category"."name" ILIKE ?`, "%"+value+"%")
			case "category_type":
				query = query.Where(`"Category"."type" = ?`, value)
			default:
				query = query.Where(key+" ILIKE ?", "%"+value+"%")
			}
		}
	}

	// COUNT TOTAL RECORDS
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// CALCULATE TOTAL PAGES
	totalPages := total / int64(queryParams.Limit)
	if total%int64(queryParams.Limit) != 0 {
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
	orderBy := queryParams.SortBy + " " + queryParams.Order
	query = query.Offset(offset).Limit(queryParams.Limit).Order(orderBy)

	if err := query.Preload("Category").Find(&templates).Error; err != nil {
		return nil, 0, 0, err
	}

	return &templates, total, totalPages, nil
}

func (r *ExpenseTemplateRepository) GetDefaultTemplates() (*[]models.ExpenseTemplate, error) {
	var templates []models.ExpenseTemplate

	err := r.db.Preload("Category.BucketType").Where("is_default = ?", true).Find(&templates).Error
	if err != nil {
		return nil, err
	}

	return &templates, nil
}

func (r *ExpenseTemplateRepository) Create(template *models.ExpenseTemplate) error {
	return r.db.Create(template).Error
}

func (r *ExpenseTemplateRepository) CreateMany(templates []*models.ExpenseTemplate) error {
	return r.db.Create(&templates).Error
}

func (r *ExpenseTemplateRepository) GetByID(id uint) (*models.ExpenseTemplate, error) {
	var template models.ExpenseTemplate

	err := r.db.Preload("Category").Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (r *ExpenseTemplateRepository) Update(template *models.ExpenseTemplate) error {
	return r.db.Save(template).Error
}

func (r *ExpenseTemplateRepository) Delete(template *models.ExpenseTemplate) error {
	return r.db.Delete(&template).Error
}
