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
		url                string
		session            string
		password           string
		remote             string
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
  clauded
  clauded --url=https://clauded.example.com
  clauded --url=https://clauded.example.com --password=my-secret-pass
  clauded --flags='--model claude-opus-4'
  clauded -e ANTHROPIC_BASE_URL=https://api.custom.com`,
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

			// Set default URL if not specified
			if url == "" {
				url = "https://clauded.friddle.me"
			}

			// Create configuration
			config := &src.Config{
				Name:               name,
				Host:               url,
				Session:            session,
				Password:           password,
				Remote:             remote,
				Flags:              flags,
				EnvVars:            envVars,
				AutoExit:           autoExit,
				InsecureSkipVerify: insecureSkipVerify,
			}

			// Validate configuration
			if err := config.Validate(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// Show security warning for default host
			if config.IsDefaultHost() {
				fmt.Printf("========================================\n")
				fmt.Printf("⚠️  WARNING: Using public server!\n")
				fmt.Printf("========================================\n")
				fmt.Printf("Session: %s\n", config.Session)
				fmt.Printf("Password: %s\n", config.Password)
				fmt.Printf("\nYour session is accessible at:\n")
				fmt.Printf("%s/%s\n", url, config.Session)
				fmt.Printf("\n⚠️  This is a PUBLIC server - NOT SECURE!\n")
				fmt.Printf("Anyone with the session ID can access your session.\n")
				fmt.Printf("For security, deploy your own server.\n")
				fmt.Printf("========================================\n\n")
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
	cmd.Flags().StringVar(&url, "url", "", "Remote server URL (default: https://clauded.friddle.me)")
	cmd.Flags().StringVar(&session, "session", "", "Session ID (auto-generated for default server)")
	cmd.Flags().StringVar(&password, "password", "", "Password for authentication (auto-generated for default server)")
	cmd.Flags().StringVar(&remote, "remote", "", "Remote piko server address (e.g., clauded.friddle.me:8022)")
	cmd.Flags().StringVar(&flags, "flags", "", "Flags to pass to claude-code (e.g., '--model opus')")
	cmd.Flags().StringArrayVar(&envVars, "env", []string{}, "Environment variables to pass (e.g., -e KEY=value)")
	cmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Enable 24-hour auto exit (default: true)")
	cmd.Flags().BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip HTTPS certificate verification (default: false)")
	cmd.Flags().BoolVar(&skipInstall, "skip-install-check", false, "Skip claude-code installation check (default: false)")

	return cmd
}
