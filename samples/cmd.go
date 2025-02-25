package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nagara-stack/samples/config"
)

func startAgentService() {
	// Initialize logger
	logger := log.New(os.Stdout, "[ThirdPartyAgent] ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := config.LoadConfig("../config/")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Create Gin instance
	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))

	// init ProposalDB
	proposalDB, err := NewProposalDB("./data")
	if err != nil {
		logger.Fatalf("Failed to initialize proposal database: %v", err)
	}
	defer proposalDB.Close()

	// Create service instance@
	httpHandler := NewHTTPHandler(logger, proposalDB)

	// Initialize HAC node client and register service
	hacClient := NewHACNodeClient(cfg.AuthToken)
	agentUrl := cfg.AgentUrl
	baseURL := fmt.Sprintf("%s:%d/api", cfg.HTTPUrl, cfg.HTTPPort)

	hacClient.baseURL = baseURL
	if err := hacClient.Register(context.Background(), "ThirdPartyAgent", agentUrl, "Third-party integration agent"); err != nil {
		logger.Fatalf("Agent registration failed: %v", err)
	}

	// Register HTTP routes
	httpHandler.RegisterRoutes(router)

	// Create HTTP serve
	parsedURL, err := url.Parse(cfg.AgentUrl)
	if err != nil {
		logger.Fatalf("Unable to parse AgentUrl: %v", err)
	}

	serverAddr := parsedURL.Host
	if serverAddr == "" {
		logger.Fatalf("Unable to extract a valid hostname and port from AgentUrl: %s", cfg.AgentUrl)
	}
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Start service
	go func() {
		logger.Printf("Starting HTTP service on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Shutdown service
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exiting")
}

// Custom logging middleware
func loggingMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Printf("%s %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
