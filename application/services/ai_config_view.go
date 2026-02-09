package services

import "github.com/drama-generator/backend/domain/models"

// AIServiceConfigView is safe to return to clients (API key is never exposed).
type AIServiceConfigView struct {
	ID            uint              `json:"id"`
	UserID        uint              `json:"user_id"`
	ServiceType   string            `json:"service_type"`
	Provider      string            `json:"provider"`
	Name          string            `json:"name"`
	BaseURL       string            `json:"base_url"`
	APIKey        string            `json:"api_key"`
	APIKeySet     bool              `json:"api_key_set"`
	Model         models.ModelField `json:"model"`
	Endpoint      string            `json:"endpoint"`
	QueryEndpoint string            `json:"query_endpoint"`
	Priority      int               `json:"priority"`
	IsDefault     bool              `json:"is_default"`
	IsActive      bool              `json:"is_active"`
	Settings      string            `json:"settings"`
	CreatedAt     string            `json:"created_at,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
}

func ToAIServiceConfigView(cfg models.AIServiceConfig) AIServiceConfigView {
	return AIServiceConfigView{
		ID:            cfg.ID,
		UserID:        cfg.UserID,
		ServiceType:   cfg.ServiceType,
		Provider:      cfg.Provider,
		Name:          cfg.Name,
		BaseURL:       cfg.BaseURL,
		APIKey:        "", // never expose
		APIKeySet:     cfg.APIKey != "",
		Model:         cfg.Model,
		Endpoint:      cfg.Endpoint,
		QueryEndpoint: cfg.QueryEndpoint,
		Priority:      cfg.Priority,
		IsDefault:     cfg.IsDefault,
		IsActive:      cfg.IsActive,
		Settings:      cfg.Settings,
		CreatedAt:     cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToAIServiceConfigViews(cfgs []models.AIServiceConfig) []AIServiceConfigView {
	out := make([]AIServiceConfigView, 0, len(cfgs))
	for _, c := range cfgs {
		out = append(out, ToAIServiceConfigView(c))
	}
	return out
}
