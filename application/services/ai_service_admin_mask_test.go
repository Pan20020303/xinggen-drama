package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
)

func TestAIConfigMask_PublicView_DoesNotExposeAPIKey(t *testing.T) {
	cfg := models.AIServiceConfig{
		ID:          1,
		UserID:      0,
		ServiceType: "text",
		Provider:    "openai",
		Name:        "platform",
		BaseURL:     "https://api.example.com",
		APIKey:      "super-secret",
		Model:       models.ModelField{"gpt-x"},
		IsActive:    true,
		Priority:    10,
	}

	view := ToAIServiceConfigView(cfg)
	if view.APIKey != "" {
		t.Fatalf("expected api_key to be empty in view, got %q", view.APIKey)
	}
	if !view.APIKeySet {
		t.Fatalf("expected api_key_set=true when stored api key exists")
	}
}

func TestAIConfigMask_PublicView_APIKeySetFalseWhenEmpty(t *testing.T) {
	cfg := models.AIServiceConfig{
		ID:          1,
		UserID:      0,
		ServiceType: "text",
		Provider:    "openai",
		Name:        "platform",
		BaseURL:     "https://api.example.com",
		APIKey:      "",
		Model:       models.ModelField{"gpt-x"},
		IsActive:    true,
		Priority:    10,
	}

	view := ToAIServiceConfigView(cfg)
	if view.APIKey != "" {
		t.Fatalf("expected api_key to be empty in view, got %q", view.APIKey)
	}
	if view.APIKeySet {
		t.Fatalf("expected api_key_set=false when stored api key is empty")
	}
}
