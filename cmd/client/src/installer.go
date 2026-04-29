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

// Installer handles AI command tool installation
type Installer struct {
	codeCmd string
	dryRun  bool
}

// NewInstaller creates a new installer instance
func NewInstaller() *Installer {
	return &Installer{}
}

// NewInstallerForCmd creates an installer for a specific command
func NewInstallerForCmd(codeCmd string) *Installer {
	return &Installer{codeCmd: codeCmd}
}

// IsCommandInstalled checks if the specified command is installed
func (i *Installer) IsCommandInstalled(cmd string) bool {
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}
	return false
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

// Install runs the installation process for the specified code command
func (i *Installer) Install() error {
	cmd := i.codeCmd
	if cmd == "" {
		cmd = "claude"
	}

	switch cmd {
	case "claude":
		return i.installClaude()
	case "opencode":
		return i.installOpenCode()
	case "gemini":
		return i.installGemini()
	default:
		fmt.Printf("⚠️  No auto-install support for '%s', skipping installation\n", cmd)
		return nil
	}
}

// installClaude installs Claude Code
func (i *Installer) installClaude() error {
	fmt.Println("Checking claude installation status...")

	if i.IsClaudeCodeInstalled() {
		version, err := i.GetClaudeCodeVersion()
		if err == nil {
			fmt.Printf("claude is already installed: %s\n", version)
			return nil
		}
	}

	fmt.Println("claude is not installed, starting installation...")

	installCmd := exec.Command("bash", "-c", "curl -fsSL https://claude.ai/install.sh | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("⚠️  claude installation via curl failed: %v\n", err)
		fmt.Println("⚠️  Trying embedded install script as fallback...")

		if err := i.runInstallScript(); err != nil {
			return fmt.Errorf("claude installation failed: %w", err)
		}
	}

	fmt.Println("✅ claude installation completed!")

	if i.IsClaudeCodeInstalled() {
		version, err := i.GetClaudeCodeVersion()
		if err == nil {
			fmt.Printf("Installed version: %s\n", version)
		}
		return nil
	}

	return fmt.Errorf("claude installation verification failed")
}

// installOpenCode installs opencode
func (i *Installer) installOpenCode() error {
	fmt.Println("Checking opencode installation status...")

	if i.IsCommandInstalled("opencode") {
		fmt.Printf("✓ opencode is already installed\n")
		return nil
	}

	fmt.Println("opencode is not installed, starting installation...")

	installCmd := exec.Command("bash", "-c", "curl -fsSL https://opencode.ai/install | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("⚠️  opencode installation failed: %v\n", err)
		fmt.Println("⚠️  Please install opencode manually: https://opencode.ai")
		return fmt.Errorf("opencode installation failed: %w", err)
	}

	if i.IsCommandInstalled("opencode") {
		fmt.Println("✅ opencode installation completed!")
		return nil
	}

	return fmt.Errorf("opencode installation verification failed")
}

// installGemini installs Gemini CLI
func (i *Installer) installGemini() error {
	fmt.Println("Checking gemini installation status...")

	if i.IsCommandInstalled("gemini") {
		fmt.Printf("✓ gemini is already installed\n")
		return nil
	}

	fmt.Println("gemini is not installed, starting installation...")

	installCmd := exec.Command("bash", "-c", "npm install -g @anthropic-ai/gemini-cli 2>/dev/null || npm install -g @google/gemini-cli")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("⚠️  gemini installation failed: %v\n", err)
		fmt.Println("⚠️  Trying alternative installation method...")

		installCmd2 := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/nicepkg/gemini-cli/main/install.sh | bash")
		installCmd2.Stdout = os.Stdout
		installCmd2.Stderr = os.Stderr
		if err := installCmd2.Run(); err != nil {
			fmt.Printf("⚠️  gemini alternative installation also failed: %v\n", err)
			fmt.Println("⚠️  Please install gemini manually: https://github.com/nicepkg/gemini-cli")
			return fmt.Errorf("gemini installation failed: %w", err)
		}
	}

	if i.IsCommandInstalled("gemini") {
		fmt.Println("✅ gemini installation completed!")
		return nil
	}

	return fmt.Errorf("gemini installation verification failed")
}

// InstallDefault tries to install any available AI command tool
// Priority: claude -> opencode -> gemini
func (i *Installer) InstallDefault() error {
	if _, err := exec.LookPath("claude"); err == nil {
		fmt.Println("✓ claude is available, skipping installation")
		return nil
	}

	fmt.Println("claude not found, trying to install claude...")
	if claudeErr := i.installClaude(); claudeErr == nil {
		return nil
	} else {
		fmt.Printf("⚠️  claude installation failed: %v\n", claudeErr)
	}

	if _, err := exec.LookPath("opencode"); err == nil {
		fmt.Println("✓ opencode is available, using it instead")
		return nil
	}

	fmt.Println("opencode not found, trying to install opencode...")
	if openCodeErr := i.installOpenCode(); openCodeErr == nil {
		return nil
	} else {
		fmt.Printf("⚠️  opencode installation failed: %v\n", openCodeErr)
	}

	if _, err := exec.LookPath("gemini"); err == nil {
		fmt.Println("✓ gemini is available, using it instead")
		return nil
	}

	fmt.Println("gemini not found, trying to install gemini...")
	if geminiErr := i.installGemini(); geminiErr == nil {
		return nil
	} else {
		fmt.Printf("⚠️  gemini installation failed: %v\n", geminiErr)
	}

	return fmt.Errorf("no AI command tool could be installed (tried claude, opencode, gemini)")
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
