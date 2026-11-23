package repositories

import (
	"fmt"
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BUDGET PLAN SORT ALLOWLIST (query name -> SQL column)
var budgetPlanSortColumns = map[string]string{
	"id":          "budget_plans.id",
	"name":        "budget_plans.name",
	"description": "budget_plans.description",
	"is_template": "budget_plans.is_template",
	"created_at":  "budget_plans.created_at",
	"updated_at":  "budget_plans.updated_at",
}

// BUDGET PLAN FILTER ALLOWLIST (query name -> SQL expression, cast to text for ILIKE)
var budgetPlanFilterColumns = map[string]string{
	"id":          "budget_plans.id::text",
	"name":        "budget_plans.name",
	"description": "budget_plans.description",
	"created_at":  "budget_plans.created_at::text",
	"updated_at":  "budget_plans.updated_at::text",
}

type BudgetPlanRepository struct {
	db *gorm.DB
}

type TotalSpendingByBucket struct {
	BucketName    string
	TotalSpending float64
}

func NewBudgetPlanRepository(db *gorm.DB) *BudgetPlanRepository {
	return &BudgetPlanRepository{db}
}

func (r *BudgetPlanRepository) GetBudgetPlanTemplates() ([]models.BudgetPlan, error) {
	var templates []models.BudgetPlan

	err := r.db.Preload("Buckets.BucketType").Where("is_template = ?", true).Find(&templates).Error

	return templates, err
}

func (r *BudgetPlanRepository) GetBudgetPlansByUserID(userID uint, queryParams utils.QueryParams) (*[]models.BudgetPlan, int64, int64, error) {
	var budgetPlans []models.BudgetPlan
	var total int64

	query := r.db.Model(&models.BudgetPlan{}).Where("user_id = ?", userID)

	// APPLY FILTERS
	if value := queryParams.Filters["start_date"]; value != "" {
		query = query.Where("budget_plans.created_at >= ?", value)
	}
	if value := queryParams.Filters["end_date"]; value != "" {
		query = query.Where("budget_plans.created_at <= ?", value)
	}
	for column, value := range utils.SafeFilters(queryParams.Filters, budgetPlanFilterColumns) {
		if value != "" {
			query = query.Where(column+" ILIKE ?", "%"+value+"%")
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
	order := "desc"
	if strings.ToLower(queryParams.Order) == "asc" {
		order = "asc"
	}
	sortColumn := utils.SafeSort(queryParams.SortBy, budgetPlanSortColumns, "budget_plans.created_at")
	query = query.Order(sortColumn + " " + order)

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	query = query.Offset(offset).Limit(queryParams.Limit)

	if err := query.Preload("Buckets.BucketType").Find(&budgetPlans).Error; err != nil {
		return nil, 0, 0, err
	}

	return &budgetPlans, total, totalPages, nil
}

func (r *BudgetPlanRepository) CreateBudgetPlan(budgetPlan *models.BudgetPlan) error {
	return r.db.Preload("Buckets.BucketType").Create(budgetPlan).Error
}

func (r *BudgetPlanRepository) GetBudgetPlanByID(budgetPlanID uint, userID uint) (*models.BudgetPlan, error) {
	var budgetPlan models.BudgetPlan

	err := r.db.Preload("Buckets.BucketType").First(&budgetPlan, "id = ? AND user_id = ?", budgetPlanID, userID).Error

	return &budgetPlan, err
}

func (r *BudgetPlanRepository) UpdateBudgetPlanAndSyncBuckets(budgetPlan *models.BudgetPlan, toCreate, toUpdate []models.BudgetBucket, toDeleteIDs []uint) error {
	fmt.Printf("toCreate: %v, toUpdate: %v, toDeleteIDs: %v\n", toCreate, toUpdate, toDeleteIDs)
	return r.db.Transaction(func(tx *gorm.DB) error {
		// UPDATE BUDGET PLAN DETAILS
		if err := tx.Model(budgetPlan).Updates(map[string]interface{}{
			"Name":        budgetPlan.Name,
			"Description": budgetPlan.Description,
		}).Error; err != nil {
			return err
		}

		// DELETE BUDGET BUCKETS THAT ARE MARKED FOR DELETION
		if len(toDeleteIDs) > 0 {
			if err := tx.Where("id IN ?", toDeleteIDs).Delete(&models.BudgetBucket{}).Error; err != nil {
				return err
			}
		}

		// UPDATE EXISTING BUDGET BUCKETS
		if len(toUpdate) > 0 {
			for _, bucket := range toUpdate {
				if err := tx.Save(&bucket).Error; err != nil {
					return err
				}
			}
		}

		// CREATE NEW BUDGET BUCKETS
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *BudgetPlanRepository) DeleteBudgetPlan(budgetPlan *models.BudgetPlan) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// DELETE ALL RELATED BUDGET BUCKETS
		if err := tx.Where("budget_plan_id = ?", budgetPlan.ID).Delete(&models.BudgetBucket{}).Error; err != nil {
			return err
		}

		// DELETE BUDGET PLAN
		if err := tx.Delete(budgetPlan).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *BudgetPlanRepository) GetTotalSpendingByBucket(budgetPlanID uint, userID uint, queryParams utils.QueryParams) ([]TotalSpendingByBucket, error) {
	var result []TotalSpendingByBucket
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

	// PARSE DATE RANGE
	loc := time.Now().Location()
	startDate, err := time.ParseInLocation("2006-01-02", expenseStartDate, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_start_date: %w", err)
	}
	endDate, err := time.ParseInLocation("2006-01-02", expenseEndDate, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_end_date: %w", err)
	}

	err = r.db.Table("budget_buckets bb").
		Select("bb.name as bucket_name, COALESCE(SUM(e.amount), 0) as total_spending").
		Joins("LEFT JOIN expenses e ON e.bucket_type_id = bb.bucket_type_id AND e.user_id = ? AND e.deleted_at IS NULL AND e.created_at >= ? AND e.created_at < ?",
			userID, utils.StartOfDay(startDate), utils.EndOfDayExclusive(endDate)).
		Where("bb.budget_plan_id = ?", budgetPlanID).
		Group("bb.name").
		Scan(&result).Error

	return result, err
}
