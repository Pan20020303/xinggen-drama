package routes

import (
	handlers "github.com/drama-generator/backend/api/handlers"
	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, log *logger.Logger, localStorage interface{}) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middlewares2.LoggerMiddleware(log))
	r.Use(middlewares2.CORSMiddleware(cfg.Server.CORSOrigins))

	// 静态文件服务（用户上传的文件）
	r.Static("/static", cfg.Storage.LocalPath)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	deps, err := buildAppDependencies(cfg, db, log, localStorage)
	if err != nil {
		log.Fatalw("Failed to build app dependencies", "error", err)
	}

	api := r.Group("/api/v1")
	{
		api.Use(middlewares2.RateLimitMiddleware())

		auth := api.Group("/auth")
		{
			loginRateLimit := middlewares2.LoginRateLimitMiddleware()
			auth.POST("/register", deps.authHandler.Register)
			auth.POST("/login", loginRateLimit, deps.authHandler.Login)
		}

		adminAuth := api.Group("/admin/auth")
		{
			loginRateLimit := middlewares2.LoginRateLimitMiddleware()
			adminAuth.POST("/login", loginRateLimit, deps.adminAuthHandler.Login)
		}

		secured := api.Group("")
		secured.Use(middlewares2.AuthMiddleware(deps.authService))
		{
			secured.GET("/auth/me", deps.authHandler.Me)
			secured.POST("/auth/refresh", deps.authHandler.RefreshToken)
			secured.PUT("/auth/password", deps.authHandler.ChangePassword)
			secured.GET("/billing/pricing", deps.billingPricingHandler.GetPricing)
		}

		adminSecured := api.Group("/admin")
		adminSecured.Use(middlewares2.AdminAuthMiddleware(deps.authService))
		{
			adminUsers := adminSecured.Group("/users")
			{
				adminUsers.GET("", deps.adminUserHandler.ListUsers)
				adminUsers.PATCH("/:id/status", deps.adminUserHandler.UpdateUserStatus)
				adminUsers.PATCH("/:id/role", deps.adminUserHandler.UpdateUserRole)
			}

			adminBilling := adminSecured.Group("/billing")
			{
				adminBilling.POST("/recharge", deps.adminBillingHandler.Recharge)
				adminBilling.GET("/transactions", deps.adminBillingHandler.ListTransactions)
				adminBilling.GET("/token-stats", deps.adminBillingHandler.GetTokenStats)
			}

			adminAI := adminSecured.Group("/ai-configs")
			{
				adminAI.GET("", deps.adminAIConfigHandler.ListConfigs)
				adminAI.POST("", deps.adminAIConfigHandler.CreateConfig)
				adminAI.PUT("/:id", deps.adminAIConfigHandler.UpdateConfig)
				adminAI.DELETE("/:id", deps.adminAIConfigHandler.DeleteConfig)
				adminAI.POST("/test", deps.adminAIConfigHandler.TestConnection)
			}
		}

		dramas := secured.Group("/dramas")
		{
			dramas.GET("", deps.dramaHandler.ListDramas)
			dramas.POST("", deps.dramaHandler.CreateDrama)
			dramas.GET("/stats", deps.dramaHandler.GetDramaStats) // 统计接口放在/:id之前
			dramas.GET("/:id", deps.dramaHandler.GetDrama)
			dramas.PUT("/:id", deps.dramaHandler.UpdateDrama)
			dramas.DELETE("/:id", deps.dramaHandler.DeleteDrama)

			dramas.PUT("/:id/outline", deps.dramaHandler.SaveOutline)
			dramas.GET("/:id/characters", deps.dramaHandler.GetCharacters)
			dramas.PUT("/:id/characters", deps.dramaHandler.SaveCharacters)
			dramas.PUT("/:id/episodes", deps.dramaHandler.SaveEpisodes)
			dramas.PUT("/:id/progress", deps.dramaHandler.SaveProgress)
			dramas.GET("/:id/props", deps.propHandler.ListProps) // Added prop list route
		}

		generation := secured.Group("/generation")
		{
			generation.POST("/characters", deps.scriptGenHandler.GenerateCharacters)
			generation.POST("/script/polish", deps.scriptGenHandler.PolishScriptText)
		}

		// 角色库路由
		characterLibrary := secured.Group("/character-library")
		{
			characterLibrary.GET("", deps.characterLibraryHandler.ListLibraryItems)
			characterLibrary.POST("", deps.characterLibraryHandler.CreateLibraryItem)
			characterLibrary.GET("/:id", deps.characterLibraryHandler.GetLibraryItem)
			characterLibrary.DELETE("/:id", deps.characterLibraryHandler.DeleteLibraryItem)
		}

		// 角色图片相关路由
		characters := secured.Group("/characters")
		{
			characters.PUT("/:id", deps.characterLibraryHandler.UpdateCharacter)
			characters.DELETE("/:id", deps.characterLibraryHandler.DeleteCharacter)
			characters.POST("/batch-generate-images", deps.characterLibraryHandler.BatchGenerateCharacterImages)
			characters.POST("/:id/generate-image", deps.characterLibraryHandler.GenerateCharacterImage)
			characters.POST("/:id/upload-image", deps.uploadHandler.UploadCharacterImage)
			characters.PUT("/:id/image", deps.characterLibraryHandler.UploadCharacterImage)
			characters.PUT("/:id/image-from-library", deps.characterLibraryHandler.ApplyLibraryItemToCharacter)
			characters.POST("/:id/add-to-library", deps.characterLibraryHandler.AddCharacterToLibrary)
		}

		props := secured.Group("/props")
		{
			props.POST("", deps.propHandler.CreateProp)
			props.PUT("/:id", deps.propHandler.UpdateProp)
			props.DELETE("/:id", deps.propHandler.DeleteProp)
			props.POST("/:id/generate", deps.propHandler.GenerateImage)
		}

		// 文件上传路由
		upload := secured.Group("/upload")
		{
			upload.POST("/image", deps.uploadHandler.UploadImage)
		}

		// 分镜头路由
		episodes := secured.Group("/episodes")
		{
			// 分镜头
			episodes.POST("/:episode_id/storyboards", deps.storyboardHandler.GenerateStoryboard)
			episodes.POST("/:episode_id/polish-script", deps.scriptGenHandler.PolishEpisodeScript)
			episodes.POST("/:episode_id/props/extract", deps.propHandler.ExtractProps)
			episodes.POST("/:episode_id/characters/extract", deps.characterLibraryHandler.ExtractCharacters)
			episodes.GET("/:episode_id/storyboards", deps.sceneHandler.GetStoryboardsForEpisode)
			episodes.POST("/:episode_id/finalize", deps.dramaHandler.FinalizeEpisode)
			episodes.GET("/:episode_id/download", deps.dramaHandler.DownloadEpisodeVideo)
		}

		// 任务路由
		tasks := secured.Group("/tasks")
		{
			tasks.GET("/:task_id", deps.taskHandler.GetTaskStatus)
			tasks.GET("", deps.taskHandler.GetResourceTasks)
		}

		// 场景路由
		scenes := secured.Group("/scenes")
		{
			scenes.PUT("/:scene_id", deps.sceneHandler.UpdateScene)
			scenes.PUT("/:scene_id/prompt", deps.sceneHandler.UpdateScenePrompt)
			scenes.DELETE("/:scene_id", deps.sceneHandler.DeleteScene)

			scenes.POST("/generate-image", deps.sceneHandler.GenerateSceneImage)
			scenes.POST("", deps.sceneHandler.CreateScene)
		}

		images := secured.Group("/images")
		{
			images.GET("", deps.imageGenHandler.ListImageGenerations)
			images.POST("", deps.imageGenHandler.GenerateImage)
			images.GET("/:id", deps.imageGenHandler.GetImageGeneration)
			images.DELETE("/:id", deps.imageGenHandler.DeleteImageGeneration)
			images.POST("/scene/:scene_id", deps.imageGenHandler.GenerateImagesForScene)
			images.POST("/upload", deps.imageGenHandler.UploadImage)
			images.GET("/episode/:episode_id/backgrounds", deps.imageGenHandler.GetBackgroundsForEpisode)
			images.POST("/episode/:episode_id/backgrounds/extract", deps.imageGenHandler.ExtractBackgroundsForEpisode)
			images.POST("/episode/:episode_id/batch", deps.imageGenHandler.BatchGenerateForEpisode)
		}

		videos := secured.Group("/videos")
		{
			videos.GET("", deps.videoGenHandler.ListVideoGenerations)
			videos.POST("", deps.videoGenHandler.GenerateVideo)
			videos.GET("/:id", deps.videoGenHandler.GetVideoGeneration)
			videos.DELETE("/:id", deps.videoGenHandler.DeleteVideoGeneration)
			videos.POST("/image/:image_gen_id", deps.videoGenHandler.GenerateVideoFromImage)
			videos.POST("/episode/:episode_id/batch", deps.videoGenHandler.BatchGenerateForEpisode)
		}

		videoMerges := secured.Group("/video-merges")
		{
			videoMerges.GET("", deps.videoMergeHandler.ListMerges)
			videoMerges.POST("", deps.videoMergeHandler.MergeVideos)
			videoMerges.GET("/:merge_id", deps.videoMergeHandler.GetMerge)
			videoMerges.DELETE("/:merge_id", deps.videoMergeHandler.DeleteMerge)
		}

		assets := secured.Group("/assets")
		{
			assets.GET("", deps.assetHandler.ListAssets)
			assets.POST("", deps.assetHandler.CreateAsset)
			assets.GET("/:id", deps.assetHandler.GetAsset)
			assets.PUT("/:id", deps.assetHandler.UpdateAsset)
			assets.DELETE("/:id", deps.assetHandler.DeleteAsset)
			assets.POST("/import/image/:image_gen_id", deps.assetHandler.ImportFromImageGen)
			assets.POST("/import/video/:video_gen_id", deps.assetHandler.ImportFromVideoGen)
		}

		storyboards := secured.Group("/storyboards")
		{
			storyboards.GET("/episode/:episode_id/generate", deps.storyboardHandler.GenerateStoryboard)
			storyboards.POST("", deps.storyboardHandler.CreateStoryboard)
			storyboards.PUT("/:id", deps.storyboardHandler.UpdateStoryboard)
			storyboards.DELETE("/:id", deps.storyboardHandler.DeleteStoryboard)
			storyboards.POST("/:id/props", deps.propHandler.AssociateProps)
			storyboards.POST("/:id/frame-prompt", deps.framePromptHandler.GenerateFramePrompt)
			storyboards.GET("/:id/frame-prompts", handlers.GetStoryboardFramePrompts(db, log))
			storyboards.POST("/:id/optimize-video-prompt", deps.storyboardHandler.OptimizeVideoPrompt)
		}

		audio := secured.Group("/audio")
		{
			audio.POST("/extract", deps.audioExtractionHandler.ExtractAudio)
			audio.POST("/extract/batch", deps.audioExtractionHandler.BatchExtractAudio)
		}

		settings := secured.Group("/settings")
		{
			settings.GET("/language", deps.settingsHandler.GetLanguage)
			settings.PUT("/language", deps.settingsHandler.UpdateLanguage)
		}
	}

	// 前端静态文件服务（放在API路由之后，避免冲突）
	// 服务前端构建产物
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// NoRoute处理：对于所有未匹配的路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API路径，返回404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// SPA fallback - 返回index.html
		c.File("./web/dist/index.html")
	})

	return r
}
