package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andydunstall/piko/agent/config"
	"github.com/andydunstall/piko/agent/reverseproxy"
	"github.com/andydunstall/piko/client"
	"github.com/andydunstall/piko/pkg/log"
)

// PikoService manages the piko reverse proxy service
type PikoService struct {
	config               PikoConfig
	ctx                  context.Context
	insecureSkipVerify   bool
}

// PikoConfig holds the configuration for piko service
type PikoConfig struct {
	RemoteURL    string
	EndpointID   string
	LocalAddr    string
	Timeout      time.Duration
	GracePeriod  time.Duration
	AccessLog    bool
}

// NewPikoService creates a new piko service
func NewPikoService(config PikoConfig, ctx context.Context, insecureSkipVerify bool) *PikoService {
	return &PikoService{
		config:             config,
		ctx:                ctx,
		insecureSkipVerify: insecureSkipVerify,
	}
}

// Start starts the piko service
func (ps *PikoService) Start() error {
	fmt.Printf("Starting piko...")

	// Normalize remote URL
	remote := ps.config.RemoteURL
	if !strings.HasPrefix(remote, "http") {
		remote = fmt.Sprintf("http://%s", remote)
	}

	// Create piko configuration
	conf := &config.Config{
		Connect: config.ConnectConfig{
			URL:     remote,
			Timeout: ps.config.Timeout,
		},
		Listeners: []config.ListenerConfig{
			{
				EndpointID: ps.config.EndpointID,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       ps.config.LocalAddr,
				AccessLog:  ps.config.AccessLog,
				Timeout:    30 * time.Second,
				TLS:        config.TLSConfig{},
			},
		},
		Log: log.Config{
			Level:      "info",
			Subsystems: []string{},
		},
		GracePeriod: ps.config.GracePeriod,
	}

	// Create logger
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Validate configuration
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko configuration validation failed: %w", err)
	}

	// Parse connection URL
	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("failed to parse connection URL: %w", err)
	}

	// Create TLS configuration
	var tlsConfig *tls.Config
	if ps.insecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		fmt.Printf(" (HTTPS certificate verification skipped)")
	}

	// Create upstream client
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: tlsConfig,
		Logger:    logger.WithSubsystem("client"),
	}

	// Create connections for each listener
	for _, listenerConfig := range conf.Listeners {
		ln, err := upstream.Listen(ps.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("failed to listen on endpoint %s: %w", listenerConfig.EndpointID, err)
		}

		// Create HTTP proxy server
		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("failed to create HTTP proxy server")
		}

		// Start proxy server
		go func() {
			if err := server.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("proxy server runtime error: %v\n", err)
			}
		}()
	}

	fmt.Print(" done\n")
	return nil
}
