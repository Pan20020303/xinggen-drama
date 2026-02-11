package database

import (
	"path/filepath"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newDataFixTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "datafix.db")
	db, err := NewDatabase(config.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	return db
}

func seedEpisodeForUser(t *testing.T, db *gorm.DB, userID uint) models.Episode {
	t.Helper()
	drama := models.Drama{
		UserID: userID,
		Title:  "test drama",
	}
	require.NoError(t, db.Create(&drama).Error)

	ep := models.Episode{
		UserID:     userID,
		DramaID:    drama.ID,
		EpisodeNum: 1,
		Title:      "ep1",
	}
	require.NoError(t, db.Create(&ep).Error)
	return ep
}

func TestBackfillStoryboardUserID_FillsMissingUserID(t *testing.T) {
	db := newDataFixTestDB(t)
	ep := seedEpisodeForUser(t, db, 42)

	sb := models.Storyboard{
		UserID:           0,
		EpisodeID:        ep.ID,
		StoryboardNumber: 1,
		Duration:         5,
	}
	require.NoError(t, db.Create(&sb).Error)

	report, err := BackfillStoryboardUserID(db)
	require.NoError(t, err)
	require.EqualValues(t, 1, report.BackfilledRows)
	require.EqualValues(t, 0, report.RemainingZeroRows)
	require.EqualValues(t, 0, report.MismatchRows)
	require.EqualValues(t, 0, report.OrphanRows)

	var got models.Storyboard
	require.NoError(t, db.First(&got, sb.ID).Error)
	require.EqualValues(t, 42, got.UserID)
}

func TestBackfillStoryboardUserID_ReportsMismatch(t *testing.T) {
	db := newDataFixTestDB(t)
	ep := seedEpisodeForUser(t, db, 7)

	sb := models.Storyboard{
		UserID:           99,
		EpisodeID:        ep.ID,
		StoryboardNumber: 1,
		Duration:         5,
	}
	require.NoError(t, db.Create(&sb).Error)

	report, err := BackfillStoryboardUserID(db)
	require.NoError(t, err)
	require.EqualValues(t, 0, report.BackfilledRows)
	require.EqualValues(t, 1, report.MismatchRows)
}
