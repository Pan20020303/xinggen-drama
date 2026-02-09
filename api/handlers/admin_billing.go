package handlers

import (
	"errors"
	"strconv"

	"github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminBillingHandler struct {
	billingService *services.AdminBillingService
	log            *logger.Logger
}

type AdminRechargeRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Amount int    `json:"amount" binding:"required"`
	Note   string `json:"note"`
}

func NewAdminBillingHandler(db *gorm.DB, log *logger.Logger) *AdminBillingHandler {
	auditSvc := services.NewAdminAuditService(db)
	return &AdminBillingHandler{
		billingService: services.NewAdminBillingService(db, log, auditSvc),
		log:            log,
	}
}

func (h *AdminBillingHandler) Recharge(c *gin.Context) {
	adminID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "invalid admin context")
		return
	}

	var req AdminRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, txn, err := h.billingService.RechargeUser(adminID, req.UserID, req.Amount, req.Note, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		if err.Error() == "recharge amount must be positive" {
			response.BadRequest(c, err.Error())
			return
		}
		h.log.Errorw("failed to recharge user", "error", err, "admin_id", adminID, "user_id", req.UserID)
		response.InternalError(c, "充值失败")
		return
	}

	response.Success(c, gin.H{
		"user":        user,
		"transaction": txn,
	})
}

func (h *AdminBillingHandler) ListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var userIDPtr *uint
	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid user_id")
			return
		}
		u := uint(userID)
		userIDPtr = &u
	}

	txns, total, err := h.billingService.ListCreditTransactions(userIDPtr, page, pageSize)
	if err != nil {
		h.log.Errorw("failed to list credit transactions", "error", err)
		response.InternalError(c, "查询流水失败")
		return
	}
	response.SuccessWithPagination(c, txns, total, page, pageSize)
}
