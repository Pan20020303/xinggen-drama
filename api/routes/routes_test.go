package routes_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drama-generator/backend/api/routes"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newRouterForStaticFileTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file:routes_static_file_test?mode=memory&cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.CreditTransaction{}, &models.AIServiceConfig{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cfg := &config.Config{
		App: config.AppConfig{Debug: true},
		Server: config.ServerConfig{
			CORSOrigins: []string{"http://localhost:3012"},
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: "./data/drama_generator.db",
		},
		Storage: config.StorageConfig{
			Type:      "local",
			LocalPath: "./data/storage",
			BaseURL:   "http://localhost:5678/static",
		},
		Auth: config.AuthConfig{
			JWTSecret:        "test-secret",
			TokenExpireHours: 1,
			InitialCredits:   0,
		},
	}

	log := logger.NewLogger(true)
	ls, _ := storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
	_ = database.AutoMigrate(db)

	return routes.SetupRouter(cfg, db, log, ls)
}

func TestSetupRouter_ServesBuiltPublicFileBeforeSPAFallback(t *testing.T) {
	r := newRouterForStaticFileTest(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/ocean-login-bg.png", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if bytes.Contains(bytes.ToLower(w.Body.Bytes()), []byte("<html")) {
		t.Fatalf("expected static png file, got html fallback")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", contentType)
	}
}
