package models

import "time"

type AdminAuditLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID    uint      `gorm:"not null;index" json:"admin_id"`
	Action     string    `gorm:"type:varchar(100);not null;index" json:"action"`
	TargetType string    `gorm:"type:varchar(100);not null;index" json:"target_type"`
	TargetID   string    `gorm:"type:varchar(100);not null;index" json:"target_id"`
	BeforeJSON *string   `gorm:"type:text" json:"before_json,omitempty"`
	AfterJSON  *string   `gorm:"type:text" json:"after_json,omitempty"`
	IP         *string   `gorm:"type:varchar(64)" json:"ip,omitempty"`
	UserAgent  *string   `gorm:"type:varchar(512)" json:"user_agent,omitempty"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
}

func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}
