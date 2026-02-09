package services

import (
	"encoding/json"

	"github.com/drama-generator/backend/domain/models"
	"gorm.io/gorm"
)

type AdminActorMeta struct {
	IP        string
	UserAgent string
}

type AdminAuditService struct {
	db *gorm.DB
}

func NewAdminAuditService(db *gorm.DB) *AdminAuditService {
	return &AdminAuditService{db: db}
}

func (s *AdminAuditService) WriteWithTx(
	tx *gorm.DB,
	adminID uint,
	action string,
	targetType string,
	targetID string,
	before interface{},
	after interface{},
	meta AdminActorMeta,
) error {
	beforeJSON, err := marshalAuditPayload(before)
	if err != nil {
		return err
	}
	afterJSON, err := marshalAuditPayload(after)
	if err != nil {
		return err
	}

	log := models.AdminAuditLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		BeforeJSON: beforeJSON,
		AfterJSON:  afterJSON,
	}
	if meta.IP != "" {
		ip := meta.IP
		log.IP = &ip
	}
	if meta.UserAgent != "" {
		ua := meta.UserAgent
		log.UserAgent = &ua
	}

	exec := tx
	if exec == nil {
		exec = s.db
	}
	return exec.Create(&log).Error
}

func marshalAuditPayload(v interface{}) (*string, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	str := string(b)
	return &str, nil
}
