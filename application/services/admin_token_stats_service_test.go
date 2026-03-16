package services

import (
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/usage"
	"github.com/stretchr/testify/require"
)

func TestAdminTokenStatsService_AggregatesByModel(t *testing.T) {
	db := newAdminServiceTestDB(t)
	svc := NewAdminTokenStatsService(db, logger.NewLogger(true))

	modelA := "doubao-1.8"
	modelB := "doubao-4.5"
	serviceType := "text"
	now := time.Now()

	txns := []models.CreditTransaction{
		{
			UserID:           1,
			Amount:           -10,
			Type:             models.CreditTxnAIText,
			ServiceType:      &serviceType,
			Model:            &modelA,
			PromptTokens:     intPtr(120),
			CompletionTokens: intPtr(80),
			TotalTokens:      intPtr(200),
			CreatedAt:        now,
		},
		{
			UserID:           2,
			Amount:           -12,
			Type:             models.CreditTxnAIText,
			ServiceType:      &serviceType,
			Model:            &modelA,
			PromptTokens:     intPtr(60),
			CompletionTokens: intPtr(40),
			TotalTokens:      intPtr(100),
			CreatedAt:        now,
		},
		{
			UserID:           3,
			Amount:           -15,
			Type:             models.CreditTxnAIText,
			ServiceType:      &serviceType,
			Model:            &modelB,
			PromptTokens:     intPtr(50),
			CompletionTokens: intPtr(50),
			TotalTokens:      intPtr(100),
			CreatedAt:        now,
		},
	}
	require.NoError(t, db.Create(&txns).Error)

	stats, summary, err := svc.GetTokenStats(&serviceType, nil, nil)
	require.NoError(t, err)
	require.Len(t, stats, 2)
	require.Equal(t, 230, summary.PromptTokens)
	require.Equal(t, 170, summary.CompletionTokens)
	require.Equal(t, 400, summary.TotalTokens)
	require.Equal(t, 2, summary.ModelCount)
	require.Equal(t, "doubao-1.8", stats[0].Model)
	require.Equal(t, 300, stats[0].TotalTokens)
}

func TestBillingService_RecordAIUsageUpdatesMatchingReferenceOnly(t *testing.T) {
	db := newAdminServiceTestDB(t)
	billing := NewBillingService(db, &config.Config{}, logger.NewLogger(true))

	serviceType := "text"
	model := "doubao-1.8"
	refA := "ref-a"
	refB := "ref-b"
	desc := "call"

	txns := []models.CreditTransaction{
		{
			UserID:      1,
			Amount:      -10,
			Type:        models.CreditTxnAIText,
			ReferenceID: &refA,
			ServiceType: &serviceType,
			Model:       &model,
			Description: &desc,
		},
		{
			UserID:      1,
			Amount:      -10,
			Type:        models.CreditTxnAIText,
			ReferenceID: &refB,
			ServiceType: &serviceType,
			Model:       &model,
			Description: &desc,
		},
	}
	require.NoError(t, db.Create(&txns).Error)

	err := billing.RecordAIUsage(refA, usage.TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 40,
		TotalTokens:      140,
	})
	require.NoError(t, err)

	var updatedA models.CreditTransaction
	require.NoError(t, db.First(&updatedA, txns[0].ID).Error)
	require.NotNil(t, updatedA.TotalTokens)
	require.Equal(t, 140, *updatedA.TotalTokens)

	var untouchedB models.CreditTransaction
	require.NoError(t, db.First(&untouchedB, txns[1].ID).Error)
	require.Nil(t, untouchedB.TotalTokens)
}

func intPtr(v int) *int {
	return &v
}
