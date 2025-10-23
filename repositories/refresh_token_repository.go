package repositories

import (
	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("token = ? AND is_revoked = ?", token, false).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *RefreshTokenRepository) GetByUserID(userID uint) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("user_id = ? AND is_revoked = ?", userID, false).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *RefreshTokenRepository) GetByJTI(jti string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("jti = ? AND is_revoked = ?", jti, false).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *RefreshTokenRepository) Revoke(token string) error {
	return r.db.Model(&models.RefreshToken{}).Where("token = ?", token).Update("is_revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeByJTI(jti string) (int64, error) {
	result := r.db.Model(&models.RefreshToken{}).Where("jti = ? AND is_revoked = ?", jti, false).Update("is_revoked", true)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// RESERVERD FOR FUTURE USE (IF NEEDED)
func (r *RefreshTokenRepository) RevokeAllByUserID(userID uint) (int64, error) {
	result := r.db.Model(&models.RefreshToken{}).Where("user_id = ? AND is_revoked = ?", userID, false).Update("is_revoked", true)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
