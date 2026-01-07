package src

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"clauded-client/src/commands"
	"clauded-client/src/platform"
)

//go:embed scripts/*.sh
var scriptsFS embed.FS

// Installer handles claude-code installation
type Installer struct {
	dryRun bool
}

// NewInstaller creates a new installer instance
func NewInstaller() *Installer {
	return &Installer{}
}

// IsClaudeCodeInstalled checks if claude or claude-code is installed
func (i *Installer) IsClaudeCodeInstalled() bool {
	finder := commands.NewFinder("claude")
	return finder.IsInstalled()
}

// GetClaudeCodeVersion returns the installed claude-code version
func (i *Installer) GetClaudeCodeVersion() (string, error) {
	finder := commands.NewFinder("claude")
	return finder.GetVersion()
}

// Install runs the installation process
func (i *Installer) Install() error {
	fmt.Println("Checking claude-code installation status...")

	// Check if already installed
	if i.IsClaudeCodeInstalled() {
		version, err := i.GetClaudeCodeVersion()
		if err == nil {
			fmt.Printf("claude-code is already installed: %s\n", version)
			return nil
		}
	}

	fmt.Println("claude-code is not installed, starting installation...")

	// Extract and run install script
	if err := i.runInstallScript(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("✅ claude-code installation completed!")

	// Verify installation
	if i.IsClaudeCodeInstalled() {
		version, err := i.GetClaudeCodeVersion()
		if err == nil {
			fmt.Printf("Installed version: %s\n", version)
		}
		return nil
	}

	return fmt.Errorf("installation verification failed")
}

// runInstallScript extracts and runs the embedded install script
func (i *Installer) runInstallScript() error {
	// Read the install script from embedded FS
	scriptContent, err := scriptsFS.ReadFile("scripts/install.sh")
	if err != nil {
		return fmt.Errorf("failed to read install script: %w", err)
	}

	// Create a temporary file for the script
	tmpFile, err := os.CreateTemp("", "claude-code-install-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write script content to temp file
	if _, err := tmpFile.Write(scriptContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute the script
	fmt.Printf("Running installation script on %s...\n", runtime.GOOS)
	cmd := exec.Command(tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

// GetSupportedOS returns a list of supported operating systems
func (i *Installer) GetSupportedOS() []string {
	return []string{"darwin", "linux"}
}

// IsOSSupported checks if the current OS is supported
func (i *Installer) IsOSSupported() bool {
	return platform.IsDarwin() || platform.IsLinux()
}

// DetectLinuxDistro detects the Linux distribution
func (i *Installer) DetectLinuxDistro() (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("not running on Linux")
	}

	// Try to read /etc/os-release
	if _, err := os.Stat("/etc/os-release"); err == nil {
		cmd := exec.Command("sh", "-c", ". /etc/os-release && echo $ID")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	// Fallback: check for distribution-specific files
	distroFiles := map[string]string{
		"debian": "/etc/debian_version",
		"ubuntu": "/etc/lsb-release",
		"alpine": "/etc/alpine-release",
	}

	for distro, file := range distroFiles {
		if _, err := os.Stat(file); err == nil {
			return distro, nil
		}
	}

	return "unknown", nil
}

// ListScripts lists all embedded scripts (for debugging)
func (i *Installer) ListScripts() ([]string, error) {
	var scripts []string

	err := fs.WalkDir(scriptsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sh") {
			scripts = append(scripts, path)
		}
		return nil
	})

	return scripts, err
}
