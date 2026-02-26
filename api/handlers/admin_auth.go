package handlers

import (
	"errors"

	"github.com/drama-generator/backend/application/dto"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminAuthHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewAdminAuthHandler(authService *services.AuthService, log *logger.Logger) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService: authService,
		log:         log,
	}
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.AdminLogin(&req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) || errors.Is(err, services.ErrAdminAccessDenied) {
			response.Unauthorized(c, "账号或密码错误，或非管理员")
			return
		}
		if errors.Is(err, services.ErrUserDisabled) {
			response.Forbidden(c, "账号已禁用")
			return
		}
		h.log.Errorw("admin login failed", "error", err)
		response.InternalError(c, "管理员登录失败")
		return
	}

	response.Success(c, resp)
}
