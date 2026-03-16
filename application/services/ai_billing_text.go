package services

import (
	"fmt"

	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/usage"
)

// reserveTextClient reserves credits for a single text-model call and returns a client for the resolved model.
// Callers should refund (billing.RefundAI(refID)) if later logic fails and the action is considered failed.
func reserveTextClient(aiService *AIService, billing *BillingService, userID uint, modelHint string, detail string) (ai.AIClient, string, string, error) {
	cfg, actualModel, err := aiService.GetBillingConfig("text", modelHint, userID)
	if err != nil {
		return nil, "", "", err
	}

	refID, err := billing.ReserveAI(userID, "text", actualModel, cfg.CreditCost, detail)
	if err != nil {
		return nil, "", "", err
	}

	client, err := aiService.GetAIClientForModelWithUser("text", actualModel, userID)
	if err != nil {
		if refID != "" {
			_ = billing.RefundAI(refID)
		}
		return nil, "", "", fmt.Errorf("failed to get AI client: %w", err)
	}

	return client, actualModel, refID, nil
}

func recordTextUsage(billing *BillingService, refID string, client ai.AIClient) {
	if billing == nil || refID == "" || client == nil {
		return
	}
	_ = billing.RecordAIUsage(refID, client.GetLastUsage())
}

func hasTokenUsage(tokenUsage usage.TokenUsage) bool {
	return tokenUsage.PromptTokens > 0 || tokenUsage.CompletionTokens > 0 || tokenUsage.TotalTokens > 0
}
