package services

import (
	"context"
	"fmt"
	"os"

	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

type GottyService struct {
	config   GottyConfig
	ctx      context.Context
	notifier *server.Notifier
}

type GottyConfig struct {
	Address         string
	Port            int
	Path            string
	SessionName     string
	PermitWrite     bool
	TitleFormat     string
	WSOrigin        string
	Auth            bool
	AuthName        string
	Password        string
	EnableBasicAuth bool
	EnableNotify    bool
	NotifyWebhook   string
	StaticIndex     string
	AttachPort      string
	TitleVariables  map[string]interface{}
	Command         string
	Args            []string
}

func NewGottyService(config GottyConfig, ctx context.Context) *GottyService {
	return &GottyService{
		config: config,
		ctx:    ctx,
	}
}

func (gs *GottyService) Start() error {
	fmt.Print("Starting gotty...")

	options := &server.Options{
		Address:        gs.config.Address,
		Port:           fmt.Sprintf("%d", gs.config.Port),
		Path:           gs.config.Path,
		SessionName:    gs.config.SessionName,
		PermitWrite:    gs.config.PermitWrite,
		TitleFormat:    gs.config.TitleFormat,
		WSOrigin:       gs.config.WSOrigin,
		Auth:           gs.config.Auth,
		EnableNotify:   gs.config.EnableNotify,
		NotifyWebhook:  gs.config.NotifyWebhook,
		StaticIndex:    gs.config.StaticIndex,
		AttachPort:     gs.config.AttachPort,
		TitleVariables: gs.config.TitleVariables,
	}

	if gs.config.Auth {
		options.AuthName = gs.config.AuthName
		options.Password = gs.config.Password
		options.EnableBasicAuth = true
	}

	notifier := server.NewNotifier(gs.config.NotifyWebhook)
	notifier.Start(gs.config.EnableNotify, "", gs.config.SessionName)

	backendOptions := &localcommand.Options{}
	if prefix := notifier.PathPrefix(); prefix != "" {
		backendOptions.EnvExtra = map[string]string{
			"PATH": prefix + os.Getenv("PATH"),
		}
	}

	var factory *localcommand.Factory
	var err error

	if gs.config.Command != "" {
		factory, err = localcommand.NewFactory(gs.config.Command, gs.config.Args, backendOptions)
	} else {
		factory, err = localcommand.NewFactory(gs.getShell(), []string{}, backendOptions)
	}

	if err != nil {
		return fmt.Errorf("failed to create gotty factory: %w", err)
	}

	srv, err := server.NewWithNotifier(factory, options, notifier)
	if err != nil {
		return fmt.Errorf("failed to create gotty server: %w", err)
	}

	go func() {
		err := srv.Run(gs.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("gotty server runtime error: %v\n", err)
		}
	}()

	fmt.Print(" done\n")
	return nil
}

func (gs *GottyService) getShell() string {
	return "bash"
}
