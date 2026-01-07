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
		session            string
		password           string
		codeCmd            string
		remote             string
		flags              string
		envVars            []string
		autoExit           bool
		insecureSkipVerify bool
		skipInstall        bool
		daemon             bool
	)

	rootCmd := &cobra.Command{
		Use:   "clauded",
		Short: "Claude Code remote client - Expose Claude Code via web terminal",
		Long: `clauded is a command-line tool that exposes your local Claude Code terminal session
through gotty and piko services to a remote server, allowing you to access and use
Claude Code from anywhere via a web browser.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(name, session, password, codeCmd, remote, flags, envVars, autoExit, insecureSkipVerify, skipInstall, daemon)
		},
	}

	// Add command line flags to root command
	rootCmd.Flags().StringVar(&name, "name", "", "Client identifier name (deprecated, use --session)")
	rootCmd.Flags().StringVar(&remote, "remote", "", "Remote server address (default: https://clauded.friddle.me)")
	rootCmd.Flags().StringVar(&session, "session", "", "Session ID (auto-generated for default server)")
	rootCmd.Flags().StringVar(&password, "password", "", "Password for authentication (auto-generated for default server)")
	rootCmd.Flags().StringVar(&codeCmd, "codecmd", "claude", "AI command tool to use (claude, opencode, kimi, gemini)")
	rootCmd.Flags().StringVar(&flags, "flags", "", "Flags to pass to codecmd (e.g., '--model opus')")
	rootCmd.Flags().StringArrayVar(&envVars, "env", []string{}, "Environment variables to pass (e.g., -e KEY=value)")
	rootCmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Enable 24-hour auto exit (default: true)")
	rootCmd.Flags().BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip HTTPS certificate verification (default: false)")
	rootCmd.Flags().BoolVar(&skipInstall, "skip-install-check", false, "Skip claude-code installation check (default: false)")
	rootCmd.Flags().BoolVarP(&daemon, "daemon", "d", true, "Run as daemon in background (default: true)")

	// Subcommand: session
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Manage clauded sessions",
	}
	rootCmd.AddCommand(sessionCmd)

	// Subcommand: list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all running clauded sessions",
		Run: func(cmd *cobra.Command, args []string) {
			src.ListSessions()
		},
	}
	sessionCmd.AddCommand(listCmd)

	// Subcommand: kill
	killCmd := &cobra.Command{
		Use:   "kill [session_id|all]",
		Short: "Kill a specific session or all sessions",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			src.KillSession(args[0])
		},
	}
	sessionCmd.AddCommand(killCmd)

	// Subcommand: kill-all
	killAllCmd := &cobra.Command{
		Use:   "kill-all",
		Short: "Kill all sessions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			src.KillSession("all")
		},
	}
	sessionCmd.AddCommand(killAllCmd)

	return rootCmd
}

func runServe(name, session, password, codeCmd, remote, flags string, envVars []string, autoExit, insecureSkipVerify, skipInstall, daemon bool) error {
	// Check and install claude-code if needed (only for claude command)
	if !skipInstall && codeCmd == "claude" {
		installer := src.NewInstaller()
		if !installer.IsClaudeCodeInstalled() {
			fmt.Println("claude-code not found, starting automatic installation...")
			if err := installer.Install(); err != nil {
				return fmt.Errorf("failed to install claude-code: %w", err)
			}
		}
	}

	// Set default remote if not specified
	if remote == "" {
		remote = "https://clauded.friddle.me:8022"
	}

	// Create configuration
	config := &src.Config{
		Name:               name,
		Remote:             remote,
		Session:            session,
		Password:           password,
		CodeCmd:            codeCmd,
		Flags:              flags,
		EnvVars:            envVars,
		AutoExit:           autoExit,
		InsecureSkipVerify: insecureSkipVerify,
		Daemon:             daemon,
		SkipInstall:        skipInstall,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Check for Claude authentication (API Key or Login)
	if err := src.ValidateAuthConfig(config); err != nil {
		return err
	}

	// Show connection information
	fmt.Printf("========================================\n")
	fmt.Printf("✓ Session started successfully!\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Session ID: %s\n", config.Session)
	if config.Password != "" {
		fmt.Printf("Password: %s\n", config.Password)
	}
	fmt.Printf("\nAccess URL:\n")
	fmt.Printf("%s/%s\n", config.GetHTTPURL(), config.Session)
	if config.IsDefaultHost() {
		fmt.Printf("\n⚠️  WARNING: Using public demo server!\n")
		fmt.Printf("For security, deploy your own server.\n")
	}
	fmt.Printf("========================================\n\n")

	// Create service manager
	manager := src.NewServiceManager(config)

	// Start services (blocks until service stops)
	if err := manager.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}
