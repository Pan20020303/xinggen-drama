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
)

type CreditTransaction struct {
	ID          uint                  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint                  `gorm:"not null;index" json:"user_id"`
	Amount      int                   `gorm:"not null" json:"amount"`
	Type        CreditTransactionType `gorm:"type:varchar(50);not null" json:"type"`
	Description *string               `gorm:"type:varchar(255)" json:"description,omitempty"`
	CreatedAt   time.Time             `gorm:"not null;autoCreateTime" json:"created_at"`
}

func (CreditTransaction) TableName() string {
	return "credit_transactions"
}
