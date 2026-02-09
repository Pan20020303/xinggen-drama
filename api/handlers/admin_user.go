package handlers

import (
	"errors"
	"strconv"

	"github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminUserHandler struct {
	userService *services.AdminUserService
	log         *logger.Logger
}

type AdminUpdateStatusRequest struct {
	Status models.UserStatus `json:"status" binding:"required"`
}

type AdminUpdateRoleRequest struct {
	Role models.UserRole `json:"role" binding:"required"`
}

func NewAdminUserHandler(db *gorm.DB, log *logger.Logger) *AdminUserHandler {
	auditSvc := services.NewAdminAuditService(db)
	return &AdminUserHandler{
		userService: services.NewAdminUserService(db, log, auditSvc),
		log:         log,
	}
}

func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.ListUsers(page, pageSize)
	if err != nil {
		h.log.Errorw("failed to list users", "error", err)
		response.InternalError(c, "查询用户失败")
		return
	}
	response.SuccessWithPagination(c, users, total, page, pageSize)
}

func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	adminID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "invalid admin context")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req AdminUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserStatus(adminID, uint(userID), req.Status, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		if err.Error() == "invalid user status" {
			response.BadRequest(c, err.Error())
			return
		}
		h.log.Errorw("failed to update user status", "error", err, "admin_id", adminID, "user_id", userID)
		response.InternalError(c, "更新用户状态失败")
		return
	}

	response.Success(c, user)
}

func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	adminID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "invalid admin context")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req AdminUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserRole(adminID, uint(userID), req.Role, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		if err.Error() == "invalid user role" {
			response.BadRequest(c, err.Error())
			return
		}
		h.log.Errorw("failed to update user role", "error", err, "admin_id", adminID, "user_id", userID)
		response.InternalError(c, "更新用户角色失败")
		return
	}

	response.Success(c, user)
}
