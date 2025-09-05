package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/PlakarKorp/plakar/restful/config"
	"github.com/PlakarKorp/plakar/restful/handlers"
	"github.com/PlakarKorp/plakar/restful/middleware/auth"
	"github.com/PlakarKorp/plakar/restful/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	s3Storage, err := storage.NewS3Storage(cfg.AWS)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Custom middleware
	e.Use(auth.TokenAuthMiddleware(cfg.Auth.Token))

	// Initialize handlers
	h := handlers.New(s3Storage, cfg)

	// API routes
	api := e.Group("/api")

	// System routes
	api.GET("/info", h.GetInfo)

	// Repository routes
	repo := api.Group("/repository")
	repo.POST("/create", h.CreateRepository)
	repo.GET("/info", h.GetRepositoryInfo)
	repo.GET("/snapshots", h.ListSnapshots)
	repo.GET("/locate-pathname", h.LocatePathname)
	repo.GET("/importer-types", h.GetImporterTypes)
	repo.GET("/states", h.GetRepositoryStates)
	repo.GET("/state/:state", h.GetRepositoryState)
	repo.POST("/maintenance", h.RunMaintenance)
	repo.POST("/prune", h.PruneRepository)
	repo.POST("/sync", h.SyncRepository)

	// Snapshot routes
	snapshots := api.Group("/snapshots")
	snapshots.POST("/create", h.CreateSnapshot)
	snapshots.DELETE("/remove", h.RemoveSnapshots)

	snapshot := api.Group("/snapshot")
	snapshot.GET("/:snapshot", h.GetSnapshotHeader)
	snapshot.POST("/:snapshot/restore", h.RestoreSnapshot)
	snapshot.POST("/:snapshot/check", h.CheckSnapshot)
	snapshot.GET("/:snapshot/diff/:target_snapshot", h.DiffSnapshots)
	snapshot.POST("/:snapshot/mount", h.MountSnapshot)
	snapshot.POST("/:snapshot/unmount", h.UnmountSnapshot)

	// VFS routes
	vfs := snapshot.Group("/:snapshot/vfs")
	vfs.GET("/*", h.BrowseVFS)
	vfs.GET("/children/*", h.ListVFSChildren)
	vfs.GET("/chunks/*", h.GetVFSChunks)
	vfs.GET("/search/*", h.SearchVFS)
	vfs.GET("/errors/*", h.GetVFSErrors)
	vfs.POST("/downloader/*", h.CreateDownloadPackage)
	vfs.GET("/downloader-sign-url/:id", h.GetSignedDownloadURL)

	// File operations
	files := api.Group("/files")
	files.GET("/cat/*", h.GetFileContent)
	files.GET("/digest/*", h.GetFileDigest)

	// Reader routes
	reader := api.Group("/snapshot/reader")
	reader.GET("/*", h.ReadFile)
	reader.POST("/reader-sign-url/*", h.CreateSignedURL)

	// Search routes
	search := api.Group("/search")
	search.GET("/locate", h.LocateFiles)

	// Authentication routes (no auth required)
	auth := api.Group("/authentication")
	auth.POST("/login/github", h.LoginGitHub, auth.NoAuthMiddleware())
	auth.POST("/login/email", h.LoginEmail, auth.NoAuthMiddleware())
	auth.POST("/logout", h.Logout)

	// Integration routes
	integrations := api.Group("/integrations")
	integrations.POST("/install", h.InstallIntegration)
	integrations.DELETE("/:id", h.UninstallIntegration)

	// Scheduler routes (no auth required for status)
	scheduler := api.Group("/scheduler")
	scheduler.POST("/start", h.StartScheduler, auth.NoAuthMiddleware())
	scheduler.POST("/stop", h.StopScheduler, auth.NoAuthMiddleware())
	scheduler.GET("/status", h.GetSchedulerStatus, auth.NoAuthMiddleware())

	// Proxy routes
	proxy := api.Group("/proxy")
	proxy.Any("/v1/*", h.ProxyRequest)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"version": "1.0.2",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Start server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Starting server on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}