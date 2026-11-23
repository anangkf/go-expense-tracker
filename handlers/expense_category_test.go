package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OPENS A TRANSACTION ON TEST_DATABASE_DSN AND ROLLS IT BACK AT THE END, SO NO TEST DATA PERSISTS
func newHandlerTestTx(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is empty, skipping handler integration test")
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

type expenseCategoryFixture struct {
	tx          *gorm.DB
	router      *gin.Engine
	user        models.User
	needsCat    models.Category // expense, bucket 1
	wantsCat    models.Category // expense, bucket 2
	incomeCat   models.Category
	expenseRepo *repositories.ExpenseRepository
}

func newExpenseCategoryFixture(t *testing.T) *expenseCategoryFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tx := newHandlerTestTx(t)

	// BUCKET TYPES 1 AND 2 MAY ALREADY EXIST (SEED); ROLLED BACK OTHERWISE
	for id, name := range map[uint]string{1: "Needs", 2: "Wants"} {
		bt := models.BucketType{ID: id, Name: name}
		if err := tx.Where("id = ?", id).FirstOrCreate(&bt).Error; err != nil {
			t.Fatalf("bucket type %d: %v", id, err)
		}
	}

	user := models.User{
		Email:    fmt.Sprintf("handler-test-%d@example.test", time.Now().UnixNano()),
		Name:     "Handler Test",
		Password: "x",
	}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}

	needsID, wantsID := uint(1), uint(2)
	mk := func(name, typ string, bucket *uint) models.Category {
		c := models.Category{Name: name, Type: typ, UserID: &user.ID, BucketTypeID: bucket, IconName: "i"}
		if err := tx.Create(&c).Error; err != nil {
			t.Fatalf("category %s: %v", name, err)
		}
		return c
	}
	f := &expenseCategoryFixture{
		tx:        tx,
		user:      user,
		needsCat:  mk("Test Needs", "expense", &needsID),
		wantsCat:  mk("Test Wants", "expense", &wantsID),
		incomeCat: mk("Test Income", "income", nil),
	}

	categoryRepo := repositories.NewCategoryRepository(tx)
	f.expenseRepo = repositories.NewExpenseRepository(tx)
	userRepo := repositories.NewUserRepository(tx)
	expenseHandler := NewExpenseHandler(f.expenseRepo, userRepo, categoryRepo)
	categoryHandler := NewCategoryHandler(categoryRepo, userRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	r.PUT("/expenses/:id", expenseHandler.UpdateExpense)
	r.DELETE("/categories/:id", categoryHandler.DeleteCategory)
	f.router = r
	return f
}

func (f *expenseCategoryFixture) addExpense(t *testing.T, cat models.Category) *models.Expense {
	t.Helper()
	e := &models.Expense{Name: "t", Amount: 10, UserID: f.user.ID, CategoryID: cat.ID}
	if cat.Type == "expense" {
		e.BucketTypeID = cat.BucketTypeID
	}
	if err := f.tx.Omit("Category", "BucketType").Create(e).Error; err != nil {
		t.Fatalf("create expense: %v", err)
	}
	return e
}

func (f *expenseCategoryFixture) updateExpense(t *testing.T, id uint, categoryID uint) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(models.ExpenseRequest{Name: "updated", Amount: 25, CategoryID: categoryID})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/expenses/%d", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func (f *expenseCategoryFixture) deleteCategory(id uint) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/categories/%d", id), nil)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func (f *expenseCategoryFixture) reload(t *testing.T, id uint) models.Expense {
	t.Helper()
	var e models.Expense
	if err := f.tx.Omit("Category", "BucketType").First(&e, id).Error; err != nil {
		t.Fatalf("reload expense: %v", err)
	}
	return e
}

func TestUpdateExpenseCategoryExpenseToExpenseCopiesBucketType(t *testing.T) {
	f := newExpenseCategoryFixture(t)
	e := f.addExpense(t, f.needsCat)

	w := f.updateExpense(t, e.ID, f.wantsCat.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	got := f.reload(t, e.ID)
	if got.CategoryID != f.wantsCat.ID {
		t.Errorf("category_id = %d, want %d", got.CategoryID, f.wantsCat.ID)
	}
	if got.BucketTypeID == nil || *got.BucketTypeID != 2 {
		t.Errorf("bucket_type_id = %v, want 2", got.BucketTypeID)
	}
}

func TestUpdateExpenseCategoryExpenseToIncomeClearsBucketType(t *testing.T) {
	f := newExpenseCategoryFixture(t)
	e := f.addExpense(t, f.needsCat)

	w := f.updateExpense(t, e.ID, f.incomeCat.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	got := f.reload(t, e.ID)
	if got.CategoryID != f.incomeCat.ID {
		t.Errorf("category_id = %d, want %d", got.CategoryID, f.incomeCat.ID)
	}
	if got.BucketTypeID != nil {
		t.Errorf("bucket_type_id = %d, want NULL", *got.BucketTypeID)
	}
}

func TestDeleteCategoryUsedByExpenseReturnsConflict(t *testing.T) {
	f := newExpenseCategoryFixture(t)
	f.addExpense(t, f.needsCat)

	w := f.deleteCategory(f.needsCat.ID)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body = %s", w.Code, w.Body.String())
	}

	var count int64
	if err := f.tx.Model(&models.Category{}).Where("id = ?", f.needsCat.ID).Count(&count).Error; err != nil || count != 1 {
		t.Errorf("category should still exist, count = %d, err = %v", count, err)
	}
}

func TestDeleteCategoryUnusedSucceeds(t *testing.T) {
	f := newExpenseCategoryFixture(t)

	w := f.deleteCategory(f.wantsCat.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var count int64
	if err := f.tx.Model(&models.Category{}).Where("id = ?", f.wantsCat.ID).Count(&count).Error; err != nil || count != 0 {
		t.Errorf("category should be soft-deleted, count = %d, err = %v", count, err)
	}
}

func TestDeleteCategoryUsedOnlyBySoftDeletedExpenseSucceeds(t *testing.T) {
	f := newExpenseCategoryFixture(t)
	e := f.addExpense(t, f.needsCat)
	if err := f.tx.Delete(&models.Expense{}, e.ID).Error; err != nil {
		t.Fatalf("soft delete expense: %v", err)
	}

	w := f.deleteCategory(f.needsCat.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
