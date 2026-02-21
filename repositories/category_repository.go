package repositories

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"
	"time"

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

func (r *CategoryRepository) GetByUserID(userID uint, queryParams utils.QueryParams, withTotal bool) (*[]models.Category, int64, int64, error) {
	var categories []models.Category
	var total int64

	query := r.db.Model(&models.Category{}).Where("categories.user_id = ?", userID)

	// APPLY FILTERS
	for key, value := range queryParams.Filters {
		if value != "" {
			if key == "start_date" {
				query = query.Where("categories.created_at >= ?", value)
				continue
			}
			if key == "end_date" {
				query = query.Where("categories.created_at <= ?", value)
				continue
			}
			query = query.Where("categories."+key+" ILIKE ?", "%"+value+"%")
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
		query = query.Order("categories." + queryParams.SortBy + " " + order)
	}

	if withTotal {
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
