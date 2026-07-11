//go:build !windows

package ipm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// RegisterProtocol registers the ipm:// URL scheme handler.
// On macOS it installs a minimal .app bundle and registers it with Launch Services.
// On Linux it installs a .desktop file and updates the MIME database.
// If handler is empty, the path to the current executable is used.
func RegisterProtocol(handler string) error {
	if handler == "" {
		var err error
		handler, err = os.Executable()
		if err != nil {
			return fmt.Errorf("register protocol: %w", err)
		}
		handler, err = filepath.Abs(handler)
		if err != nil {
			return fmt.Errorf("register protocol: %w", err)
		}
	}

	switch runtime.GOOS {
	case "darwin":
		return registerProtocolDarwin(handler)
	case "linux":
		return registerProtocolLinux(handler)
	default:
		return fmt.Errorf("register protocol: unsupported OS %q", runtime.GOOS)
	}
}

func registerProtocolDarwin(handler string) error {
	// Write a minimal Info.plist into a stub .app bundle so that Launch
	// Services can associate the ipm:// scheme with this executable.
	appDir := filepath.Join(os.Getenv("HOME"), "Applications", "ipm.app", "Contents")
	macosDir := filepath.Join(appDir, "MacOS")
	if err := os.MkdirAll(macosDir, 0o755); err != nil {
		return fmt.Errorf("register protocol (darwin): %w", err)
	}

	// Symlink the real binary into the .app bundle.
	stub := filepath.Join(macosDir, "ipm")
	_ = os.Remove(stub)
	if err := os.Symlink(handler, stub); err != nil {
		return fmt.Errorf("register protocol (darwin): %w", err)
	}

	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleIdentifier</key>
  <string>com.ipm.ipm</string>
  <key>CFBundleName</key>
  <string>ipm</string>
  <key>CFBundleExecutable</key>
  <string>ipm</string>
  <key>CFBundleURLTypes</key>
  <array>
    <dict>
      <key>CFBundleURLSchemes</key>
      <array>
        <string>ipm</string>
      </array>
      <key>CFBundleURLName</key>
      <string>ipm Protocol</string>
    </dict>
  </array>
</dict>
</plist>
`
	plistPath := filepath.Join(appDir, "Info.plist")
	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("register protocol (darwin): %w", err)
	}

	// Register with Launch Services.
	out, err := exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister",
		"-f", filepath.Dir(appDir)).CombinedOutput()
	if err != nil {
		// lsregister failure is non-fatal; the app bundle may still be picked up.
		_ = out
	}
	return nil
}

func registerProtocolLinux(handler string) error {
	desktopDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "applications")
	if err := os.MkdirAll(desktopDir, 0o755); err != nil {
		return fmt.Errorf("register protocol (linux): %w", err)
	}
	content := fmt.Sprintf("[Desktop Entry]\nName=ipm\nExec=%s %%u\nType=Application\nMimeType=x-scheme-handler/ipm;\n", handler)
	desktopFile := filepath.Join(desktopDir, "ipm.desktop")
	if err := os.WriteFile(desktopFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("register protocol (linux): %w", err)
	}
	// Update MIME database (best-effort).
	_ = exec.Command("update-desktop-database", desktopDir).Run()
	return nil
}

// UnregisterProtocol removes the ipm:// URL scheme handler.
func UnregisterProtocol() error {
	switch runtime.GOOS {
	case "darwin":
		appDir := filepath.Join(os.Getenv("HOME"), "Applications", "ipm.app")
		if err := os.RemoveAll(appDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unregister protocol (darwin): %w", err)
		}
		return nil
	case "linux":
		desktopFile := filepath.Join(os.Getenv("HOME"), ".local", "share", "applications", "ipm.desktop")
		if err := os.Remove(desktopFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unregister protocol (linux): %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unregister protocol: unsupported OS %q", runtime.GOOS)
	}
}
