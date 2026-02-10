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

type AuthHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(db, cfg, log),
		log:         log,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			response.BadRequest(c, "邮箱已被注册")
			return
		}
		h.log.Errorw("register failed", "error", err)
		response.InternalError(c, "注册失败")
		return
	}

	response.Created(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Unauthorized(c, "邮箱或密码错误")
			return
		}
		if errors.Is(err, services.ErrUserDisabled) {
			response.Forbidden(c, "账号已禁用")
			return
		}
		h.log.Errorw("login failed", "error", err)
		response.InternalError(c, "登录失败")
		return
	}

	response.Success(c, resp)
}
