package repositories

import (
	"errors"
	"testing"
	"time"

	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

func newVerificationFixture(t *testing.T) (*EmailVerificationRepository, *gorm.DB, models.User) {
	t.Helper()
	tx := newTestTx(t)
	if err := tx.AutoMigrate(&models.EmailVerification{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	user := models.User{Name: "Budi", Email: "budi-" + time.Now().Format("150405.000000") + "@test.local", Password: "x"}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return NewEmailVerificationRepository(tx), tx, user
}

func reload(t *testing.T, tx *gorm.DB, id uint) models.User {
	t.Helper()
	var u models.User
	if err := tx.First(&u, id).Error; err != nil {
		t.Fatal(err)
	}
	return u
}

func TestEmailVerification_VerifySuccess(t *testing.T) {
	repo, tx, user := newVerificationFixture(t)
	now := time.Now()

	if err := repo.Issue(user.ID, "hash", now, 10*time.Minute, time.Minute, 5); err != nil {
		t.Fatal(err)
	}
	if err := repo.Verify(user.ID, "hash", now.Add(time.Minute), 5); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if reload(t, tx, user.ID).EmailVerifiedAt == nil {
		t.Fatal("user should be verified")
	}
	// CODE IS SINGLE USE
	if err := repo.Verify(user.ID, "hash", now.Add(time.Minute), 5); !errors.Is(err, ErrOTPNotFound) {
		t.Fatalf("expected ErrOTPNotFound on reuse, got %v", err)
	}
}

func TestEmailVerification_Expired(t *testing.T) {
	repo, tx, user := newVerificationFixture(t)
	now := time.Now()
	_ = repo.Issue(user.ID, "hash", now, 10*time.Minute, time.Minute, 5)

	if err := repo.Verify(user.ID, "hash", now.Add(11*time.Minute), 5); !errors.Is(err, ErrOTPExpired) {
		t.Fatalf("expected ErrOTPExpired, got %v", err)
	}
	if reload(t, tx, user.ID).EmailVerifiedAt != nil {
		t.Fatal("user must stay unverified")
	}
}

func TestEmailVerification_MaxAttempts(t *testing.T) {
	repo, tx, user := newVerificationFixture(t)
	now := time.Now()
	_ = repo.Issue(user.ID, "hash", now, 10*time.Minute, time.Minute, 5)

	for i := 1; i <= 4; i++ {
		if err := repo.Verify(user.ID, "bad", now, 5); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("attempt %d: expected ErrOTPInvalid, got %v", i, err)
		}
	}
	if err := repo.Verify(user.ID, "bad", now, 5); !errors.Is(err, ErrOTPTooManyTries) {
		t.Fatalf("expected ErrOTPTooManyTries, got %v", err)
	}
	// EVEN THE CORRECT CODE IS REJECTED NOW
	if err := repo.Verify(user.ID, "hash", now, 5); !errors.Is(err, ErrOTPTooManyTries) {
		t.Fatalf("expected lockout, got %v", err)
	}
	if reload(t, tx, user.ID).EmailVerifiedAt != nil {
		t.Fatal("user must stay unverified")
	}
	// RESEND RESETS ATTEMPTS
	if err := repo.Issue(user.ID, "new", now.Add(2*time.Minute), 10*time.Minute, time.Minute, 5); err != nil {
		t.Fatal(err)
	}
	if err := repo.Verify(user.ID, "new", now.Add(2*time.Minute), 5); err != nil {
		t.Fatalf("expected success after resend, got %v", err)
	}
}

func TestEmailVerification_ResendLimits(t *testing.T) {
	repo, _, user := newVerificationFixture(t)
	now := time.Now()

	if err := repo.Issue(user.ID, "h1", now, 10*time.Minute, time.Minute, 3); err != nil {
		t.Fatal(err)
	}
	if err := repo.Issue(user.ID, "h2", now.Add(10*time.Second), 10*time.Minute, time.Minute, 3); !errors.Is(err, ErrResendCooldown) {
		t.Fatalf("expected cooldown, got %v", err)
	}
	if err := repo.Issue(user.ID, "h2", now.Add(2*time.Minute), 10*time.Minute, time.Minute, 3); err != nil {
		t.Fatal(err)
	}
	if err := repo.Issue(user.ID, "h3", now.Add(4*time.Minute), 10*time.Minute, time.Minute, 3); err != nil {
		t.Fatal(err)
	}
	if err := repo.Issue(user.ID, "h4", now.Add(6*time.Minute), 10*time.Minute, time.Minute, 3); !errors.Is(err, ErrResendLimit) {
		t.Fatalf("expected daily limit, got %v", err)
	}
	// NEW WINDOW AFTER 24H
	if err := repo.Issue(user.ID, "h5", now.Add(25*time.Hour), 10*time.Minute, time.Minute, 3); err != nil {
		t.Fatalf("expected new window to allow send, got %v", err)
	}
}

func TestEmailVerification_DeleteStaleUnverifiedUsers(t *testing.T) {
	repo, tx, stale := newVerificationFixture(t)
	now := time.Now()
	old := now.Add(-10 * 24 * time.Hour)
	cutoff := now.Add(-7 * 24 * time.Hour)

	mk := func(email string, verified bool, created time.Time) models.User {
		u := models.User{Name: "U", Email: email + "@test.local", Password: "x", CreatedAt: created}
		if verified {
			u.EmailVerifiedAt = &created
		}
		if err := tx.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		return u
	}
	tx.Model(&stale).Update("created_at", old)
	_ = repo.Issue(stale.ID, "h", old, time.Minute, time.Minute, 5)

	oldVerified := mk("oldverified", true, old)
	recent := mk("recent", false, now)
	resentRecently := mk("resent", false, old)
	_ = repo.Issue(resentRecently.ID, "h", now.Add(-time.Hour), time.Minute, time.Minute, 5)

	n, err := repo.DeleteStaleUnverifiedUsers(cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected at least the stale user deleted, got %d", n)
	}

	var count int64
	tx.Unscoped().Model(&models.User{}).Where("id = ?", stale.ID).Count(&count)
	if count != 0 {
		t.Fatal("stale unverified user should be hard deleted")
	}
	for name, id := range map[string]uint{"verified": oldVerified.ID, "recent": recent.ID, "recently resent": resentRecently.ID} {
		tx.Model(&models.User{}).Where("id = ?", id).Count(&count)
		if count != 1 {
			t.Fatalf("%s user must be kept", name)
		}
	}
}
