package services

import (
	"fmt"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newAdminServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.CreditTransaction{}, &models.AdminAuditLog{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func seedAdminServiceUser(t *testing.T, db *gorm.DB, email string, role models.UserRole, status models.UserStatus, credits int) models.User {
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

func TestAdminService_ListUsersPagination(t *testing.T) {
	db := newAdminServiceTestDB(t)
	log := logger.NewLogger(true)
	userSvc := NewAdminUserService(db, log, NewAdminAuditService(db))

	seedAdminServiceUser(t, db, "u1@example.com", models.RoleUser, models.UserStatusActive, 10)
	seedAdminServiceUser(t, db, "u2@example.com", models.RoleUser, models.UserStatusActive, 20)
	seedAdminServiceUser(t, db, "u3@example.com", models.RoleVIP, models.UserStatusActive, 30)

	users, total, err := userSvc.ListUsers(1, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestAdminService_UpdateUserStatus(t *testing.T) {
	db := newAdminServiceTestDB(t)
	log := logger.NewLogger(true)
	auditSvc := NewAdminAuditService(db)
	userSvc := NewAdminUserService(db, log, auditSvc)

	admin := seedAdminServiceUser(t, db, "admin@example.com", models.RolePlatformAdmin, models.UserStatusActive, 0)
	target := seedAdminServiceUser(t, db, "target@example.com", models.RoleUser, models.UserStatusActive, 10)

	updated, err := userSvc.UpdateUserStatus(admin.ID, target.ID, models.UserStatusDisabled, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.UserStatusDisabled {
		t.Fatalf("expected status disabled, got %s", updated.Status)
	}

	var audit models.AdminAuditLog
	if err := db.Where("admin_id = ? AND action = ? AND target_id = ?", admin.ID, "user.update_status", fmt.Sprintf("%d", target.ID)).First(&audit).Error; err != nil {
		t.Fatalf("expected status audit log, got error: %v", err)
	}
}

func TestAdminService_UpdateUserRole(t *testing.T) {
	db := newAdminServiceTestDB(t)
	log := logger.NewLogger(true)
	auditSvc := NewAdminAuditService(db)
	userSvc := NewAdminUserService(db, log, auditSvc)

	admin := seedAdminServiceUser(t, db, "admin2@example.com", models.RolePlatformAdmin, models.UserStatusActive, 0)
	target := seedAdminServiceUser(t, db, "target2@example.com", models.RoleUser, models.UserStatusActive, 10)

	updated, err := userSvc.UpdateUserRole(admin.ID, target.ID, models.RoleVIP, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Role != models.RoleVIP {
		t.Fatalf("expected role vip, got %s", updated.Role)
	}

	var audit models.AdminAuditLog
	if err := db.Where("admin_id = ? AND action = ? AND target_id = ?", admin.ID, "user.update_role", fmt.Sprintf("%d", target.ID)).First(&audit).Error; err != nil {
		t.Fatalf("expected role audit log, got error: %v", err)
	}
}

func TestAdminService_RechargeWritesTxnAndAudit(t *testing.T) {
	db := newAdminServiceTestDB(t)
	log := logger.NewLogger(true)
	auditSvc := NewAdminAuditService(db)
	billingSvc := NewAdminBillingService(db, log, auditSvc)

	admin := seedAdminServiceUser(t, db, "admin3@example.com", models.RolePlatformAdmin, models.UserStatusActive, 0)
	target := seedAdminServiceUser(t, db, "target3@example.com", models.RoleUser, models.UserStatusActive, 10)

	updated, txn, err := billingSvc.RechargeUser(admin.ID, target.ID, 30, "manual recharge", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Credits != 40 {
		t.Fatalf("expected credits=40, got %d", updated.Credits)
	}
	if txn.Type != models.CreditTxnRecharge || txn.Amount != 30 {
		t.Fatalf("unexpected txn: %+v", txn)
	}

	var audit models.AdminAuditLog
	if err := db.Where("admin_id = ? AND action = ? AND target_id = ?", admin.ID, "billing.recharge", fmt.Sprintf("%d", target.ID)).First(&audit).Error; err != nil {
		t.Fatalf("expected billing audit log, got error: %v", err)
	}
}
