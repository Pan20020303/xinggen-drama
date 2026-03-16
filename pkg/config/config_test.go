package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseConfigDSN_SQLite(t *testing.T) {
	cfg := DatabaseConfig{
		Type: "sqlite",
		Path: "./data/drama_generator.db",
	}

	require.Equal(t, "./data/drama_generator.db", cfg.DSN())
}

func TestDatabaseConfigDSN_MySQL(t *testing.T) {
	cfg := DatabaseConfig{
		Type:     "mysql",
		Host:     "mysql",
		Port:     3306,
		User:     "xinggen",
		Password: "secret",
		Database: "xinggen_drama",
		Charset:  "utf8mb4",
	}

	require.Equal(
		t,
		"xinggen:secret@tcp(mysql:3306)/xinggen_drama?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DSN(),
	)
}

func TestDatabaseConfigDSN_MySQLUsesDefaultCharset(t *testing.T) {
	cfg := DatabaseConfig{
		Type:     "mysql",
		Host:     "mysql",
		Port:     3306,
		User:     "xinggen",
		Password: "secret",
		Database: "xinggen_drama",
	}

	require.Equal(
		t,
		"xinggen:secret@tcp(mysql:3306)/xinggen_drama?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DSN(),
	)
}
