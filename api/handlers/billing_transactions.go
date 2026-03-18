package handlers

import (
	"strconv"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
	"github.com/gin-gonic/gin"
)

type BillingTransactionsHandler struct {
	billingService *services.BillingService
	log            *logger.Logger
}

func NewBillingTransactionsHandler(billingService *services.BillingService, log *logger.Logger) *BillingTransactionsHandler {
	return &BillingTransactionsHandler{
		billingService: billingService,
		log:            log,
	}
}

func (h *BillingTransactionsHandler) ListTransactions(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := h.billingService.ListTransactions(userID, page, pageSize)
	if err != nil {
		h.log.Errorw("failed to list user credit transactions", "error", err, "user_id", userID)
		response.InternalError(c, "查询积分流水失败")
		return
	}

	response.SuccessWithPagination(c, items, total, page, pageSize)
}
