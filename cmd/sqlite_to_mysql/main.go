package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/pkg/config"
)

func main() {
	sourcePath := flag.String("source", "./data/drama_generator.db", "sqlite source database path")
	markerPath := flag.String("marker", "", "marker file written after a successful migration")
	waitAttempts := flag.Int("wait-attempts", 30, "mysql readiness attempts")
	waitInterval := flag.Duration("wait-interval", 2*time.Second, "mysql readiness retry interval")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.Database.Type != "mysql" {
		log.Fatalf("sqlite_to_mysql requires database.type=mysql, got %s", cfg.Database.Type)
	}

	if _, err := os.Stat(*sourcePath); err != nil {
		log.Fatalf("sqlite source file not found: %v", err)
	}

	if err := database.WaitForMySQL(cfg.Database, *waitAttempts, *waitInterval); err != nil {
		log.Fatalf("mysql is not ready: %v", err)
	}

	migrator, err := database.NewSQLiteToMySQLMigrator(*sourcePath, cfg.Database)
	if err != nil {
		log.Fatalf("failed to initialize sqlite_to_mysql migrator: %v", err)
	}

	report, err := migrator.Migrate()
	if err != nil {
		if errors.Is(err, database.ErrSQLiteToMySQLTargetNotEmpty) {
			fmt.Println("mysql target already contains data, skipping sqlite import")
			if err := database.WriteSQLiteToMySQLMarker(*markerPath); err != nil {
				log.Fatalf("sqlite_to_mysql marker write failed: %v", err)
			}
			return
		}
		log.Fatalf("sqlite_to_mysql migration failed: %v", err)
	}

	for _, table := range report.Tables {
		fmt.Printf("migrated table=%s rows=%d\n", table.Table, table.Rows)
	}

	if err := database.WriteSQLiteToMySQLMarker(*markerPath); err != nil {
		log.Fatalf("sqlite_to_mysql marker write failed: %v", err)
	}
}
