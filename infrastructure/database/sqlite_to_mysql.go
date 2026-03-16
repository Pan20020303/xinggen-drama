package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	"gorm.io/gorm"
)

type SQLiteToMySQLMigrator struct {
	Source *gorm.DB
	Target *gorm.DB
}

type SQLiteToMySQLReport struct {
	Tables []SQLiteToMySQLTableReport
}

type SQLiteToMySQLTableReport struct {
	Table string
	Rows  int64
}

var ErrSQLiteToMySQLTargetNotEmpty = errors.New("mysql target database is not empty")

func sqliteToMySQLMigrationTables() []string {
	return []string{
		"users",
		"credit_transactions",
		"admin_audit_logs",
		"dramas",
		"episodes",
		"characters",
		"scenes",
		"props",
		"storyboards",
		"frame_prompts",
		"image_generations",
		"video_generations",
		"video_merges",
		"ai_service_configs",
		"ai_service_providers",
		"assets",
		"character_libraries",
		"async_tasks",
		"episode_characters",
		"storyboard_characters",
		"storyboard_props",
	}
}

func normalizeSQLiteValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return value
	}
}

func shouldAbortSQLiteToMySQLMigration(targetCounts map[string]int64) bool {
	for _, count := range targetCounts {
		if count > 0 {
			return true
		}
	}
	return false
}

func NewSQLiteToMySQLMigrator(sourcePath string, targetCfg config.DatabaseConfig) (*SQLiteToMySQLMigrator, error) {
	source, err := NewDatabase(config.DatabaseConfig{
		Type: "sqlite",
		Path: sourcePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite source: %w", err)
	}

	target, err := NewDatabase(targetCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql target: %w", err)
	}

	return &SQLiteToMySQLMigrator{
		Source: source,
		Target: target,
	}, nil
}

func WaitForMySQL(cfg config.DatabaseConfig, attempts int, interval time.Duration) error {
	var lastErr error

	for i := 0; i < attempts; i++ {
		db, err := NewDatabase(cfg)
		if err == nil {
			sqlDB, closeErr := db.DB()
			if closeErr == nil {
				_ = sqlDB.Close()
			}
			return nil
		}
		lastErr = err
		time.Sleep(interval)
	}

	return fmt.Errorf("mysql is not ready after %d attempts: %w", attempts, lastErr)
}

func (m *SQLiteToMySQLMigrator) Migrate() (*SQLiteToMySQLReport, error) {
	if err := AutoMigrate(m.Target); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate mysql schema: %w", err)
	}

	targetCounts, err := m.targetCounts()
	if err != nil {
		return nil, err
	}
	if shouldAbortSQLiteToMySQLMigration(targetCounts) {
		return nil, fmt.Errorf("%w, aborting sqlite migration to avoid duplicate imports", ErrSQLiteToMySQLTargetNotEmpty)
	}

	sourceTables, err := m.sourceTableSet()
	if err != nil {
		return nil, err
	}

	if err := m.Target.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		return nil, fmt.Errorf("failed to disable mysql foreign key checks: %w", err)
	}
	defer m.Target.Exec("SET FOREIGN_KEY_CHECKS=1")

	report := &SQLiteToMySQLReport{}
	for _, table := range sqliteToMySQLMigrationTables() {
		if !sourceTables[table] {
			continue
		}

		rows, err := m.copyTable(table)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate table %s: %w", table, err)
		}
		report.Tables = append(report.Tables, SQLiteToMySQLTableReport{
			Table: table,
			Rows:  rows,
		})
	}

	return report, nil
}

func (m *SQLiteToMySQLMigrator) sourceTableSet() (map[string]bool, error) {
	tables, err := m.Source.Migrator().GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to list sqlite tables: %w", err)
	}

	set := make(map[string]bool, len(tables))
	for _, table := range tables {
		set[table] = true
	}
	return set, nil
}

func (m *SQLiteToMySQLMigrator) targetCounts() (map[string]int64, error) {
	counts := make(map[string]int64)

	for _, table := range sqliteToMySQLMigrationTables() {
		if !m.Target.Migrator().HasTable(table) {
			continue
		}
		var count int64
		if err := m.Target.Table(table).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count mysql table %s: %w", table, err)
		}
		counts[table] = count
	}

	return counts, nil
}

func (m *SQLiteToMySQLMigrator) copyTable(table string) (int64, error) {
	rows, err := m.Source.Table(table).Rows()
	if err != nil {
		return 0, fmt.Errorf("failed to query sqlite table %s: %w", table, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return 0, fmt.Errorf("failed to read sqlite columns for %s: %w", table, err)
	}

	var imported int64
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanTargets := make([]interface{}, len(columns))
		for i := range values {
			scanTargets[i] = &values[i]
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return imported, fmt.Errorf("failed to scan sqlite row for %s: %w", table, err)
		}

		record := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			record[column] = normalizeSQLiteValue(values[i])
		}

		if err := m.Target.Table(table).Create(record).Error; err != nil {
			return imported, fmt.Errorf("failed to insert mysql row into %s: %w", table, err)
		}
		imported++
	}

	if err := rows.Err(); err != nil {
		return imported, fmt.Errorf("sqlite row iteration failed for %s: %w", table, err)
	}

	return imported, nil
}

func WriteSQLiteToMySQLMarker(markerPath string) error {
	if markerPath == "" {
		return nil
	}

	dir := filepath.Dir(markerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create marker directory: %w", err)
	}

	content := []byte(time.Now().Format(time.RFC3339))
	if err := os.WriteFile(markerPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write marker file: %w", err)
	}
	return nil
}
