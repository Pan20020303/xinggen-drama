package database

import "gorm.io/gorm"

type StoryboardUserIDBackfillReport struct {
	BackfilledRows    int64
	RemainingZeroRows int64
	MismatchRows      int64
	OrphanRows        int64
}

// BackfillStoryboardUserID backfills missing storyboard.user_id from episodes.user_id.
// It also returns integrity counters for startup self-check logging.
func BackfillStoryboardUserID(db *gorm.DB) (*StoryboardUserIDBackfillReport, error) {
	result := db.Exec(`
UPDATE storyboards
SET user_id = (
	SELECT episodes.user_id
	FROM episodes
	WHERE episodes.id = storyboards.episode_id
)
WHERE storyboards.user_id = 0
  AND EXISTS (
	SELECT 1
	FROM episodes
	WHERE episodes.id = storyboards.episode_id
)`)
	if result.Error != nil {
		return nil, result.Error
	}

	report := &StoryboardUserIDBackfillReport{
		BackfilledRows: result.RowsAffected,
	}

	if err := db.Raw(`SELECT COUNT(1) FROM storyboards WHERE user_id = 0`).Scan(&report.RemainingZeroRows).Error; err != nil {
		return nil, err
	}
	if err := db.Raw(`
SELECT COUNT(1)
FROM storyboards s
JOIN episodes e ON e.id = s.episode_id
WHERE s.user_id <> e.user_id`).Scan(&report.MismatchRows).Error; err != nil {
		return nil, err
	}
	if err := db.Raw(`
SELECT COUNT(1)
FROM storyboards s
LEFT JOIN episodes e ON e.id = s.episode_id
WHERE e.id IS NULL`).Scan(&report.OrphanRows).Error; err != nil {
		return nil, err
	}

	return report, nil
}

