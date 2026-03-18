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

	result, total, err := svc.ListImageGenerations(userID, &dramaID, nil, &characterID, nil, "", "", string(completed), 1, 20)
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

func TestListImageGenerations_FiltersByImageType(t *testing.T) {
	db := newImageGenerationServiceTestDB(t)
	svc := &ImageGenerationService{db: db}

	userID := uint(11)
	dramaID := uint(20)

	items := []models.ImageGeneration{
		{
			UserID:    userID,
			DramaID:   dramaID,
			ImageType: "toolbox",
			Provider:  "openai",
			Prompt:    "toolbox image",
			Model:     "seedream",
			Size:      "1024x1024",
			Quality:   "standard",
			Status:    models.ImageStatusCompleted,
		},
		{
			UserID:    userID,
			DramaID:   dramaID,
			ImageType: string(models.ImageTypeStoryboard),
			Provider:  "openai",
			Prompt:    "storyboard image",
			Model:     "seedream",
			Size:      "1024x1024",
			Quality:   "standard",
			Status:    models.ImageStatusCompleted,
		},
	}

	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to seed image generation: %v", err)
		}
	}

	result, total, err := svc.ListImageGenerations(userID, nil, nil, nil, nil, "", "toolbox", string(models.ImageStatusCompleted), 1, 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(result) != 1 || result[0].ImageType != "toolbox" {
		t.Fatalf("expected only toolbox image, got %#v", result)
	}
}

func TestNormalizeDimensionsForModel_ExpandsSeedream45TooSmallRequest(t *testing.T) {
	req := &GenerateImageRequest{
		Model:  "doubao-seedream-4-5-251128",
		Size:   "1920x1080",
		Width:  intPtr(1920),
		Height: intPtr(1080),
	}

	normalizeDimensionsForModel(req)

	if req.Width == nil || req.Height == nil {
		t.Fatalf("expected normalized dimensions to be set")
	}
	if *req.Width != 2560 || *req.Height != 1440 {
		t.Fatalf("expected 2560x1440, got %dx%d", *req.Width, *req.Height)
	}
	if req.Size != "2560x1440" {
		t.Fatalf("expected size 2560x1440, got %s", req.Size)
	}
}

func TestNormalizeDimensionsForModel_LeavesLargeEnoughRequestUnchanged(t *testing.T) {
	req := &GenerateImageRequest{
		Model:  "doubao-seedream-4-5-251128",
		Size:   "2560x1440",
		Width:  intPtr(2560),
		Height: intPtr(1440),
	}

	normalizeDimensionsForModel(req)

	if *req.Width != 2560 || *req.Height != 1440 {
		t.Fatalf("expected dimensions unchanged, got %dx%d", *req.Width, *req.Height)
	}
	if req.Size != "2560x1440" {
		t.Fatalf("expected size unchanged, got %s", req.Size)
	}
}

func intPtr(v int) *int {
	return &v
}
