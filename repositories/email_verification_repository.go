package repositories

import (
	"crypto/subtle"
	"errors"
	"time"

	"go-expense-tracker-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrResendCooldown  = errors.New("resend cooldown active")
	ErrResendLimit     = errors.New("daily resend limit reached")
	ErrOTPInvalid      = errors.New("invalid verification code")
	ErrOTPExpired      = errors.New("verification code expired")
	ErrOTPTooManyTries = errors.New("too many attempts")
	ErrOTPNotFound     = errors.New("no pending verification code")
)

const resendWindow = 24 * time.Hour

type EmailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

// STORE A NEW CODE FOR THE USER (REPLACING THE OLD ONE) AFTER ENFORCING COOLDOWN AND DAILY LIMIT
func (r *EmailVerificationRepository) Issue(userID uint, codeHash string, now time.Time, expiresIn, cooldown time.Duration, maxPerWindow int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var ev models.EmailVerification
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&ev).Error

		// FIRST CODE FOR THIS USER
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&models.EmailVerification{
				UserID:          userID,
				CodeHash:        codeHash,
				ExpiresAt:       now.Add(expiresIn),
				LastSentAt:      now,
				SendCount:       1,
				SendWindowStart: now,
			}).Error
		}
		if err != nil {
			return err
		}

		if now.Sub(ev.LastSentAt) < cooldown {
			return ErrResendCooldown
		}

		// START A NEW WINDOW WHEN THE OLD ONE HAS PASSED
		if now.Sub(ev.SendWindowStart) >= resendWindow {
			ev.SendWindowStart = now
			ev.SendCount = 0
		}
		if ev.SendCount >= maxPerWindow {
			return ErrResendLimit
		}

		return tx.Model(&ev).Updates(map[string]any{
			"code_hash":         codeHash,
			"expires_at":        now.Add(expiresIn),
			"attempts":          0,
			"last_sent_at":      now,
			"send_count":        ev.SendCount + 1,
			"send_window_start": ev.SendWindowStart,
		}).Error
	})
}

// CHECK THE CODE. ON SUCCESS THE USER IS MARKED VERIFIED AND THE CODE IS DELETED IN ONE TRANSACTION.
// WRONG GUESSES ARE COUNTED AND PERSISTED EVEN THOUGH AN ERROR IS RETURNED.
func (r *EmailVerificationRepository) Verify(userID uint, codeHash string, now time.Time, maxAttempts int) error {
	var result error

	txErr := r.db.Transaction(func(tx *gorm.DB) error {
		var ev models.EmailVerification
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&ev).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result = ErrOTPNotFound
			return nil
		}
		if err != nil {
			return err
		}

		if now.After(ev.ExpiresAt) {
			result = ErrOTPExpired
			return nil
		}
		if ev.Attempts >= maxAttempts {
			result = ErrOTPTooManyTries
			return nil
		}

		if subtle.ConstantTimeCompare([]byte(ev.CodeHash), []byte(codeHash)) != 1 {
			attempts := ev.Attempts + 1
			if err := tx.Model(&ev).Update("attempts", attempts).Error; err != nil {
				return err
			}
			if attempts >= maxAttempts {
				result = ErrOTPTooManyTries
			} else {
				result = ErrOTPInvalid
			}
			return nil
		}

		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("email_verified_at", now).Error; err != nil {
			return err
		}
		return tx.Delete(&ev).Error
	})
	if txErr != nil {
		return txErr
	}
	return result
}

// HARD DELETE UNVERIFIED USERS CREATED BEFORE cutoff AND WITHOUT A CODE SENT SINCE cutoff.
// VERIFIED USERS ARE NEVER TOUCHED. HARD DELETE FREES THE UNIQUE EMAIL FOR A NEW REGISTRATION.
func (r *EmailVerificationRepository) DeleteStaleUnverifiedUsers(cutoff time.Time) (int64, error) {
	var deleted int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		stale := tx.Unscoped().Model(&models.User{}).
			Select("users.id").
			Joins("LEFT JOIN email_verifications ev ON ev.user_id = users.id").
			Where("users.email_verified_at IS NULL AND users.created_at < ? AND (ev.id IS NULL OR ev.last_sent_at < ?)", cutoff, cutoff)

		var ids []uint
		if err := stale.Scan(&ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := tx.Where("user_id IN ?", ids).Delete(&models.EmailVerification{}).Error; err != nil {
			return err
		}
		// RE-CHECK email_verified_at SO A USER VERIFIED MEANWHILE IS NOT DELETED
		res := tx.Unscoped().Where("id IN ? AND email_verified_at IS NULL", ids).Delete(&models.User{})
		deleted = res.RowsAffected
		return res.Error
	})
	return deleted, err
}
