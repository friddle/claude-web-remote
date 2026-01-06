package main

import (
	"fmt"
	"os"

	"clauded-client/src"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := MakeMainCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func MakeMainCmd() *cobra.Command {
	var (
		name               string
		host               string
		session            string
		password           string
		flags              string
		envVars            []string
		autoExit           bool
		insecureSkipVerify bool
		skipInstall        bool
	)

	cmd := &cobra.Command{
		Use:   "clauded",
		Short: "Claude Code remote client - Expose Claude Code via web terminal",
		Long: `clauded is a command-line tool that exposes your local Claude Code terminal session
through gotty and piko services to a remote server, allowing you to access and use
Claude Code from anywhere via a web browser.

Examples:
  clauded --host=clauded.friddle.me
  clauded --host=clauded.friddle.me --session=my-session
  clauded --host=clauded.friddle.me --password=my-secret-pass
  clauded --host=clauded.friddle.me --flags='--model claude-opus-4'
  clauded --host=clauded.friddle.me -e ANTHROPIC_BASE_URL=https://api.custom.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check and install claude-code if needed
			if !skipInstall {
				installer := src.NewInstaller()
				if !installer.IsClaudeCodeInstalled() {
					fmt.Println("claude-code not found, starting automatic installation...")
					if err := installer.Install(); err != nil {
						return fmt.Errorf("failed to install claude-code: %w", err)
					}
				}
			}

			// Create configuration
			config := &src.Config{
				Name:               name,
				Host:               host,
				Session:            session,
				Password:           password,
				Flags:              flags,
				EnvVars:            envVars,
				AutoExit:           autoExit,
				InsecureSkipVerify: insecureSkipVerify,
			}

			// Validate configuration
			if err := config.Validate(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// Create service manager
			manager := src.NewServiceManager(config)

			// Start services (blocks until service stops)
			if err := manager.Start(); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}

			return nil
		},
	}

	// Add command line flags
	cmd.Flags().StringVar(&name, "name", "", "Client identifier name (deprecated, use --session)")
	cmd.Flags().StringVar(&host, "host", "", "Remote server host (e.g., clauded.friddle.me)")
	cmd.Flags().StringVar(&session, "session", "", "Session ID (auto-generated if not specified)")
	cmd.Flags().StringVar(&password, "password", "", "Password for authentication")
	cmd.Flags().StringVar(&flags, "flags", "", "Flags to pass to claude-code (e.g., '--model opus')")
	cmd.Flags().StringArrayVar(&envVars, "env", []string{}, "Environment variables to pass (e.g., -e KEY=value)")
	cmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Enable 24-hour auto exit (default: true)")
	cmd.Flags().BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip HTTPS certificate verification (default: false)")
	cmd.Flags().BoolVar(&skipInstall, "skip-install-check", false, "Skip claude-code installation check (default: false)")

	// Set required flags
	cmd.MarkFlagRequired("host")

	return cmd
}
