package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drama-generator/backend/api/routes"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logr := logger.NewLogger(cfg.App.Debug)
	defer logr.Sync()

	logr.Info("Starting Drama Generator API Server...")

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logr.Fatal("Failed to connect to database", "error", err)
	}
	logr.Info("Database connected successfully")

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(db); err != nil {
		logr.Fatal("Failed to migrate database", "error", err)
	}
	logr.Info("Database tables migrated successfully")

	// 启动时执行历史数据修复与一致性自检（兼容旧数据 storyboards.user_id=0）
	if report, err := database.BackfillStoryboardUserID(db); err != nil {
		logr.Warnw("Storyboard user ownership backfill failed", "error", err)
	} else {
		logr.Infow("Storyboard user ownership check",
			"backfilled_rows", report.BackfilledRows,
			"remaining_zero_rows", report.RemainingZeroRows,
			"mismatch_rows", report.MismatchRows,
			"orphan_rows", report.OrphanRows)
		if report.RemainingZeroRows > 0 || report.MismatchRows > 0 || report.OrphanRows > 0 {
			logr.Warnw("Storyboard ownership integrity warnings detected",
				"remaining_zero_rows", report.RemainingZeroRows,
				"mismatch_rows", report.MismatchRows,
				"orphan_rows", report.OrphanRows)
		}
	}

	// 初始化本地存储
	var localStorage *storage.LocalStorage
	if cfg.Storage.Type == "local" {
		localStorage, err = storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
		if err != nil {
			logr.Fatal("Failed to initialize local storage", "error", err)
		}
		logr.Info("Local storage initialized successfully", "path", cfg.Storage.LocalPath)
	}

	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router, shutdownBackground := routes.SetupRouter(cfg, db, logr, localStorage)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	go func() {
		// Bind early so we don't log "ready" if the port is already taken.
		ln, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			logr.Fatal("Failed to start server", "error", err)
		}

		logr.Infow("🚀 Server starting...",
			"port", cfg.Server.Port,
			"mode", gin.Mode())
		logr.Info("📍 Access URLs:")
		logr.Info(fmt.Sprintf("   Frontend:  http://localhost:%d", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   API:       http://localhost:%d/api/v1", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Health:    http://localhost:%d/health", cfg.Server.Port))
		logr.Info("📁 Static files:")
		logr.Info(fmt.Sprintf("   Uploads:   http://localhost:%d/static", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Assets:    http://localhost:%d/assets", cfg.Server.Port))
		logr.Info("✅ Server is ready!")

		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logr.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down server...")

	// 清理资源
	// CRITICAL FIX: Properly close database connection to prevent resource leaks
	// SQLite connections should be closed gracefully to avoid database lock issues
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logr.Warnw("Failed to close database connection", "error", err)
		} else {
			logr.Info("Database connection closed")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutdownBackground != nil {
		if err := shutdownBackground(ctx); err != nil {
			logr.Warnw("Failed to shutdown background services", "error", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatal("Server forced to shutdown", "error", err)
	}

	logr.Info("Server exited")
}
