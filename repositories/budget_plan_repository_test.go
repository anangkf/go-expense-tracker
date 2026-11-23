package repositories

import (
	"testing"

	"go-expense-tracker-api/models"
)

func TestUpdateBudgetPlanAndSyncBucketsPersistsBucketTypeID(t *testing.T) {
	f := newDashboardFixture(t)
	repo := NewBudgetPlanRepository(f.tx)

	// ENSURE BUCKET TYPES 2 AND 3 EXIST (ROLLED BACK WITH THE TX)
	for id, name := range map[uint]string{2: "Wants", 3: "Savings"} {
		bt := models.BucketType{ID: id, Name: name}
		if err := f.tx.Where("id = ?", id).FirstOrCreate(&bt).Error; err != nil {
			t.Fatalf("bucket type %d: %v", id, err)
		}
	}

	var existing models.BudgetBucket
	if err := f.tx.First(&existing, "budget_plan_id = ?", f.planID).Error; err != nil {
		t.Fatalf("existing bucket: %v", err)
	}

	plan := &models.BudgetPlan{ID: f.planID, Name: "Test Plan"}
	toUpdate := []models.BudgetBucket{{ID: existing.ID, BudgetPlanID: f.planID, BucketTypeID: 2, Name: "Wants", Percentage: 60}}
	toCreate := []models.BudgetBucket{{BudgetPlanID: f.planID, BucketTypeID: 3, Name: "Savings", Percentage: 40}}

	if err := repo.UpdateBudgetPlanAndSyncBuckets(plan, toCreate, toUpdate, nil); err != nil {
		t.Fatalf("update: %v", err)
	}

	var updated models.BudgetBucket
	if err := f.tx.First(&updated, existing.ID).Error; err != nil {
		t.Fatalf("reload updated: %v", err)
	}
	if updated.BucketTypeID != 2 || updated.Name != "Wants" || updated.Percentage != 60 {
		t.Errorf("updated bucket = %+v, want bucket_type_id 2, name Wants, percentage 60", updated)
	}
	if updated.CreatedAt.IsZero() || updated.CreatedAt.Year() < 2000 {
		t.Errorf("created_at was clobbered: %v", updated.CreatedAt)
	}

	var created models.BudgetBucket
	if err := f.tx.First(&created, "budget_plan_id = ? AND name = ?", f.planID, "Savings").Error; err != nil {
		t.Fatalf("reload created: %v", err)
	}
	if created.BucketTypeID != 3 {
		t.Errorf("created bucket_type_id = %d, want 3", created.BucketTypeID)
	}
}
