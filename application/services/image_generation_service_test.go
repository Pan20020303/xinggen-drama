package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newImageGenerationServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	return db
}

func TestListImageGenerations_FiltersByCharacterID(t *testing.T) {
	db := newImageGenerationServiceTestDB(t)
	svc := &ImageGenerationService{db: db}

	characterID := uint(1001)
	otherCharacterID := uint(1002)
	userID := uint(7)
	dramaID := uint(9)
	completed := models.ImageStatusCompleted

	images := []models.ImageGeneration{
		{
			UserID:      userID,
			DramaID:     dramaID,
			CharacterID: &characterID,
			ImageType:   string(models.ImageTypeCharacter),
			Provider:    "openai",
			Prompt:      "角色A-1",
			Model:       "seedream-4.5",
			Size:        "1024x1024",
			Quality:     "standard",
			Status:      completed,
		},
		{
			UserID:      userID,
			DramaID:     dramaID,
			CharacterID: &otherCharacterID,
			ImageType:   string(models.ImageTypeCharacter),
			Provider:    "openai",
			Prompt:      "角色B-1",
			Model:       "seedream-4.5",
			Size:        "1024x1024",
			Quality:     "standard",
			Status:      completed,
		},
	}

	for _, image := range images {
		if err := db.Create(&image).Error; err != nil {
			t.Fatalf("failed to seed image generation: %v", err)
		}
	}

	result, total, err := svc.ListImageGenerations(userID, &dramaID, nil, &characterID, nil, "", string(completed), 1, 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result))
	}
	if result[0].CharacterID == nil || *result[0].CharacterID != characterID {
		t.Fatalf("expected character_id=%d, got %#v", characterID, result[0].CharacterID)
	}
}
