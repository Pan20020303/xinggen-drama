package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type apiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newAdminAIConfigTestEnv(t *testing.T) (*gin.Engine, *gorm.DB, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file:admin_ai_config_test_" + time.Now().Format("150405.000000000") + "?mode=memory&cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.AIServiceConfig{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:        "test-secret",
			TokenExpireHours: 1,
			InitialCredits:   0,
		},
	}
	log := logger.NewLogger(true)
	authSvc := services.NewAuthService(db, cfg, log)
	adminToken, err := authSvc.GenerateAdminToken(models.User{ID: 1, Email: "admin@example.com", Role: models.RolePlatformAdmin})
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	h := NewAdminAIConfigHandler(db, cfg, log)

	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("/admin")
	admin.Use(middlewares.AdminAuthMiddleware(authSvc))

	admin.GET("/ai-configs", h.ListConfigs)
	admin.POST("/ai-configs", h.CreateConfig)
	admin.PUT("/ai-configs/:id", h.UpdateConfig)
	admin.DELETE("/ai-configs/:id", h.DeleteConfig)
	admin.POST("/ai-configs/test", h.TestConnection)

	return r, db, adminToken
}

func doReq(t *testing.T, r http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminAIConfig_CRUD_ListAndMaskedAPIKey(t *testing.T) {
	r, db, token := newAdminAIConfigTestEnv(t)

	createResp := doReq(t, r, http.MethodPost, "/api/v1/admin/ai-configs", token, map[string]interface{}{
		"service_type": "text",
		"name":         "platform-text",
		"provider":     "openai",
		"base_url":     "https://api.example.com",
		"api_key":      "secret",
		"model":        []string{"gpt-1"},
		"priority":     10,
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d: %s", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	// Ensure DB record stored as platform config with real api key.
	var stored models.AIServiceConfig
	if err := db.Order("id DESC").First(&stored).Error; err != nil {
		t.Fatalf("expected stored config, got %v", err)
	}
	if stored.UserID != 0 {
		t.Fatalf("expected platform user_id=0, got %d", stored.UserID)
	}
	if stored.APIKey != "secret" {
		t.Fatalf("expected api key stored, got %q", stored.APIKey)
	}

	// List should not expose api_key.
	listResp := doReq(t, r, http.MethodGet, "/api/v1/admin/ai-configs?service_type=text", token, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, listResp.Code, listResp.Body.String())
	}

	var list apiResponse[[]map[string]interface{}]
	if err := json.Unmarshal(listResp.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to unmarshal list response: %v", err)
	}
	if !list.Success {
		t.Fatalf("expected success=true")
	}
	if len(list.Data) == 0 {
		t.Fatalf("expected non-empty list")
	}
	if v, ok := list.Data[0]["api_key"]; ok && v.(string) != "" {
		t.Fatalf("expected api_key to be empty in response, got %v", v)
	}
	if v, ok := list.Data[0]["api_key_set"]; !ok || v.(bool) != true {
		t.Fatalf("expected api_key_set=true, got %v", list.Data[0]["api_key_set"])
	}

	// Update name
	updateResp := doReq(t, r, http.MethodPut, "/api/v1/admin/ai-configs/1", token, map[string]interface{}{
		"name":      "platform-text-updated",
		"is_active": true,
	})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, updateResp.Code, updateResp.Body.String())
	}

	// Delete
	delResp := doReq(t, r, http.MethodDelete, "/api/v1/admin/ai-configs/1", token, nil)
	if delResp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, delResp.Code, delResp.Body.String())
	}
}

func TestAdminAIConfig_TestConnection_ValidatesPayload(t *testing.T) {
	r, _, token := newAdminAIConfigTestEnv(t)

	resp := doReq(t, r, http.MethodPost, "/api/v1/admin/ai-configs/test", token, map[string]interface{}{
		"base_url": "not-a-url",
		"api_key":  "secret",
		"model":    []string{"gpt-1"},
		"provider": "openai",
	})
	// Should fail fast on binding validation (BAD_REQUEST).
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}
