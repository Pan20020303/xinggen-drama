package handlers

import (
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminAIConfigHandler struct {
	aiService *services.AIService
	log       *logger.Logger
}

func NewAdminAIConfigHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AdminAIConfigHandler {
	_ = cfg
	return &AdminAIConfigHandler{
		aiService: services.NewAIService(db, log),
		log:       log,
	}
}

func (h *AdminAIConfigHandler) CreateConfig(c *gin.Context) {
	var req services.CreateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cfg, err := h.aiService.CreateConfig(&req /* platform scope */)
	if err != nil {
		h.log.Errorw("failed to create platform ai config", "error", err)
		response.InternalError(c, "创建失败")
		return
	}
	response.Created(c, services.ToAIServiceConfigView(*cfg))
}

func (h *AdminAIConfigHandler) ListConfigs(c *gin.Context) {
	serviceType := c.Query("service_type")
	cfgs, err := h.aiService.ListPlatformConfigs(serviceType)
	if err != nil {
		response.InternalError(c, "获取列表失败")
		return
	}
	response.Success(c, services.ToAIServiceConfigViews(cfgs))
}

func (h *AdminAIConfigHandler) UpdateConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	var req services.UpdateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cfg, err := h.aiService.UpdatePlatformConfig(uint(configID), &req)
	if err != nil {
		if err.Error() == "config not found" {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "更新失败")
		return
	}
	response.Success(c, services.ToAIServiceConfigView(*cfg))
}

func (h *AdminAIConfigHandler) DeleteConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}
	if err := h.aiService.DeletePlatformConfig(uint(configID)); err != nil {
		if err.Error() == "config not found" {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *AdminAIConfigHandler) TestConnection(c *gin.Context) {
	var req services.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.aiService.TestConnection(&req); err != nil {
		response.BadRequest(c, "连接测试失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "连接测试成功"})
}
