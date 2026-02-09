package services

import (
	"errors"
	"fmt"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			return errors.New("insufficient credits")
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

func (s *BillingService) GetCosts() (framePrompt int, imageGen int) {
	return s.framePromptCredits, s.imageGenerateCredits
}

func BillingDetail(resource string, id interface{}) string {
	return fmt.Sprintf("%s:%v", resource, id)
}
