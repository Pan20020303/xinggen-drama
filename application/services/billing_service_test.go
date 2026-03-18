package services

import (
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

func TestBillingService_ListTransactionsFiltersCurrentUser(t *testing.T) {
	db := newAdminServiceTestDB(t)
	log := logger.NewLogger(true)
	svc := NewBillingService(db, &config.Config{}, log)

	userA := seedAdminServiceUser(t, db, "billing-a@example.com", models.RoleUser, models.UserStatusActive, 100)
	userB := seedAdminServiceUser(t, db, "billing-b@example.com", models.RoleUser, models.UserStatusActive, 100)

	descA := "image generate"
	descB := "manual recharge"
	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)

	if err := db.Create(&models.CreditTransaction{
		UserID:      userA.ID,
		Amount:      -5,
		Type:        models.CreditTxnAIImage,
		Description: &descA,
		CreatedAt:   older,
	}).Error; err != nil {
		t.Fatalf("failed to seed user A transaction: %v", err)
	}

	if err := db.Create(&models.CreditTransaction{
		UserID:      userB.ID,
		Amount:      20,
		Type:        models.CreditTxnRecharge,
		Description: &descB,
		CreatedAt:   newer,
	}).Error; err != nil {
		t.Fatalf("failed to seed user B transaction: %v", err)
	}

	descA2 := "video generate"
	if err := db.Create(&models.CreditTransaction{
		UserID:      userA.ID,
		Amount:      -15,
		Type:        models.CreditTxnAIVideo,
		Description: &descA2,
		CreatedAt:   newer.Add(10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("failed to seed second user A transaction: %v", err)
	}

	items, total, err := svc.ListTransactions(userA.ID, 1, 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].UserID != userA.ID || items[1].UserID != userA.ID {
		t.Fatalf("expected only current user's transactions, got %+v", items)
	}

	if items[0].Type != models.CreditTxnAIVideo {
		t.Fatalf("expected latest transaction first, got %s", items[0].Type)
	}
}
