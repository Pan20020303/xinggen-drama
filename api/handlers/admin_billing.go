package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
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
	adminID, err := tenant.GetUserID(c)
	if err != nil {
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

func (h *AdminBillingHandler) GetTokenStats(c *gin.Context) {
	var serviceTypePtr *string
	serviceType := c.Query("service_type")
	if serviceType != "" {
		serviceTypePtr = &serviceType
	}

	var startDatePtr *time.Time
	if raw := c.Query("start_date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			response.BadRequest(c, "invalid start_date, expected YYYY-MM-DD")
			return
		}
		startDatePtr = &parsed
	}

	var endDatePtr *time.Time
	if raw := c.Query("end_date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			response.BadRequest(c, "invalid end_date, expected YYYY-MM-DD")
			return
		}
		parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endDatePtr = &parsed
	}

	items, summary, err := h.billingService.GetTokenStats(serviceTypePtr, startDatePtr, endDatePtr)
	if err != nil {
		h.log.Errorw("failed to get token stats", "error", err)
		response.InternalError(c, "查询 token 统计失败")
		return
	}

	response.Success(c, gin.H{
		"items":    items,
		"summary":  summary,
		"filters": gin.H{
			"service_type": serviceType,
			"start_date":   c.Query("start_date"),
			"end_date":     c.Query("end_date"),
		},
	})
}
