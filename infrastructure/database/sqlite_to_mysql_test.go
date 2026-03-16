package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLiteToMySQLMigrationTables_HasStableDependencyOrder(t *testing.T) {
	tables := sqliteToMySQLMigrationTables()

	require.Contains(t, tables, "users")
	require.Contains(t, tables, "dramas")
	require.Contains(t, tables, "episodes")
	require.Contains(t, tables, "storyboards")
	require.Contains(t, tables, "image_generations")
	require.Contains(t, tables, "video_generations")

	indexOf := func(name string) int {
		for i, table := range tables {
			if table == name {
				return i
			}
		}
		return -1
	}

	require.Less(t, indexOf("users"), indexOf("dramas"))
	require.Less(t, indexOf("dramas"), indexOf("episodes"))
	require.Less(t, indexOf("episodes"), indexOf("storyboards"))
	require.Less(t, indexOf("storyboards"), indexOf("image_generations"))
	require.Less(t, indexOf("storyboards"), indexOf("video_generations"))
}

func TestSQLiteToMySQLNormalizeValue(t *testing.T) {
	require.Equal(t, "hello", normalizeSQLiteValue([]byte("hello")))
	require.Nil(t, normalizeSQLiteValue(nil))
	require.Equal(t, 12, normalizeSQLiteValue(12))
}

func TestSQLiteToMySQLShouldAbortWhenTargetNotEmpty(t *testing.T) {
	require.True(t, shouldAbortSQLiteToMySQLMigration(map[string]int64{
		"users": 1,
	}))
	require.False(t, shouldAbortSQLiteToMySQLMigration(map[string]int64{
		"users": 0,
		"dramas": 0,
	}))
}
