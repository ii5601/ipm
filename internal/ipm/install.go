package ipm

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const defaultTree = "main"

// InstallStep represents a single manifest installation action.
type InstallStep struct {
	RunShell           *RunShellStep           `json:"runShell,omitempty"`
	DownloadAndExecute *DownloadAndExecuteStep `json:"downloadAndExecute,omitempty"`
	AddGPGKey          *AddGPGKeyStep          `json:"addGPGKey,omitempty"`
}

// Validate verifies that the step contains exactly one action.
func (s InstallStep) Validate() error {
	actions := 0
	if s.RunShell != nil {
		actions++
	}
	if s.DownloadAndExecute != nil {
		actions++
	}
	if s.AddGPGKey != nil {
		actions++
	}
	if actions == 0 {
		return errors.New("step must define an action")
	}
	if actions > 1 {
		return errors.New("step must define exactly one action")
	}
	if s.RunShell != nil && strings.TrimSpace(s.RunShell.Command) == "" {
		return errors.New("runShell.command is required")
	}
	if s.DownloadAndExecute != nil && strings.TrimSpace(s.DownloadAndExecute.URL) == "" {
		return errors.New("downloadAndExecute.url is required")
	}
	if s.AddGPGKey != nil {
		if strings.TrimSpace(s.AddGPGKey.URL) == "" && strings.TrimSpace(s.AddGPGKey.Content) == "" {
			return errors.New("addGPGKey.url or addGPGKey.content is required")
		}
		if strings.TrimSpace(s.AddGPGKey.URL) != "" && strings.TrimSpace(s.AddGPGKey.Content) != "" {
			return errors.New("addGPGKey.url and addGPGKey.content cannot be used together")
		}
	}
	return nil
}

// RunShellStep runs a shell command.
type RunShellStep struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Dir     string            `json:"dir,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// DownloadAndExecuteStep downloads a file and executes it.
type DownloadAndExecuteStep struct {
	URL    string            `json:"url"`
	Args   []string          `json:"args,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
	SHA256 string            `json:"sha256,omitempty"`
}

// AddGPGKeyStep stores a GPG key in the user's keyring directory.
type AddGPGKeyStep struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	Content string `json:"content,omitempty"`
}

// InstallPackage resolves and executes a package manifest.
func (m *Manager) InstallPackage(ref string) (*Manifest, error) {
	tree, pkg := parsePackageRef(ref)
	manifest, err := m.FindManifest(tree, pkg)
	if err != nil {
		return nil, err
	}
	if err := executeInstallSteps(m.root, manifest.Install); err != nil {
		return nil, err
	}
	return manifest, nil
}

func parsePackageRef(ref string) (tree, pkg string) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return defaultTree, ref
}

func executeInstallSteps(root string, steps []InstallStep) error {
	for i, step := range steps {
		if err := executeInstallStep(root, step); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

func executeInstallStep(root string, step InstallStep) error {
	switch {
	case step.RunShell != nil:
		return runShellStep(root, *step.RunShell)
	case step.DownloadAndExecute != nil:
		return downloadAndExecuteStep(*step.DownloadAndExecute)
	case step.AddGPGKey != nil:
		return addGPGKeyStep(*step.AddGPGKey)
	default:
		return errors.New("unsupported install step")
	}
}

func runShellStep(root string, step RunShellStep) error {
	if err := validateCommandSafety(step.Command, step.Args); err != nil {
		return err
	}
	cmd := shellCommand(step.Command, step.Args...)
	cmd.Dir = resolveStepDir(root, step.Dir)
	cmd.Env = append(os.Environ(), formatEnv(step.Env)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runShell failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func downloadAndExecuteStep(step DownloadAndExecuteStep) error {
	resp, err := http.Get(step.URL)
	if err != nil {
		return fmt.Errorf("downloadAndExecute download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloadAndExecute download: server returned %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "ipm-download-*")
	if err != nil {
		return fmt.Errorf("downloadAndExecute temp file: %w", err)
	}
	path := tmpFile.Name()
	defer os.Remove(path)

	hasher := sha256.New()
	writer := io.Writer(tmpFile)
	if strings.TrimSpace(step.SHA256) != "" {
		writer = io.MultiWriter(tmpFile, hasher)
	}
	if _, err := io.Copy(writer, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("downloadAndExecute write: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("downloadAndExecute close: %w", err)
	}
	if strings.TrimSpace(step.SHA256) != "" {
		sum := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(sum, step.SHA256) {
			return fmt.Errorf("downloadAndExecute checksum mismatch: got %s", sum)
		}
	}
	if err := os.Chmod(path, 0o755); err != nil && !errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("downloadAndExecute chmod: %w", err)
	}

	cmd := exec.Command(path, step.Args...)
	cmd.Env = append(os.Environ(), formatEnv(step.Env)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("downloadAndExecute failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func addGPGKeyStep(step AddGPGKeyStep) error {
	keyData := step.Content
	if strings.TrimSpace(step.URL) != "" {
		resp, err := http.Get(step.URL)
		if err != nil {
			return fmt.Errorf("addGPGKey download: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("addGPGKey download: server returned %s", resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("addGPGKey download: %w", err)
		}
		keyData = string(body)
	}
	if strings.TrimSpace(keyData) == "" {
		return errors.New("addGPGKey key data is empty")
	}

	keyName := sanitizeKeyName(step.Name)
	if keyName == "" {
		keyName = "default"
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("addGPGKey keyring dir: %w", err)
	}
	targetDir := filepath.Join(dir, "ipm", "keyrings")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("addGPGKey keyring dir: %w", err)
	}
	target := filepath.Join(targetDir, keyName+".asc")
	if err := os.WriteFile(target, []byte(keyData), 0o644); err != nil {
		return fmt.Errorf("addGPGKey write key: %w", err)
	}
	return nil
}

func resolveStepDir(root, dir string) string {
	if strings.TrimSpace(dir) == "" {
		return root
	}
	if filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(root, dir)
}

func formatEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func shellCommand(command string, args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", append([]string{"/C", command}, args...)...)
	}
	return exec.Command("/bin/sh", "-c", joinShellCommand(command, args...))
}

func joinShellCommand(command string, args ...string) string {
	if len(args) == 0 {
		return command
	}
	parts := []string{command}
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

var dangerousCommandPatterns = []struct {
	re     *regexp.Regexp
	reason string
}{
	{regexp.MustCompile(`(?i)(^|[;&|])\s*(sudo\s+)?rm\s+-[^\n]*r[^\n]*f[^\n]*\s+(/|\*|/\*|\.|\./\*|~|~/\*|--no-preserve-root)`), "rm -rf on a root or wildcard path is blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*(sudo\s+)?(mkfs(\.[a-z0-9]+)?|fdisk|sfdisk|parted)\b`), "disk formatting commands are blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*dd\s+[^|;\n]*\bof=/dev/`), "writing directly to block devices is blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*(shutdown|reboot|poweroff|halt|init\s+0)\b`), "system shutdown commands are blocked"},
	{regexp.MustCompile(`:\(\)\s*\{\s*:\|:\&\s*\};:`), "fork bombs are blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*(del|erase)\s+/[a-z]*\s+[a-z]:\\\*`), "destructive Windows delete commands are blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*rmdir\s+/s\s+/q\s+[a-z]:\\`), "destructive Windows directory removal is blocked"},
	{regexp.MustCompile(`(?i)(^|[;&|])\s*format\s+[a-z]:`), "disk formatting commands are blocked"},
}

func validateCommandSafety(command string, args []string) error {
	full := strings.TrimSpace(joinShellCommand(command, args...))
	for _, pattern := range dangerousCommandPatterns {
		if pattern.re.MatchString(full) {
			return fmt.Errorf("unsafe command blocked: %s", pattern.reason)
		}
	}
	return nil
}

func sanitizeKeyName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, `\`, "-")
	return name
}
