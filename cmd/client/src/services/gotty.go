package services

import (
	"context"
	"fmt"

	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

// GottyService manages the gotty web terminal service
type GottyService struct {
	config      GottyConfig
	ctx         context.Context
}

// GottyConfig holds the configuration for gotty service
type GottyConfig struct {
	Address         string
	Port            int
	Path            string
	PermitWrite     bool
	TitleFormat     string
	WSOrigin        string
	EnableBasicAuth bool
	Credential      string
	Command         string
	Args            []string
}

// NewGottyService creates a new gotty service
func NewGottyService(config GottyConfig, ctx context.Context) *GottyService {
	return &GottyService{
		config: config,
		ctx:    ctx,
	}
}

// Start starts the gotty service
func (gs *GottyService) Start() error {
	fmt.Print("Starting gotty...")

	// Create gotty server options
	options := &server.Options{
		Address:         gs.config.Address,
		Port:            fmt.Sprintf("%d", gs.config.Port),
		Path:            gs.config.Path,
		PermitWrite:     gs.config.PermitWrite,
		TitleFormat:     gs.config.TitleFormat,
		WSOrigin:        gs.config.WSOrigin,
		EnableBasicAuth: gs.config.EnableBasicAuth,
	}

	if gs.config.EnableBasicAuth && gs.config.Credential != "" {
		options.Credential = gs.config.Credential
	}

	// Create local command factory
	backendOptions := &localcommand.Options{}
	factory, err := localcommand.NewFactory(gs.config.Command, gs.config.Args, backendOptions)
	if err != nil {
		return fmt.Errorf("failed to create gotty factory: %w", err)
	}

	// Create gotty server
	srv, err := server.New(factory, options)
	if err != nil {
		return fmt.Errorf("failed to create gotty server: %w", err)
	}

	// Start gotty server in a separate goroutine
	go func() {
		err := srv.Run(gs.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("gotty server runtime error: %v\n", err)
		}
	}()

	fmt.Print(" done\n")
	return nil
}
