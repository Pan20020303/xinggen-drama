package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newMiddlewareTestAuthService(t *testing.T) (*services.AuthService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.CreditTransaction{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:        "test-secret",
			TokenExpireHours: 1,
			InitialCredits:   0,
		},
	}
	return services.NewAuthService(db, cfg, logger.NewLogger(true)), db
}

func seedMiddlewareUser(t *testing.T, db *gorm.DB, email string, role models.UserRole) models.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("Passw0rd123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Status:       models.UserStatusActive,
		Credits:      0,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func performRequestWithToken(r http.Handler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminMiddleware_UserTokenCannotAccessAdminRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc, db := newMiddlewareTestAuthService(t)
	user := seedMiddlewareUser(t, db, "user@example.com", models.RoleUser)

	userToken, err := authSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	r := gin.New()
	admin := r.Group("/admin")
	admin.Use(AdminAuthMiddleware(authSvc))
	admin.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	resp := performRequestWithToken(r, http.MethodGet, "/admin/ping", userToken)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, resp.Code)
	}
}

func TestAdminMiddleware_AdminTokenCanAccessAdminRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc, db := newMiddlewareTestAuthService(t)
	adminUser := seedMiddlewareUser(t, db, "admin@example.com", models.RolePlatformAdmin)

	adminToken, err := authSvc.GenerateAdminToken(adminUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	r := gin.New()
	admin := r.Group("/admin")
	admin.Use(AdminAuthMiddleware(authSvc))
	admin.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	resp := performRequestWithToken(r, http.MethodGet, "/admin/ping", adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestAdminMiddleware_AdminTokenCannotAccessUserRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc, db := newMiddlewareTestAuthService(t)
	adminUser := seedMiddlewareUser(t, db, "admin2@example.com", models.RolePlatformAdmin)

	adminToken, err := authSvc.GenerateAdminToken(adminUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	r := gin.New()
	user := r.Group("/user")
	user.Use(AuthMiddleware(authSvc))
	user.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	resp := performRequestWithToken(r, http.MethodGet, "/user/ping", adminToken)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, resp.Code)
	}
}
