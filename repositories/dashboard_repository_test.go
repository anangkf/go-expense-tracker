package repositories

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OPENS A TRANSACTION ON TEST_DATABASE_DSN AND ROLLS IT BACK AT THE END, SO NO TEST DATA PERSISTS
func newTestTx(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is empty, skipping repository integration test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// AUTO MIGRATE ONLY ADDS MISSING TABLES/COLUMNS, NEVER DROPS
	if err := db.AutoMigrate(
		&models.User{},
		&models.BudgetPlan{},
		&models.BucketType{},
		&models.BudgetBucket{},
		&models.Category{},
		&models.Expense{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin tx: %v", tx.Error)
	}
	t.Cleanup(func() {
		tx.Rollback()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return tx
}

type dashboardFixture struct {
	tx             *gorm.DB
	user           models.User
	planID         uint
	incomeDefault  models.Category // DEFAULT CATEGORY (user_id NULL)
	expenseDefault models.Category
	rangeStart     time.Time
	rangeEnd       time.Time
}

func newDashboardFixture(t *testing.T) *dashboardFixture {
	t.Helper()
	tx := newTestTx(t)

	// BUCKET TYPE 1 MAY ALREADY EXIST (SEED); ROLLED BACK OTHERWISE
	bt := models.BucketType{ID: 1, Name: "Needs"}
	if err := tx.Where("id = ?", 1).FirstOrCreate(&bt).Error; err != nil {
		t.Fatalf("bucket type: %v", err)
	}

	user := models.User{
		Email:    fmt.Sprintf("repo-test-%d@example.test", time.Now().UnixNano()),
		Name:     "Repo Test",
		Password: "x",
	}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}

	bucketID := uint(1)
	incomeCat := models.Category{Name: "Test Salary", Type: "income", IsDefault: true, IconName: "i"}
	if err := tx.Create(&incomeCat).Error; err != nil {
		t.Fatalf("income category: %v", err)
	}
	expenseCat := models.Category{Name: "Test Food", Type: "expense", IsDefault: true, BucketTypeID: &bucketID, IconName: "i"}
	if err := tx.Create(&expenseCat).Error; err != nil {
		t.Fatalf("expense category: %v", err)
	}

	plan := models.BudgetPlan{UserID: &user.ID, Name: "Test Plan"}
	if err := tx.Create(&plan).Error; err != nil {
		t.Fatalf("plan: %v", err)
	}
	bucket := models.BudgetBucket{BudgetPlanID: plan.ID, BucketTypeID: 1, Name: "Needs", Percentage: 100}
	if err := tx.Create(&bucket).Error; err != nil {
		t.Fatalf("bucket: %v", err)
	}

	loc := time.Now().Location()
	return &dashboardFixture{
		tx:             tx,
		user:           user,
		planID:         plan.ID,
		incomeDefault:  incomeCat,
		expenseDefault: expenseCat,
		rangeStart:     time.Date(2020, 1, 1, 0, 0, 0, 0, loc),
		rangeEnd:       time.Date(2020, 1, 31, 0, 0, 0, 0, loc),
	}
}

func (f *dashboardFixture) addExpense(t *testing.T, cat models.Category, amount float64, at time.Time) *models.Expense {
	t.Helper()
	e := &models.Expense{Name: "t", Amount: amount, UserID: f.user.ID, CategoryID: cat.ID, CreatedAt: at}
	if cat.Type == "expense" {
		e.BucketTypeID = cat.BucketTypeID
	}
	if err := f.tx.Omit("Category", "BucketType").Create(e).Error; err != nil {
		t.Fatalf("create expense: %v", err)
	}
	return e
}

func (f *dashboardFixture) params() utils.QueryParams {
	return utils.QueryParams{
		ExpenseStartDate: f.rangeStart.Format("2006-01-02"),
		ExpenseEndDate:   f.rangeEnd.Format("2006-01-02"),
	}
}

func TestDashboardGetOverview_CountsDefaultCategoryIncome(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewDashboardRepository(f.tx)

	f.addExpense(t, f.incomeDefault, 1000, f.rangeStart.Add(10*time.Hour))
	f.addExpense(t, f.expenseDefault, 300, f.rangeStart.Add(11*time.Hour))

	o, err := repo.GetOverview(f.user.ID, f.rangeStart, f.rangeEnd)
	if err != nil {
		t.Fatal(err)
	}
	if o.TotalSpent != 300 || o.LeftInBudget != 700 {
		t.Fatalf("got spent=%v left=%v, want 300/700", o.TotalSpent, o.LeftInBudget)
	}
}

func TestDashboardGetOverview_IncludesLastDayLateEntry(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewDashboardRepository(f.tx)

	f.addExpense(t, f.incomeDefault, 1000, f.rangeEnd.Add(23*time.Hour+30*time.Minute))
	f.addExpense(t, f.expenseDefault, 200, f.rangeEnd.Add(23*time.Hour+30*time.Minute))
	// FIRST SECOND OF NEXT DAY MUST BE EXCLUDED
	f.addExpense(t, f.expenseDefault, 50, f.rangeEnd.AddDate(0, 0, 1))

	o, err := repo.GetOverview(f.user.ID, f.rangeStart, f.rangeEnd)
	if err != nil {
		t.Fatal(err)
	}
	if o.TotalSpent != 200 || o.LeftInBudget != 800 {
		t.Fatalf("got spent=%v left=%v, want 200/800", o.TotalSpent, o.LeftInBudget)
	}

	daily, err := repo.GetDailyExpenses(f.user.ID, f.rangeStart, f.rangeEnd)
	if err != nil {
		t.Fatal(err)
	}
	if len(daily) != 1 || daily[0].TotalAmount != 200 {
		t.Fatalf("unexpected daily expenses: %+v", daily)
	}
}

func TestDashboardGetOverview_ExcludesSoftDeleted(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewDashboardRepository(f.tx)

	f.addExpense(t, f.expenseDefault, 100, f.rangeStart.Add(time.Hour))
	del := f.addExpense(t, f.expenseDefault, 400, f.rangeStart.Add(2*time.Hour))
	delIncome := f.addExpense(t, f.incomeDefault, 900, f.rangeStart.Add(3*time.Hour))
	f.addExpense(t, f.incomeDefault, 500, f.rangeStart.Add(4*time.Hour))
	if err := f.tx.Delete(del).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.tx.Delete(delIncome).Error; err != nil {
		t.Fatal(err)
	}

	o, err := repo.GetOverview(f.user.ID, f.rangeStart, f.rangeEnd)
	if err != nil {
		t.Fatal(err)
	}
	if o.TotalSpent != 100 || o.LeftInBudget != 400 {
		t.Fatalf("got spent=%v left=%v, want 100/400", o.TotalSpent, o.LeftInBudget)
	}

	daily, err := repo.GetDailyExpenses(f.user.ID, f.rangeStart, f.rangeEnd)
	if err != nil {
		t.Fatal(err)
	}
	if len(daily) != 1 || daily[0].TotalAmount != 100 {
		t.Fatalf("unexpected daily expenses: %+v", daily)
	}
}

func TestExpenseGetTotalIncomeAndExpenses(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewExpenseRepository(f.tx)

	f.addExpense(t, f.incomeDefault, 1000, f.rangeEnd.Add(23*time.Hour+30*time.Minute)) // LAST DAY
	f.addExpense(t, f.expenseDefault, 200, f.rangeEnd.Add(23*time.Hour+30*time.Minute)) // LAST DAY
	f.addExpense(t, f.expenseDefault, 50, f.rangeEnd.AddDate(0, 0, 1))                  // OUT OF RANGE
	del := f.addExpense(t, f.expenseDefault, 400, f.rangeStart.Add(time.Hour))
	delIncome := f.addExpense(t, f.incomeDefault, 700, f.rangeStart.Add(time.Hour))
	if err := f.tx.Delete(del).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.tx.Delete(delIncome).Error; err != nil {
		t.Fatal(err)
	}

	income, spent, err := repo.GetTotalIncomeAndExpenses(f.user.ID, f.params())
	if err != nil {
		t.Fatal(err)
	}
	if income != 1000 || spent != 200 {
		t.Fatalf("got income=%v spent=%v, want 1000/200", income, spent)
	}
}

func TestBudgetPlanGetTotalSpendingByBucket(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewBudgetPlanRepository(f.tx)

	f.addExpense(t, f.expenseDefault, 200, f.rangeEnd.Add(23*time.Hour+30*time.Minute)) // LAST DAY
	f.addExpense(t, f.expenseDefault, 50, f.rangeEnd.AddDate(0, 0, 1))                  // OUT OF RANGE
	del := f.addExpense(t, f.expenseDefault, 400, f.rangeStart.Add(time.Hour))
	if err := f.tx.Delete(del).Error; err != nil {
		t.Fatal(err)
	}

	res, err := repo.GetTotalSpendingByBucket(f.planID, f.user.ID, f.params())
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].BucketName != "Needs" || res[0].TotalSpending != 200 {
		t.Fatalf("unexpected result: %+v", res)
	}
}
