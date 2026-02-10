package services

import (
	"errors"
	"fmt"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
)

type BillingService struct {
	db                  *gorm.DB
	log                 *logger.Logger
	framePromptCredits  int
	imageGenerateCredits int
}

func NewBillingService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *BillingService {
	frameCost := cfg.Billing.FramePromptCredits
	if frameCost <= 0 {
		frameCost = 10
	}
	imageCost := cfg.Billing.ImageGenerationCredits
	if imageCost <= 0 {
		imageCost = 5
	}

	return &BillingService{
		db:                   db,
		log:                  log,
		framePromptCredits:   frameCost,
		imageGenerateCredits: imageCost,
	}
}

func (s *BillingService) ConsumeForFramePrompt(userID uint, detail string) error {
	return s.consume(userID, s.framePromptCredits, models.CreditTxnGenerateFrame, detail)
}

func (s *BillingService) ConsumeForImageGeneration(userID uint, detail string) error {
	return s.consume(userID, s.imageGenerateCredits, models.CreditTxnGenerateImage, detail)
}

func (s *BillingService) consume(userID uint, cost int, txnType models.CreditTransactionType, detail string) error {
	if cost <= 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		if user.Credits < cost {
			return ErrInsufficientCredits
		}

		newCredits := user.Credits - cost
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("credits", newCredits).Error; err != nil {
			return err
		}

		desc := detail
		amount := -cost
		txn := models.CreditTransaction{
			UserID:      userID,
			Amount:      amount,
			Type:        txnType,
			Description: &desc,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}

		s.log.Infow("credits consumed", "user_id", userID, "cost", cost, "type", txnType)
		return nil
	})
}

func creditTxnTypeForServiceType(serviceType string) (reserve models.CreditTransactionType, refund models.CreditTransactionType, err error) {
	switch serviceType {
	case "text":
		return models.CreditTxnAIText, models.CreditTxnAITextRefund, nil
	case "image":
		return models.CreditTxnAIImage, models.CreditTxnAIImageRefund, nil
	case "video":
		return models.CreditTxnAIVideo, models.CreditTxnAIVideoRefund, nil
	default:
		return "", "", fmt.Errorf("unknown service_type: %s", serviceType)
	}
}

// ReserveAI reserves credits for a single model call. On success, credits are deducted immediately.
// If the model call later fails, call RefundAI(referenceID) to revert the reservation.
func (s *BillingService) ReserveAI(userID uint, serviceType, model string, cost int, detail string) (string, error) {
	if cost <= 0 {
		return "", nil
	}
	reserveType, _, err := creditTxnTypeForServiceType(serviceType)
	if err != nil {
		return "", err
	}
	refID := uuid.New().String()

	return refID, s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Credits < cost {
			return ErrInsufficientCredits
		}

		newCredits := user.Credits - cost
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("credits", newCredits).Error; err != nil {
			return err
		}

		desc := detail
		st := serviceType
		m := model
		txn := models.CreditTransaction{
			UserID:      userID,
			Amount:      -cost,
			Type:        reserveType,
			ReferenceID: &refID,
			ServiceType: &st,
			Model:       &m,
			Description: &desc,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}
		return nil
	})
}

// RefundAI refunds a previous reservation. This is idempotent: if already refunded, it returns nil.
func (s *BillingService) RefundAI(referenceID string) error {
	if referenceID == "" {
		return nil
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// If refund exists, do nothing.
		var existing int64
		if err := tx.Model(&models.CreditTransaction{}).
			Where("reference_id = ? AND amount > 0", referenceID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}

		var reserved models.CreditTransaction
		if err := tx.Where("reference_id = ? AND amount < 0", referenceID).
			Order("id DESC").
			First(&reserved).Error; err != nil {
			// Nothing reserved, nothing to refund.
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		cost := -reserved.Amount
		serviceType := ""
		model := ""
		if reserved.ServiceType != nil {
			serviceType = *reserved.ServiceType
		}
		if reserved.Model != nil {
			model = *reserved.Model
		}
		_, refundType, err := creditTxnTypeForServiceType(serviceType)
		if err != nil {
			return err
		}

		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, reserved.UserID).Error; err != nil {
			return err
		}
		newCredits := user.Credits + cost
		if err := tx.Model(&models.User{}).Where("id = ?", reserved.UserID).Update("credits", newCredits).Error; err != nil {
			return err
		}

		desc := "refund: " + safeStr(reserved.Description)
		st := serviceType
		m := model
		ref := referenceID
		refundTxn := models.CreditTransaction{
			UserID:      reserved.UserID,
			Amount:      cost,
			Type:        refundType,
			ReferenceID: &ref,
			ServiceType: &st,
			Model:       &m,
			Description: &desc,
		}
		return tx.Create(&refundTxn).Error
	})
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *BillingService) GetCosts() (framePrompt int, imageGen int) {
	return s.framePromptCredits, s.imageGenerateCredits
}

func BillingDetail(resource string, id interface{}) string {
	return fmt.Sprintf("%s:%v", resource, id)
}
