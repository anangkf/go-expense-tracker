package repositories

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"

	"gorm.io/gorm"
)

// EXPENSE TEMPLATE SORT ALLOWLIST (query name -> SQL column)
var expenseTemplateSortColumns = map[string]string{
	"id":         "expense_templates.id",
	"name":       "expense_templates.name",
	"amount":     "expense_templates.amount",
	"icon_name":  "expense_templates.icon_name",
	"is_default": "expense_templates.is_default",
	"created_at": "expense_templates.created_at",
	"updated_at": "expense_templates.updated_at",
}

// EXPENSE TEMPLATE FILTER ALLOWLIST (query name -> SQL expression, cast to text for ILIKE)
var expenseTemplateFilterColumns = map[string]string{
	"id":         "expense_templates.id::text",
	"name":       "expense_templates.name",
	"amount":     "expense_templates.amount::text",
	"icon_name":  "expense_templates.icon_name",
	"created_at": "expense_templates.created_at::text",
	"updated_at": "expense_templates.updated_at::text",
}

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
	if value := queryParams.Filters["category_name"]; value != "" {
		query = query.Where(`"Category"."name" ILIKE ?`, "%"+value+"%")
	}
	if value := queryParams.Filters["category_type"]; value != "" {
		query = query.Where(`"Category"."type" = ?`, value)
	}
	for column, value := range utils.SafeFilters(queryParams.Filters, expenseTemplateFilterColumns) {
		if value != "" {
			query = query.Where(column+" ILIKE ?", "%"+value+"%")
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
	order := "desc"
	if strings.ToLower(queryParams.Order) == "asc" {
		order = "asc"
	}
	sortColumn := utils.SafeSort(queryParams.SortBy, expenseTemplateSortColumns, "expense_templates.created_at")
	query = query.Order(sortColumn + " " + order)

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	query = query.Offset(offset).Limit(queryParams.Limit)

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
