package handlers

import (
	"errors"
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoGenerationHandler struct {
	videoService *services.VideoGenerationService
	log          *logger.Logger
}

func NewVideoGenerationHandler(db *gorm.DB, cfg *config.Config, transferService *services.ResourceTransferService, localStorage *storage.LocalStorage, aiService *services.AIService, log *logger.Logger, promptI18n *services.PromptI18n) *VideoGenerationHandler {
	return &VideoGenerationHandler{
		videoService: services.NewVideoGenerationService(db, cfg, transferService, localStorage, aiService, log, promptI18n),
		log:          log,
	}
}

func (h *VideoGenerationHandler) GenerateVideo(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req services.GenerateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	videoGen, err := h.videoService.GenerateVideo(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		h.log.Errorw("Failed to generate video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) GenerateVideoFromImage(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	imageGenID, err := strconv.ParseUint(c.Param("image_gen_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的图片ID")
		return
	}

	videoGen, err := h.videoService.GenerateVideoFromImage(userID, uint(imageGenID))
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		h.log.Errorw("Failed to generate video from image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) BatchGenerateForEpisode(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	episodeID := c.Param("episode_id")

	videos, err := h.videoService.BatchGenerateVideosForEpisode(userID, episodeID)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		h.log.Errorw("Failed to batch generate videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videos)
}

func (h *VideoGenerationHandler) GetVideoGeneration(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	videoGen, err := h.videoService.GetVideoGeneration(userID, uint(videoGenID))
	if err != nil {
		response.NotFound(c, "视频生成记录不存在")
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) ListVideoGenerations(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var storyboardID *uint
	// 优先使用storyboard_id参数
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		id, err := strconv.ParseUint(storyboardIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var dramaIDUint *uint
	if dramaIDStr := c.Query("drama_id"); dramaIDStr != "" {
		did, _ := strconv.ParseUint(dramaIDStr, 10, 32)
		didUint := uint(did)
		dramaIDUint = &didUint
	}

	// 计算offset：(page - 1) * pageSize
	offset := (page - 1) * pageSize
	videos, total, err := h.videoService.ListVideoGenerations(userID, dramaIDUint, storyboardID, status, pageSize, offset)

	if err != nil {
		h.log.Errorw("Failed to list videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPagination(c, videos, total, page, pageSize)
}

func (h *VideoGenerationHandler) DeleteVideoGeneration(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.videoService.DeleteVideoGeneration(userID, uint(videoGenID)); err != nil {
		h.log.Errorw("Failed to delete video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
