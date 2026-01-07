package src

import (
	"fmt"
	"os"
	"path/filepath"

	"clauded-client/src/platform"
)

// ValidateAuthConfig checks if authorized configuration is present
// and warns/errors if not.
func ValidateAuthConfig(config *Config) error {
	// 1. Check for Anthropic API Key or Auth Token (env var)
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		return nil
	}

	// 1.1 Check for Anthropic API Key/Token in config.EnvVars (command line flags)
	for _, env := range config.EnvVars {
		if (len(env) > 18 && env[:18] == "ANTHROPIC_API_KEY=") || 
		   (len(env) > 21 && env[:21] == "ANTHROPIC_AUTH_TOKEN=") {
			return nil
		}
	}

	// 2. Check for config file existence (subscription/login)
	// We check common paths where claude-code might store credentials
	configDirs := platform.GetConfigDirs()

	// Check for common config filenames
	configFiles := platform.GetConfigFiles()

	foundConfig := false
	for _, dir := range configDirs {
		for _, f := range configFiles {
			path := filepath.Join(dir, f)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				foundConfig = true
				break
			}
		}
		if foundConfig {
			break
		}
	}

	if foundConfig {
		return nil
	}

	// 3. Warning (Previously Error)
	fmt.Println("================================================================")
	fmt.Println("⚠️  WARNING: No authentication found for Claude Code.")
	fmt.Println("================================================================")
	fmt.Println("You might need to:")
	fmt.Println("1. Set the ANTHROPIC_API_KEY environment variable.")
	fmt.Println("   Example: export ANTHROPIC_API_KEY=sk-...")
	fmt.Println("   OR pass it: clauded -e ANTHROPIC_API_KEY=...")
	fmt.Println("")
	fmt.Println("2. Login via Claude Code CLI:")
	fmt.Println("   Run: claude login")
	fmt.Println("================================================================")
	fmt.Println("Continuing anyway... If Claude Code fails, check your auth setup.")

	return nil
}
