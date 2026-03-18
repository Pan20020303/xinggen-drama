package routes

import (
	"fmt"

	handlers "github.com/drama-generator/backend/api/handlers"
	services "github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/infrastructure/persistence"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type appDependencies struct {
	authService             *services.AuthService
	taskService             *services.TaskService
	aiService               *services.AIService
	adminAuditService       *services.AdminAuditService
	adminUserService        *services.AdminUserService
	adminBillingService     *services.AdminBillingService
	billingService          *services.BillingService
	transferService         *services.ResourceTransferService
	dramaService            *services.DramaService
	imageGenerationService  *services.ImageGenerationService
	scriptGenerationService *services.ScriptGenerationService
	storyboardService       *services.StoryboardService
	videoGenerationService  *services.VideoGenerationService
	videoMergeService       *services.VideoMergeService
	assetService            *services.AssetService
	audioExtractionService  *services.AudioExtractionService
	propService             *services.PropService
	authHandler             *handlers.AuthHandler
	adminAuthHandler        *handlers.AdminAuthHandler
	adminUserHandler        *handlers.AdminUserHandler
	adminBillingHandler     *handlers.AdminBillingHandler
	billingPricingHandler   *handlers.BillingPricingHandler
	billingTransactionsHandler *handlers.BillingTransactionsHandler
	adminAIConfigHandler    *handlers.AdminAIConfigHandler
	dramaHandler            *handlers.DramaHandler
	scriptGenHandler        *handlers.ScriptGenerationHandler
	imageGenHandler         *handlers.ImageGenerationHandler
	videoGenHandler         *handlers.VideoGenerationHandler
	videoMergeHandler       *handlers.VideoMergeHandler
	assetHandler            *handlers.AssetHandler
	characterLibraryHandler *handlers.CharacterLibraryHandler
	uploadHandler           *handlers.UploadHandler
	storyboardHandler       *handlers.StoryboardHandler
	sceneHandler            *handlers.SceneHandler
	taskHandler             *handlers.TaskHandler
	framePromptHandler      *handlers.FramePromptHandler
	audioExtractionHandler  *handlers.AudioExtractionHandler
	settingsHandler         *handlers.SettingsHandler
	propHandler             *handlers.PropHandler
}

func buildAppDependencies(cfg *config.Config, db *gorm.DB, log *logger.Logger, localStorage any) (*appDependencies, error) {
	localStoragePtr, err := requireLocalStorage(localStorage)
	if err != nil {
		return nil, err
	}

	aiService := services.NewAIService(db, log)
	transferService := services.NewResourceTransferService(db, log)
	promptI18n := services.NewPromptI18n(cfg)
	userRepo := persistence.NewGormUserRepository(db)
	authService := services.NewAuthService(userRepo, cfg, log)
	taskService := services.NewTaskService(db, log)
	adminAuditService := services.NewAdminAuditService(db)
	adminUserService := services.NewAdminUserService(db, log, adminAuditService)
	adminBillingService := services.NewAdminBillingService(db, log, adminAuditService)
	billingService := services.NewBillingService(db, cfg, log)
	dramaService := services.NewDramaService(db, cfg, log)
	characterLibraryService := services.NewCharacterLibraryService(db, log, cfg)
	imageGenService := services.NewImageGenerationService(db, cfg, transferService, localStoragePtr, log)
	sceneService := services.NewStoryboardCompositionService(db, log, imageGenService)
	framePromptService := services.NewFramePromptService(db, cfg, log)
	scriptGenerationService := services.NewScriptGenerationService(db, cfg, log)
	storyboardService := services.NewStoryboardService(db, cfg, log)
	videoGenerationService := services.NewVideoGenerationService(db, cfg, transferService, localStoragePtr, aiService, log, promptI18n)
	videoMergeService := services.NewVideoMergeService(db, nil, cfg.Storage.LocalPath, cfg.Storage.BaseURL, log)
	assetService := services.NewAssetService(db, log)
	audioExtractionService := services.NewAudioExtractionService(log)
	propService := services.NewPropService(db, aiService, taskService, imageGenService, log, cfg)
	uploadService, err := services.NewUploadService(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload service: %w", err)
	}

	uploadHandler := handlers.NewUploadHandler(uploadService, characterLibraryService, log)

	return &appDependencies{
		authService:             authService,
		taskService:             taskService,
		aiService:               aiService,
		adminAuditService:       adminAuditService,
		adminUserService:        adminUserService,
		adminBillingService:     adminBillingService,
		billingService:          billingService,
		transferService:         transferService,
		dramaService:            dramaService,
		imageGenerationService:  imageGenService,
		scriptGenerationService: scriptGenerationService,
		storyboardService:       storyboardService,
		videoGenerationService:  videoGenerationService,
		videoMergeService:       videoMergeService,
		assetService:            assetService,
		audioExtractionService:  audioExtractionService,
		propService:             propService,
		authHandler:             handlers.NewAuthHandler(authService, log),
		adminAuthHandler:        handlers.NewAdminAuthHandler(authService, log),
		adminUserHandler:        handlers.NewAdminUserHandler(adminUserService, log),
		adminBillingHandler:     handlers.NewAdminBillingHandler(adminBillingService, log),
		billingPricingHandler:   handlers.NewBillingPricingHandler(aiService, log),
		billingTransactionsHandler: handlers.NewBillingTransactionsHandler(billingService, log),
		adminAIConfigHandler:    handlers.NewAdminAIConfigHandler(aiService, log),
		dramaHandler:            handlers.NewDramaHandler(db, dramaService, videoMergeService, log),
		scriptGenHandler:        handlers.NewScriptGenerationHandler(scriptGenerationService, taskService, log),
		imageGenHandler:         handlers.NewImageGenerationHandler(db, cfg, log, imageGenService, taskService),
		videoGenHandler:         handlers.NewVideoGenerationHandler(videoGenerationService, log),
		videoMergeHandler:       handlers.NewVideoMergeHandler(videoMergeService, log),
		assetHandler:            handlers.NewAssetHandler(assetService, log),
		characterLibraryHandler: handlers.NewCharacterLibraryHandler(characterLibraryService, imageGenService, log),
		uploadHandler:           uploadHandler,
		storyboardHandler:       handlers.NewStoryboardHandler(storyboardService, taskService, log),
		sceneHandler:            handlers.NewSceneHandler(sceneService, log),
		taskHandler:             handlers.NewTaskHandler(taskService, log),
		framePromptHandler:      handlers.NewFramePromptHandler(framePromptService, log),
		audioExtractionHandler:  handlers.NewAudioExtractionHandler(audioExtractionService, log, cfg.Storage.LocalPath),
		settingsHandler:         handlers.NewSettingsHandler(cfg, log),
		propHandler:             handlers.NewPropHandler(propService, log),
	}, nil
}

func requireLocalStorage(localStorage any) (*storage.LocalStorage, error) {
	if localStorage == nil {
		return nil, fmt.Errorf("local storage is required")
	}

	localStoragePtr, ok := localStorage.(*storage.LocalStorage)
	if !ok || localStoragePtr == nil {
		return nil, fmt.Errorf("invalid local storage dependency")
	}

	return localStoragePtr, nil
}
