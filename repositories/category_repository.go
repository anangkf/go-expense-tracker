package repositories

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CATEGORY SORT ALLOWLIST (query name -> SQL column)
var categorySortColumns = map[string]string{
	"id":         "categories.id",
	"name":       "categories.name",
	"type":       "categories.type",
	"icon_name":  "categories.icon_name",
	"is_default": "categories.is_default",
	"created_at": "categories.created_at",
	"updated_at": "categories.updated_at",
}

// CATEGORY FILTER ALLOWLIST (query name -> SQL expression, cast to text for ILIKE)
var categoryFilterColumns = map[string]string{
	"id":         "categories.id::text",
	"name":       "categories.name",
	"type":       "categories.type",
	"icon_name":  "categories.icon_name",
	"created_at": "categories.created_at::text",
	"updated_at": "categories.updated_at::text",
}

func applyCategoryFilters(query *gorm.DB, queryParams utils.QueryParams) *gorm.DB {
	if value := queryParams.Filters["start_date"]; value != "" {
		query = query.Where("categories.created_at >= ?", value)
	}
	if value := queryParams.Filters["end_date"]; value != "" {
		query = query.Where("categories.created_at <= ?", value)
	}
	for column, value := range utils.SafeFilters(queryParams.Filters, categoryFilterColumns) {
		if value != "" {
			query = query.Where(column+" ILIKE ?", "%"+value+"%")
		}
	}
	return query
}

func applyCategorySort(query *gorm.DB, queryParams utils.QueryParams) *gorm.DB {
	order := "desc"
	if strings.ToLower(queryParams.Order) == "asc" {
		order = "asc"
	}
	sortColumn := utils.SafeSort(queryParams.SortBy, categorySortColumns, "categories.created_at")
	return query.Order(sortColumn + " " + order)
}

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

func (r *CategoryRepository) GetByUserID(userID uint, queryParams utils.QueryParams) (*[]models.Category, int64, int64, error) {
	var categories []models.Category
	var total int64

	query := r.db.Model(&models.Category{}).Where("categories.user_id = ?", userID)

	// APPLY FILTERS
	query = applyCategoryFilters(query, queryParams)

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
	query = applyCategorySort(query, queryParams)

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	if err := query.Limit(queryParams.Limit).Offset(offset).Find(&categories).Error; err != nil {
		return nil, 0, 0, err
	}
	return &categories, total, totalPages, nil
}

func (r *CategoryRepository) GetByUserIDWithTotal(userID uint, queryParams utils.QueryParams) (*[]models.CategoryWithTotal, int64, int64, error) {
	var categories []models.CategoryWithTotal
	var total int64

	query := r.db.Model(&models.Category{}).Where("categories.user_id = ?", userID)

	// APPLY FILTERS
	query = applyCategoryFilters(query, queryParams)

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
	query = applyCategorySort(query, queryParams)

	expenseStartDate := queryParams.ExpenseStartDate
	expenseEndDate := queryParams.ExpenseEndDate

	if expenseStartDate == "" || expenseEndDate == "" {
		// DEFAULT EXPENSE DATE RANGE
		now := time.Now()
		firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

		expenseStartDate = firstDayOfMonth.Format("2006-01-02")
		expenseEndDate = lastDayOfMonth.Format("2006-01-02")
	}

	joinCondition := "LEFT JOIN expenses ON expenses.category_id = categories.id AND expenses.user_id = ? AND expenses.deleted_at IS NULL AND expenses.created_at BETWEEN ? AND ?"
	query = query.
		Select("categories.*, COALESCE(SUM(expenses.amount), 0) AS total_expense").
		Joins(joinCondition, userID, expenseStartDate+" 00:00:00", expenseEndDate+" 23:59:59").
		Group("categories.id")

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
