package services

import (
	"errors"
	"fmt"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminUserService struct {
	db    *gorm.DB
	log   *logger.Logger
	audit *AdminAuditService
}

func NewAdminUserService(db *gorm.DB, log *logger.Logger, audit *AdminAuditService) *AdminUserService {
	return &AdminUserService{
		db:    db,
		log:   log,
		audit: audit,
	}
}

func (s *AdminUserService) ListUsers(page, pageSize int) ([]models.User, int64, error) {
	page, pageSize = normalizePagination(page, pageSize)

	query := s.db.Model(&models.User{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	if err := s.db.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *AdminUserService) UpdateUserStatus(adminID, userID uint, status models.UserStatus, ip, userAgent string) (*models.User, error) {
	if !isSupportedUserStatus(status) {
		return nil, errors.New("invalid user status")
	}

	var updated models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		before := map[string]interface{}{"status": user.Status}
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
			return err
		}
		user.Status = status
		after := map[string]interface{}{"status": user.Status}

		if err := s.audit.WriteWithTx(
			tx,
			adminID,
			"user.update_status",
			"user",
			fmt.Sprintf("%d", userID),
			before,
			after,
			AdminActorMeta{IP: ip, UserAgent: userAgent},
		); err != nil {
			return err
		}
		updated = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Infow("admin updated user status", "admin_id", adminID, "user_id", userID, "status", status)
	return &updated, nil
}

func (s *AdminUserService) UpdateUserRole(adminID, userID uint, role models.UserRole, ip, userAgent string) (*models.User, error) {
	if !isSupportedUserRole(role) {
		return nil, errors.New("invalid user role")
	}

	var updated models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		before := map[string]interface{}{"role": user.Role}
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error; err != nil {
			return err
		}
		user.Role = role
		after := map[string]interface{}{"role": user.Role}

		if err := s.audit.WriteWithTx(
			tx,
			adminID,
			"user.update_role",
			"user",
			fmt.Sprintf("%d", userID),
			before,
			after,
			AdminActorMeta{IP: ip, UserAgent: userAgent},
		); err != nil {
			return err
		}
		updated = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Infow("admin updated user role", "admin_id", adminID, "user_id", userID, "role", role)
	return &updated, nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func isSupportedUserStatus(status models.UserStatus) bool {
	return status == models.UserStatusActive || status == models.UserStatusDisabled
}

func isSupportedUserRole(role models.UserRole) bool {
	switch role {
	case models.RoleUser, models.RoleVIP, models.RoleAdmin, models.RolePlatformAdmin:
		return true
	default:
		return false
	}
}
