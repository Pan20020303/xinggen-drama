package handlers_test

import (
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

func newRouterForDisabledAITest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file:ai_config_disabled_test?mode=memory&cache=shared",
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

	// SetupRouter expects a LocalStorage pointer when local storage enabled.
	ls, _ := storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
	_ = database.AutoMigrate(db)
	return routes.SetupRouter(cfg, db, log, ls)
}

func TestAIConfigDisabled_UserEndpointsNotMounted(t *testing.T) {
	r := newRouterForDisabledAITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-configs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
