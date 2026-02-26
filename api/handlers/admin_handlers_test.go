package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/persistence"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type adminHandlerTestEnv struct {
	DB     *gorm.DB
	Router *gin.Engine
	Auth   *services.AuthService
	Admin  models.User
	Target models.User
}

func newAdminHandlerTestEnv(t *testing.T) *adminHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:admin_handlers_%d?mode=memory&cache=shared", time.Now().UnixNano()),
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.CreditTransaction{}, &models.AdminAuditLog{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:        "test-secret",
			TokenExpireHours: 1,
			InitialCredits:   0,
		},
	}
	log := logger.NewLogger(true)
	repo := persistence.NewGormUserRepository(db)
	authSvc := services.NewAuthService(repo, cfg, log)

	admin := seedAdminHandlerUser(t, db, "admin@example.com", models.RolePlatformAdmin, models.UserStatusActive, 0)
	target := seedAdminHandlerUser(t, db, "target@example.com", models.RoleUser, models.UserStatusActive, 10)

	authHandler := NewAdminAuthHandler(authSvc, log)
	userHandler := NewAdminUserHandler(db, log)
	billingHandler := NewAdminBillingHandler(db, log)

	r := gin.New()
	api := r.Group("/api/v1")
	api.POST("/admin/auth/login", authHandler.Login)

	adminGroup := api.Group("/admin")
	adminGroup.Use(middlewares.AdminAuthMiddleware(authSvc))
	adminGroup.GET("/users", userHandler.ListUsers)
	adminGroup.PATCH("/users/:id/status", userHandler.UpdateUserStatus)
	adminGroup.PATCH("/users/:id/role", userHandler.UpdateUserRole)
	adminGroup.POST("/billing/recharge", billingHandler.Recharge)
	adminGroup.GET("/billing/transactions", billingHandler.ListTransactions)

	return &adminHandlerTestEnv{
		DB:     db,
		Router: r,
		Auth:   authSvc,
		Admin:  admin,
		Target: target,
	}
}

func seedAdminHandlerUser(t *testing.T, db *gorm.DB, email string, role models.UserRole, status models.UserStatus, credits int) models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("Passw0rd123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Status:       status,
		Credits:      credits,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func doJSONRequest(r http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func mustAdminToken(t *testing.T, env *adminHandlerTestEnv) string {
	t.Helper()
	token, err := env.Auth.GenerateAdminToken(env.Admin)
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}
	return token
}

func TestAdminHandler_AdminLoginEndpoint(t *testing.T) {
	env := newAdminHandlerTestEnv(t)

	resp := doJSONRequest(env.Router, http.MethodPost, "/api/v1/admin/auth/login", "", map[string]interface{}{
		"email":    "admin@example.com",
		"password": "Passw0rd123",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestAdminHandler_UsersListEndpoint(t *testing.T) {
	env := newAdminHandlerTestEnv(t)
	token := mustAdminToken(t, env)

	resp := doJSONRequest(env.Router, http.MethodGet, "/api/v1/admin/users?page=1&page_size=20", token, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestAdminHandler_StatusRolePatchEndpoints(t *testing.T) {
	env := newAdminHandlerTestEnv(t)
	token := mustAdminToken(t, env)

	statusResp := doJSONRequest(env.Router, http.MethodPatch, "/api/v1/admin/users/"+strconv.FormatUint(uint64(env.Target.ID), 10)+"/status", token, map[string]interface{}{
		"status": "disabled",
	})
	if statusResp.Code != http.StatusOK {
		t.Fatalf("status patch expected %d, got %d", http.StatusOK, statusResp.Code)
	}

	roleResp := doJSONRequest(env.Router, http.MethodPatch, "/api/v1/admin/users/"+strconv.FormatUint(uint64(env.Target.ID), 10)+"/role", token, map[string]interface{}{
		"role": "vip",
	})
	if roleResp.Code != http.StatusOK {
		t.Fatalf("role patch expected %d, got %d", http.StatusOK, roleResp.Code)
	}
}

func TestAdminHandler_RechargeEndpoint(t *testing.T) {
	env := newAdminHandlerTestEnv(t)
	token := mustAdminToken(t, env)

	resp := doJSONRequest(env.Router, http.MethodPost, "/api/v1/admin/billing/recharge", token, map[string]interface{}{
		"user_id": env.Target.ID,
		"amount":  30,
		"note":    "manual recharge",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestAdminHandler_TransactionsListEndpoint(t *testing.T) {
	env := newAdminHandlerTestEnv(t)
	token := mustAdminToken(t, env)

	_ = doJSONRequest(env.Router, http.MethodPost, "/api/v1/admin/billing/recharge", token, map[string]interface{}{
		"user_id": env.Target.ID,
		"amount":  10,
		"note":    "seed txn",
	})

	resp := doJSONRequest(
		env.Router,
		http.MethodGet,
		"/api/v1/admin/billing/transactions?user_id="+strconv.FormatUint(uint64(env.Target.ID), 10)+"&page=1&page_size=20",
		token,
		nil,
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}
