package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/drama-generator/backend/pkg/tenant"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BillingPricingHandler struct {
	aiService *services.AIService
	log       *logger.Logger
}

func NewBillingPricingHandler(db *gorm.DB, log *logger.Logger) *BillingPricingHandler {
	return &BillingPricingHandler{
		aiService: services.NewAIService(db, log),
		log:       log,
	}
}

type ServicePricing struct {
	ServiceType string `json:"service_type"`
	ConfigID    uint   `json:"config_id"`
	Model       string `json:"model"`
	CreditCost  int    `json:"credit_cost"`
}

type PricingResponse struct {
	Defaults       []ServicePricing              `json:"defaults"`
	UserConfigs    []services.AIServiceConfigView `json:"user_configs"`
	PlatformConfigs []services.AIServiceConfigView `json:"platform_configs"`
}

func (h *BillingPricingHandler) GetPricing(c *gin.Context) {
	userID, err := tenant.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "用户未登录")
		return
	}

	serviceTypes := []string{"text", "image", "video"}
	defaults := make([]ServicePricing, 0, len(serviceTypes))
	for _, st := range serviceTypes {
		cfg, model, err := h.aiService.GetBillingConfig(st, "", userID)
		if err != nil {
			// If a service type isn't configured yet, just skip it.
			h.log.Warnw("No billing config for service type", "service_type", st, "error", err, "user_id", userID)
			continue
		}

		// Prefer platform pricing for the resolved model, even when user config is selected.
		priceCfg := cfg
		if model != "" {
			if pcfg, perr := h.aiService.GetConfigForModel(st, model); perr == nil {
				priceCfg = pcfg
			}
		}
		defaults = append(defaults, ServicePricing{
			ServiceType: st,
			ConfigID:    priceCfg.ID,
			Model:       model,
			CreditCost:  priceCfg.CreditCost,
		})
	}

	var userCfgs []services.AIServiceConfigView
	var platformCfgs []services.AIServiceConfigView
	for _, st := range serviceTypes {
		cfgs, err := h.aiService.ListConfigs(st, userID)
		if err == nil {
			userCfgs = append(userCfgs, services.ToAIServiceConfigViews(cfgs)...)
		}
		pcfgs, err := h.aiService.ListPlatformConfigs(st)
		if err == nil {
			platformCfgs = append(platformCfgs, services.ToAIServiceConfigViews(pcfgs)...)
		}
	}

	response.Success(c, PricingResponse{
		Defaults:        defaults,
		UserConfigs:     userCfgs,
		PlatformConfigs: platformCfgs,
	})
}
