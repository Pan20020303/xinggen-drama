package services

import (
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newAssetServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.Asset{}, &models.ImageGeneration{}, &models.VideoGeneration{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	return db
}

func TestAssetService_ListAssetsFiltersByUser(t *testing.T) {
	db := newAssetServiceTestDB(t)
	svc := NewAssetService(db, logger.NewLogger(true))

	ownAsset := models.Asset{
		UserID: 1,
		Name:   "own-image",
		Type:   models.AssetTypeImage,
		URL:    "https://example.com/a.png",
	}
	otherAsset := models.Asset{
		UserID: 2,
		Name:   "other-image",
		Type:   models.AssetTypeImage,
		URL:    "https://example.com/b.png",
	}

	if err := db.Create(&ownAsset).Error; err != nil {
		t.Fatalf("failed to seed own asset: %v", err)
	}
	if err := db.Create(&otherAsset).Error; err != nil {
		t.Fatalf("failed to seed other asset: %v", err)
	}

	items, total, err := svc.ListAssets(1, &ListAssetsRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 || items[0].UserID != 1 {
		t.Fatalf("expected only user 1 assets, got %#v", items)
	}
}

func TestAssetService_ImportFromImageGenAllowsToolImagesWithoutDrama(t *testing.T) {
	db := newAssetServiceTestDB(t)
	svc := NewAssetService(db, logger.NewLogger(true))

	imageURL := "https://example.com/tool-image.png"
	imageGen := models.ImageGeneration{
		UserID:      8,
		DramaID:     nil,
		ImageType:   "toolbox",
		Provider:    "openai",
		Prompt:      "tool image",
		Model:       "seedream",
		Size:        "1024x1024",
		Quality:     "standard",
		Status:      models.ImageStatusCompleted,
		ImageURL:    &imageURL,
		CompletedAt: ptrTime(time.Now()),
	}

	if err := db.Create(&imageGen).Error; err != nil {
		t.Fatalf("failed to seed image gen: %v", err)
	}

	asset, err := svc.ImportFromImageGen(8, imageGen.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if asset.UserID != 8 {
		t.Fatalf("expected asset user_id=8, got %d", asset.UserID)
	}
	if asset.DramaID != nil {
		t.Fatalf("expected nil drama_id for toolbox image, got %#v", asset.DramaID)
	}
	if asset.Category == nil || *asset.Category != "图片素材" {
		t.Fatalf("expected 图片素材 category, got %#v", asset.Category)
	}

	assetAgain, err := svc.ImportFromImageGen(8, imageGen.ID)
	if err != nil {
		t.Fatalf("expected idempotent import, got %v", err)
	}
	if assetAgain.ID != asset.ID {
		t.Fatalf("expected existing asset id %d, got %d", asset.ID, assetAgain.ID)
	}
}

func TestAssetService_UpdateAssetCanToggleFavorite(t *testing.T) {
	db := newAssetServiceTestDB(t)
	svc := NewAssetService(db, logger.NewLogger(true))

	asset := models.Asset{
		UserID:     5,
		Name:       "favorite-target",
		Type:       models.AssetTypeImage,
		URL:        "https://example.com/favorite.png",
		IsFavorite: false,
	}

	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("failed to seed asset: %v", err)
	}

	next := true
	updated, err := svc.UpdateAsset(5, asset.ID, &UpdateAssetRequest{IsFavorite: &next})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !updated.IsFavorite {
		t.Fatalf("expected asset to be marked favorite")
	}
}

func ptrTime(v time.Time) *time.Time {
	return &v
}
