package services

import (
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type AdminTokenStatsService struct {
	db  *gorm.DB
	log *logger.Logger
}

type AdminTokenStatsItem struct {
	Model            string `json:"model"`
	ServiceType      string `json:"service_type"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	Calls            int    `json:"calls"`
}

type AdminTokenStatsSummary struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	ModelCount       int `json:"model_count"`
}

func NewAdminTokenStatsService(db *gorm.DB, log *logger.Logger) *AdminTokenStatsService {
	return &AdminTokenStatsService{db: db, log: log}
}

func (s *AdminTokenStatsService) GetTokenStats(serviceType *string, startDate, endDate *time.Time) ([]AdminTokenStatsItem, AdminTokenStatsSummary, error) {
	query := s.db.Model(&models.CreditTransaction{}).
		Select(
			"COALESCE(model, '') AS model",
			"COALESCE(service_type, '') AS service_type",
			"COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens",
			"COALESCE(SUM(completion_tokens), 0) AS completion_tokens",
			"COALESCE(SUM(total_tokens), 0) AS total_tokens",
			"COUNT(*) AS calls",
		).
		Where("total_tokens IS NOT NULL").
		Group("model, service_type").
		Order("SUM(total_tokens) DESC")

	if serviceType != nil && *serviceType != "" {
		query = query.Where("service_type = ?", *serviceType)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var items []AdminTokenStatsItem
	if err := query.Scan(&items).Error; err != nil {
		return nil, AdminTokenStatsSummary{}, err
	}

	summary := AdminTokenStatsSummary{}
	for _, item := range items {
		summary.PromptTokens += item.PromptTokens
		summary.CompletionTokens += item.CompletionTokens
		summary.TotalTokens += item.TotalTokens
	}
	summary.ModelCount = len(items)

	return items, summary, nil
}
