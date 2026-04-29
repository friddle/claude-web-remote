package main

import (
	"fmt"
	"os"
	"os/exec"

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
		session            string
		authName           string
		remote             string
		codeCmd            string
		flags              string
		envVars            []string
		terminal           string
		pass               string
		tmux               bool
		auth               bool
		autoExit           bool
		enableNotify       bool
		notifyWebhook      string
		staticIndex        string
		attachPort         string
		insecureSkipVerify bool
		skipInstall        bool
		daemon             bool
		pidFile            string
	)

	rootCmd := &cobra.Command{
		Use:   "clauded",
		Short: "Share your AI coding terminal as a web application via piko",
		Long: `clauded is a tool that exposes your local AI coding terminal session
through gotty and piko services to a remote server, allowing you to access and use
Claude Code / OpenCode / Gemini from anywhere via a web browser.

Examples:
  clauded --remote=https://clauded.friddle.me
  clauded --remote=https://piko.example.com:8088 --session myterm --tmux=true
  clauded --remote=https://piko.example.com:8088 --auth=false
  clauded --remote=https://piko.example.com:8088 --notify-webhook=https://open.feishu.cn/...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(session, authName, pass, codeCmd, remote, flags, envVars, terminal, tmux, auth, autoExit, enableNotify, notifyWebhook, staticIndex, attachPort, insecureSkipVerify, skipInstall, daemon, pidFile)
		},
	}

	rootCmd.Flags().StringVar(&session, "session", "", "Session ID for endpoint path (default: user_dir_random)")
	rootCmd.Flags().StringVar(&authName, "auth-name", "", "Auth username for Basic Auth (auto-generated if not set)")
	rootCmd.Flags().StringVar(&remote, "remote", "https://clauded.friddle.me", "Remote piko server address")
	rootCmd.Flags().StringVar(&codeCmd, "codecmd", "claude", "AI command tool to use (claude, opencode, kimi, gemini)")
	rootCmd.Flags().StringVar(&flags, "flags", "", "Flags to pass to codecmd (e.g., '--model opus')")
	rootCmd.Flags().StringArrayVar(&envVars, "env", []string{}, "Environment variables to pass (e.g., -e KEY=value)")
	rootCmd.Flags().StringVar(&terminal, "terminal", "", "Terminal type (zsh, bash, sh, powershell, etc.)")
	rootCmd.Flags().StringVar(&pass, "pass", "", "Auth password (auto-generated if not set)")
	rootCmd.Flags().BoolVar(&tmux, "tmux", true, "Use tmux for persistent sessions")
	rootCmd.Flags().BoolVar(&auth, "auth", true, "Enable Basic Authentication")
	rootCmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Enable 24-hour auto exit")
	rootCmd.Flags().BoolVar(&enableNotify, "enable-notify", true, "Enable notify-send interception")
	rootCmd.Flags().StringVar(&notifyWebhook, "notify-webhook", "", "Webhook URL to forward notifications to (Feishu compatible)")
	rootCmd.Flags().StringVar(&staticIndex, "static-index", ".", "Local directory to serve as static files at /files/")
	rootCmd.Flags().StringVar(&attachPort, "attach-port", "", "Map a local port to /port/ path (e.g. 3000)")
	rootCmd.Flags().BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "Skip HTTPS certificate verification")
	rootCmd.Flags().BoolVar(&skipInstall, "skip-install-check", false, "Skip AI command installation check")
	rootCmd.Flags().BoolVarP(&daemon, "daemon", "d", true, "Run as daemon (background process)")
	rootCmd.Flags().StringVar(&pidFile, "pid-file", "/tmp/clauded.pid", "PID file path for daemon mode")

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

func runServe(session, authName, pass, codeCmd, remote, flags string, envVars []string, terminal string, tmux, auth, autoExit, enableNotify bool, notifyWebhook, staticIndex, attachPort string, insecureSkipVerify, skipInstall, daemon bool, pidFile string) error {
	// Check and install AI command tool if needed
	if !skipInstall {
		installer := src.NewInstallerForCmd(codeCmd)
		switch codeCmd {
		case "claude":
			if !installer.IsClaudeCodeInstalled() {
				fmt.Println("claude not found, starting automatic installation...")
				if err := installer.Install(); err != nil {
					fmt.Printf("⚠️  Warning: %v\n", err)
					fmt.Println("⚠️  The installation may have succeeded, but verification failed.")
					fmt.Println("⚠️  If claude was installed, try running with --skip-install-check flag")
				}
			}
		case "opencode":
			if !installer.IsCommandInstalled("opencode") {
				fmt.Println("opencode not found, starting automatic installation...")
				if err := installer.Install(); err != nil {
					fmt.Printf("⚠️  Warning: %v\n", err)
					fmt.Println("⚠️  Please install opencode manually: https://opencode.ai")
				}
			}
		case "gemini":
			if !installer.IsCommandInstalled("gemini") {
				fmt.Println("gemini not found, starting automatic installation...")
				if err := installer.Install(); err != nil {
					fmt.Printf("⚠️  Warning: %v\n", err)
					fmt.Println("⚠️  Please install gemini manually")
				}
			}
		default:
			// Try default installation: claude -> opencode -> gemini
			if _, err := exec.LookPath(codeCmd); err != nil {
				fmt.Printf("%s not found, trying default installation...\n", codeCmd)
				if err := installer.InstallDefault(); err != nil {
					fmt.Printf("⚠️  Warning: %v\n", err)
				}
			}
		}
	}

	// Set default remote if not specified
	if remote == "" {
		remote = "https://clauded.friddle.me"
	}

	// Setup notification hooks for AI tools
	if enableNotify {
		if err := src.SetupClaudeCodeHooks(codeCmd); err != nil {
			fmt.Printf("⚠️  Warning: failed to setup notification hooks: %v\n", err)
		}
	}

	// Create configuration
	config := &src.Config{
		Session:            session,
		AuthName:           authName,
		Password:           pass,
		Remote:             remote,
		CodeCmd:            codeCmd,
		Flags:              flags,
		EnvVars:            envVars,
		Terminal:           terminal,
		Tmux:               tmux,
		Auth:               auth,
		AutoExit:           autoExit,
		EnableNotify:       enableNotify,
		NotifyWebhook:      notifyWebhook,
		StaticIndex:        staticIndex,
		AttachPort:         attachPort,
		InsecureSkipVerify: insecureSkipVerify,
		Daemon:             daemon,
		SkipInstall:        skipInstall,
		PidFile:            pidFile,
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
