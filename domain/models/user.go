package models

import "time"

type UserRole string

const (
	RoleAdmin         UserRole = "admin" // 兼容历史数据
	RolePlatformAdmin UserRole = "platform_admin"
	RoleUser          UserRole = "user"
	RoleVIP           UserRole = "vip"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Role         UserRole   `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	Status       UserStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Credits      int        `gorm:"not null;default:0" json:"credits"`
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type CreditTransactionType string

const (
	CreditTxnRecharge         CreditTransactionType = "RECHARGE"
	CreditTxnGenerateFrame    CreditTransactionType = "GENERATE_FRAME_PROMPT"
	CreditTxnGenerateImage    CreditTransactionType = "GENERATE_IMAGE"
	CreditTxnAIText           CreditTransactionType = "AI_TEXT"
	CreditTxnAITextRefund     CreditTransactionType = "AI_TEXT_REFUND"
	CreditTxnAIImage          CreditTransactionType = "AI_IMAGE"
	CreditTxnAIImageRefund    CreditTransactionType = "AI_IMAGE_REFUND"
	CreditTxnAIVideo          CreditTransactionType = "AI_VIDEO"
	CreditTxnAIVideoRefund    CreditTransactionType = "AI_VIDEO_REFUND"
)

type CreditTransaction struct {
	ID          uint                  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint                  `gorm:"not null;index" json:"user_id"`
	Amount      int                   `gorm:"not null" json:"amount"`
	Type        CreditTransactionType `gorm:"type:varchar(50);not null" json:"type"`
	ReferenceID *string               `gorm:"type:varchar(64);index" json:"reference_id,omitempty"`
	ServiceType *string               `gorm:"type:varchar(20);index" json:"service_type,omitempty"` // text, image, video
	Model       *string               `gorm:"type:varchar(100);index" json:"model,omitempty"`
	Description *string               `gorm:"type:varchar(255)" json:"description,omitempty"`
	CreatedAt   time.Time             `gorm:"not null;autoCreateTime" json:"created_at"`
}

func (CreditTransaction) TableName() string {
	return "credit_transactions"
}
