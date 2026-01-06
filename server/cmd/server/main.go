package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"clauded-server/config"
	"clauded-server/handlers"
	"clauded-server/notification"
	"clauded-server/proxy"
	"clauded-server/session"

	"github.com/oklog/run"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create managers
	sessionMgr := session.NewManager()
	notificationSvc := notification.NewService()

	// Create proxy manager (piko proxy port is 8023)
	proxyMgr := proxy.NewManager(8023)

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

	// Start piko server as a subprocess
	pikoCmd := startPikoServer(cfg)
	if pikoCmd != nil {
		g.Add(func() error {
			log.Printf("Starting piko server on upstream port %d, proxy port 8023\n", cfg.PikoUpstreamPort)
			if err := pikoCmd.Wait(); err != nil {
				return fmt.Errorf("piko server failed: %w", err)
			}
			return nil
		}, func(error) {
			log.Println("Stopping piko server...")
			if pikoCmd.Process != nil {
				pikoCmd.Process.Signal(syscall.SIGTERM)
			}
		})

		// Wait for piko to start
		time.Sleep(2 * time.Second)
	}

	// HTTP server
	g.Add(func() error {
		log.Printf("Starting HTTP server on port %d\n", cfg.ListenPort)
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
		log.Fatalf("Failed to run services: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// startPikoServer starts piko server as a subprocess
func startPikoServer(cfg *config.Config) *exec.Cmd {
	args := []string{
		"server",
		"--upstream.bind-addr", fmt.Sprintf(":%d", cfg.PikoUpstreamPort),
		"--proxy.bind-addr", ":8023",
	}

	// Add token if configured
	if cfg.PikoToken != "" {
		args = append(args, "--token", cfg.PikoToken)
	}

	cmd := exec.Command("piko", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start piko server: %v", err)
		log.Println("WARNING: Continuing without piko server - proxy functionality will not work")
		return nil
	}

	return cmd
}
