package services

import (
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type ResourceTransferService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewResourceTransferService(db *gorm.DB, log *logger.Logger) *ResourceTransferService {
	return &ResourceTransferService{
		db:  db,
		log: log,
	}
}

// ResourceTransferService 现在只保留基本结构，MinIO相关功能已移除
// 如需资源转存功能，请使用本地存储

func (s *ResourceTransferService) BatchTransferImagesToMinio(dramaID string, limit int) (int, error) {
	s.log.Infow("Batch image transfer to MinIO skipped (feature removed)", "drama_id", dramaID, "limit", limit)
	return 0, nil
}

func (s *ResourceTransferService) BatchTransferVideosToMinio(dramaID string, limit int) (int, error) {
	s.log.Infow("Batch video transfer to MinIO skipped (feature removed)", "drama_id", dramaID, "limit", limit)
	return 0, nil
}
