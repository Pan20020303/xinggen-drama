package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
	"github.com/gin-gonic/gin"
)

type ScriptGenerationHandler struct {
	scriptService *services.ScriptGenerationService
	taskService   *services.TaskService
	log           *logger.Logger
}

func NewScriptGenerationHandler(scriptService *services.ScriptGenerationService, taskService *services.TaskService, log *logger.Logger) *ScriptGenerationHandler {
	return &ScriptGenerationHandler{
		scriptService: scriptService,
		taskService:   taskService,
		log:           log,
	}
}

func (h *ScriptGenerationHandler) GenerateCharacters(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req services.GenerateCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.scriptService.GenerateCharacters(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		h.log.Errorw("Failed to generate characters", "error", err, "drama_id", req.DramaID)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "角色生成任务已创建，正在后台处理...",
	})
}

func (h *ScriptGenerationHandler) PolishEpisodeScript(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	episodeIDUint, err := strconv.ParseUint(c.Param("episode_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}

	var req struct {
		Content   string `json:"content"`
		Model     string `json:"model"`
		SkillName string `json:"skill_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空body，默认使用数据库中的章节内容
		req.Content = ""
		req.Model = ""
		req.SkillName = ""
	}

	polished, usedSkill, err := h.scriptService.PolishEpisodeScript(
		userID,
		uint(episodeIDUint),
		req.Content,
		req.Model,
		req.SkillName,
	)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		if isUpstreamAITimeout(err) {
			response.Error(c, http.StatusBadGateway, "UPSTREAM_TIMEOUT", "AI服务连接超时，请稍后重试")
			return
		}
		if errors.Is(err, services.ErrEpisodeNotFound) {
			response.NotFound(c, "章节不存在")
			return
		}
		if err.Error() == "empty content" {
			response.BadRequest(c, "章节内容为空，无法润色")
			return
		}
		h.log.Errorw("Failed to polish episode script",
			"error", err,
			"episode_id", episodeIDUint,
			"user_id", userID)
		response.InternalError(c, "润色失败")
		return
	}

	response.Success(c, gin.H{
		"content":    polished,
		"skill_name": usedSkill,
	})
}

func (h *ScriptGenerationHandler) PolishScriptText(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req struct {
		Content   string `json:"content" binding:"required"`
		Model     string `json:"model"`
		SkillName string `json:"skill_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	polished, usedSkill, err := h.scriptService.PolishScriptText(
		userID,
		req.Content,
		req.Model,
		req.SkillName,
	)
	if err != nil {
		if errors.Is(err, services.ErrInsufficientCredits) {
			response.Forbidden(c, "积分不足")
			return
		}
		if isUpstreamAITimeout(err) {
			response.Error(c, http.StatusBadGateway, "UPSTREAM_TIMEOUT", "AI服务连接超时，请稍后重试")
			return
		}
		if err.Error() == "empty content" {
			response.BadRequest(c, "章节内容为空，无法润色")
			return
		}
		h.log.Errorw("Failed to polish script text", "error", err, "user_id", userID)
		response.InternalError(c, "润色失败")
		return
	}

	response.Success(c, gin.H{
		"content":    polished,
		"skill_name": usedSkill,
	})
}

func (h *ScriptGenerationHandler) PolishScriptTextStream(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req struct {
		Content   string `json:"content" binding:"required"`
		Model     string `json:"model"`
		SkillName string `json:"skill_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.InternalError(c, "当前环境不支持流式返回")
		return
	}

	chunkCallback := func(chunk string, totalChars int, estimatedProgress float64) {
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}
		_ = writeSSE(c, flusher, "chunk", gin.H{
			"content":   chunk,
			"total":     totalChars,
			"progress":  estimatedProgress,
		})
	}

	polished, usedSkill, err := h.scriptService.PolishScriptTextStream(
		userID,
		req.Content,
		req.Model,
		req.SkillName,
		chunkCallback,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "润色失败"
		code := "INTERNAL_ERROR"
		switch {
		case errors.Is(err, services.ErrInsufficientCredits):
			statusCode = http.StatusForbidden
			message = "积分不足"
			code = "FORBIDDEN"
		case isUpstreamAITimeout(err):
			statusCode = http.StatusBadGateway
			message = "AI服务连接超时，请稍后重试"
			code = "UPSTREAM_TIMEOUT"
		case err.Error() == "empty content":
			statusCode = http.StatusBadRequest
			message = "章节内容为空，无法润色"
			code = "BAD_REQUEST"
		}
		h.log.Errorw("Failed to polish script text stream", "error", err, "user_id", userID)
		_ = writeSSE(c, flusher, "error", gin.H{
			"code":       code,
			"status":     statusCode,
			"message":    message,
			"raw_message": err.Error(),
		})
		return
	}

	_ = writeSSE(c, flusher, "done", gin.H{
		"content":    polished,
		"skill_name": usedSkill,
	})
}

func writeSSE(c *gin.Context, flusher http.Flusher, event string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func isUpstreamAITimeout(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls handshake timeout") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "context deadline exceeded")
}
