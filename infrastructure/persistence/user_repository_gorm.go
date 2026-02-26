package persistence

import (
	"github.com/drama-generator/backend/domain/models"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.
		Select(
			"id, email, password_hash, "+
				"COALESCE(role, ?) AS role, "+
				"COALESCE(status, ?) AS status, "+
				"COALESCE(credits, 0) AS credits",
			string(models.RoleUser), string(models.UserStatusActive),
		).
		Where("email = ?", email).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.
		Select(
			"id, email, password_hash, "+
				"COALESCE(role, ?) AS role, "+
				"COALESCE(status, ?) AS status, "+
				"COALESCE(credits, 0) AS credits",
			string(models.RoleUser), string(models.UserStatusActive),
		).
		First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *GormUserRepository) UpdatePassword(userID uint, hash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

func (r *GormUserRepository) CreateWithInitialCredits(user *models.User, initialCredits int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		desc := "initial credits for new user"
		txn := models.CreditTransaction{
			UserID:      user.ID,
			Amount:      initialCredits,
			Type:        models.CreditTxnRecharge,
			Description: &desc,
		}
		return tx.Create(&txn).Error
	})
}
