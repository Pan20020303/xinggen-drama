package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newAITestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file:ai_service_platform_test?mode=memory&cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.AIServiceConfig{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedAIConfig(t *testing.T, db *gorm.DB, userID uint, serviceType, provider, name, model string, priority int) models.AIServiceConfig {
	t.Helper()

	cfg := models.AIServiceConfig{
		UserID:      userID,
		ServiceType: serviceType,
		Provider:    provider,
		Name:        name,
		BaseURL:     "https://api.example.com",
		APIKey:      "secret",
		Model:       models.ModelField{model},
		Priority:    priority,
		IsActive:    true,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}
	return cfg
}

func TestPlatformAIConfig_GetDefaultConfig_IgnoresUserOwnedConfigs(t *testing.T) {
	db := newAITestDB(t)
	svc := NewAIService(db, logger.NewLogger(true))

	// User-owned config has higher priority but must be ignored.
	_ = seedAIConfig(t, db, 123, "text", "openai", "user-high", "gpt-user", 100)
	platform := seedAIConfig(t, db, 0, "text", "openai", "platform-low", "gpt-platform", 1)

	got, err := svc.GetDefaultConfig("text")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.UserID != 0 {
		t.Fatalf("expected platform config user_id=0, got %d", got.UserID)
	}
	if got.ID != platform.ID {
		t.Fatalf("expected platform config id=%d, got %d", platform.ID, got.ID)
	}
}

func TestPlatformAIConfig_GetConfigForModel_IgnoresUserOwnedConfigs(t *testing.T) {
	db := newAITestDB(t)
	svc := NewAIService(db, logger.NewLogger(true))

	// Same model exists in both, but user-owned must be ignored.
	_ = seedAIConfig(t, db, 456, "text", "openai", "user-high", "gpt-1", 999)
	platform := seedAIConfig(t, db, 0, "text", "openai", "platform", "gpt-1", 1)

	got, err := svc.GetConfigForModel("text", "gpt-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.UserID != 0 {
		t.Fatalf("expected platform config user_id=0, got %d", got.UserID)
	}
	if got.ID != platform.ID {
		t.Fatalf("expected platform config id=%d, got %d", platform.ID, got.ID)
	}
}
