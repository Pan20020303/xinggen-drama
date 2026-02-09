package handlers

import (
	"errors"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminAuthHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewAdminAuthHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService: services.NewAuthService(db, cfg, log),
		log:         log,
	}
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.AdminLogin(&req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "invalid credentials" || err.Error() == "admin access denied" {
			response.Unauthorized(c, "账号或密码错误，或非管理员")
			return
		}
		if err.Error() == "user disabled" {
			response.Forbidden(c, "账号已禁用")
			return
		}
		h.log.Errorw("admin login failed", "error", err)
		response.InternalError(c, "管理员登录失败")
		return
	}

	response.Success(c, resp)
}
