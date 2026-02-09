package services

import (
	"errors"
	"fmt"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminBillingService struct {
	db    *gorm.DB
	log   *logger.Logger
	audit *AdminAuditService
}

func NewAdminBillingService(db *gorm.DB, log *logger.Logger, audit *AdminAuditService) *AdminBillingService {
	return &AdminBillingService{
		db:    db,
		log:   log,
		audit: audit,
	}
}

func (s *AdminBillingService) RechargeUser(adminID, userID uint, amount int, note, ip, userAgent string) (*models.User, *models.CreditTransaction, error) {
	if amount <= 0 {
		return nil, nil, errors.New("recharge amount must be positive")
	}

	var updated models.User
	var createdTxn models.CreditTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		before := map[string]interface{}{"credits": user.Credits}
		newCredits := user.Credits + amount
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("credits", newCredits).Error; err != nil {
			return err
		}

		desc := note
		if desc == "" {
			desc = "admin recharge"
		}
		txn := models.CreditTransaction{
			UserID:      userID,
			Amount:      amount,
			Type:        models.CreditTxnRecharge,
			Description: &desc,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}

		user.Credits = newCredits
		after := map[string]interface{}{"credits": user.Credits}
		if err := s.audit.WriteWithTx(
			tx,
			adminID,
			"billing.recharge",
			"user",
			fmt.Sprintf("%d", userID),
			before,
			after,
			AdminActorMeta{IP: ip, UserAgent: userAgent},
		); err != nil {
			return err
		}

		updated = user
		createdTxn = txn
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	s.log.Infow("admin recharged user", "admin_id", adminID, "user_id", userID, "amount", amount)
	return &updated, &createdTxn, nil
}

func (s *AdminBillingService) ListCreditTransactions(userID *uint, page, pageSize int) ([]models.CreditTransaction, int64, error) {
	page, pageSize = normalizePagination(page, pageSize)

	query := s.db.Model(&models.CreditTransaction{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txns []models.CreditTransaction
	q := s.db.Model(&models.CreditTransaction{}).Order("id DESC")
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&txns).Error; err != nil {
		return nil, 0, err
	}
	return txns, total, nil
}
