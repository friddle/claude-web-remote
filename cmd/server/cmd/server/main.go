package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clauded-server/config"
	"clauded-server/handlers"
	"clauded-server/notification"
	"clauded-server/proxy"
	"clauded-server/session"

	pikoserver "github.com/andydunstall/piko/server"
	pikoconfig "github.com/andydunstall/piko/server/config"
	pikolog "github.com/andydunstall/piko/pkg/log"
	"github.com/oklog/run"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create managers
	sessionMgr := session.NewManager()
	notificationSvc := notification.NewService()

	// Create proxy manager (piko proxy port is 8023)
	proxyMgr := proxy.NewManager(8023, cfg.PikoUpstreamPort)

	// Create HTTP handler
	handler := handlers.NewHandler(cfg, sessionMgr, notificationSvc, proxyMgr)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ListenPort),
		Handler: handler.SetupRoutes(),
	}

	var g run.Group

	// Create context for signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// Start Piko server as a Go library
	pikoSrv := startPikoServer(cfg)

	g.Add(func() error {
		stdlog.Printf("Starting piko server on upstream port %d, proxy port 8023\n", cfg.PikoUpstreamPort)
		if err := pikoSrv.Start(); err != nil {
			stdlog.Printf("❌ Piko server error: %v\n", err)
			return fmt.Errorf("piko server failed: %w", err)
		}
		stdlog.Println("✅ Piko server started successfully")
		
		// Wait for context cancellation
		<-ctx.Done()
		return nil
	}, func(error) {
		stdlog.Println("Stopping piko server...")
		pikoSrv.Shutdown()
		stdlog.Println("Piko server stopped")
	})

	// Wait for piko to start
	time.Sleep(2 * time.Second)

	// HTTP server
	g.Add(func() error {
		stdlog.Printf("Starting HTTP server on port %d\n", cfg.ListenPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
		return nil
	}, func(error) {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		httpServer.Shutdown(shutdownCtx)
	})

	// Notification service
	g.Add(func() error {
		notificationSvc.Start()
		<-ctx.Done()
		return nil
	}, func(error) {
		notificationSvc.Stop()
	})

	// Signal handling
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\nReceived shutdown signal, stopping...")
		cancel()
		return nil
	}, func(error) {
		cancel()
	})

	// Run all services
	if err := g.Run(); err != nil {
		stdlog.Fatalf("Failed to run services: %v", err)
	}

	stdlog.Println("Server stopped gracefully")
}

// startPikoServer starts piko server as a Go library
func startPikoServer(cfg *config.Config) *pikoserver.Server {
	// Build piko server configuration
	upstreamAddr := fmt.Sprintf(":%d", cfg.PikoUpstreamPort)
	proxyAddr := ":8023"

	// Create piko logger (use error level to reduce logs)
	logger, _ := pikolog.NewLogger("error", nil)

	// Get default config and customize it
	pikoCfg := pikoconfig.Default()
	pikoCfg.Cluster.NodeID = "clauded-server-1"
	pikoCfg.Cluster.JoinTimeout = 10 * time.Second
	pikoCfg.Cluster.AbortIfJoinFails = false  // Don't abort if cluster join fails
	pikoCfg.Cluster.Gossip.BindAddr = ":0"    // Disable gossip
	pikoCfg.Upstream.BindAddr = upstreamAddr
	pikoCfg.Proxy.BindAddr = proxyAddr
	pikoCfg.Admin.BindAddr = ":7070"
	pikoCfg.GracePeriod = 30 * time.Second

	// Validate config
	if err := pikoCfg.Validate(); err != nil {
		stdlog.Fatalf("❌ Invalid piko configuration: %v", err)
	}

	// Create piko server
	pikoSrv, err := pikoserver.NewServer(pikoCfg, logger)
	if err != nil {
		stdlog.Fatalf("❌ Failed to create piko server: %v", err)
	}

	return pikoSrv
}
