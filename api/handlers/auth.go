package handlers

import (
	"errors"
	"strings"

	"github.com/drama-generator/backend/application/dto"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewAuthHandler(authService *services.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
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
	var req dto.LoginRequest
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

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token, ok := getBearerToken(c)
	if !ok {
		response.Unauthorized(c, "缺少登录凭证")
		return
	}

	resp, err := h.authService.RefreshToken(token)
	if err != nil {
		if errors.Is(err, services.ErrTokenExpired) {
			response.Unauthorized(c, "登录已过期，请重新登录")
			return
		}
		if errors.Is(err, services.ErrTokenRefreshTooEarly) {
			response.BadRequest(c, "当前 token 暂不需要刷新")
			return
		}
		if errors.Is(err, services.ErrUserDisabled) {
			response.Forbidden(c, "账号已禁用")
			return
		}
		h.log.Errorw("refresh token failed", "error", err)
		response.InternalError(c, "刷新登录状态失败")
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		if errors.Is(err, services.ErrInvalidOldPassword) {
			response.Unauthorized(c, "旧密码错误")
			return
		}
		h.log.Errorw("change password failed", "error", err, "user_id", userID)
		response.InternalError(c, "修改密码失败")
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", gin.H{"updated": true})
}

// Me returns the current authenticated user.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.log.Errorw("get me failed", "error", err, "user_id", userID)
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.Success(c, user)
}

func getBearerToken(c *gin.Context) (string, bool) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return "", false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}
