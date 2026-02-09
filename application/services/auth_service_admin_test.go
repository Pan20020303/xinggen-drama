package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func hasAudience(audiences []string, target string) bool {
	for _, aud := range audiences {
		if aud == target {
			return true
		}
	}
	return false
}

func newTestAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        fmt.Sprintf("file:auth_admin_%d?mode=memory&cache=shared", time.Now().UnixNano()),
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
			InitialCredits:   100,
		},
	}
	svc := NewAuthService(db, cfg, logger.NewLogger(true))
	return svc, db
}

func seedAuthUser(t *testing.T, db *gorm.DB, email string, role models.UserRole, status models.UserStatus) models.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("Passw0rd123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash pwd: %v", err)
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Status:       status,
		Credits:      100,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func TestAdminLoginRejectsNonAdmin(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "user@example.com", models.RoleUser, models.UserStatusActive)

	_, err := svc.AdminLogin(&LoginRequest{
		Email:    "user@example.com",
		Password: "Passw0rd123",
	})
	if err == nil {
		t.Fatalf("expected error for non-admin user")
	}
	if !strings.Contains(err.Error(), "admin") {
		t.Fatalf("expected admin-related error, got %v", err)
	}
}

func TestAdminLoginUsesAdminAudience(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "admin@example.com", models.RolePlatformAdmin, models.UserStatusActive)

	resp, err := svc.AdminLogin(&LoginRequest{
		Email:    "admin@example.com",
		Password: "Passw0rd123",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	claims, err := svc.ParseToken(resp.Token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !hasAudience(claims.Audience, "admin") {
		t.Fatalf("expected admin audience, got %v", claims.Audience)
	}
}

func TestUserLoginUsesUserAudience(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "normal@example.com", models.RoleUser, models.UserStatusActive)

	resp, err := svc.Login(&LoginRequest{
		Email:    "normal@example.com",
		Password: "Passw0rd123",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	claims, err := svc.ParseToken(resp.Token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !hasAudience(claims.Audience, "user") {
		t.Fatalf("expected user audience, got %v", claims.Audience)
	}
}
