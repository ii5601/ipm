//go:build windows

package ipm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RegisterProtocol registers the ipm:// URL scheme handler in the Windows registry.
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

	cmds := [][]string{
		{"reg", "add", `HKCU\Software\Classes\ipm`, "/ve", "/d", "URL:ipm Protocol", "/f"},
		{"reg", "add", `HKCU\Software\Classes\ipm`, "/v", "URL Protocol", "/d", "", "/f"},
		{"reg", "add", `HKCU\Software\Classes\ipm\shell\open\command`, "/ve", "/d", fmt.Sprintf(`"%s" "%%1"`, handler), "/f"},
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("register protocol: %w: %s", err, out)
		}
	}
	return nil
}

// UnregisterProtocol removes the ipm:// URL scheme handler from the Windows registry.
func UnregisterProtocol() error {
	out, err := exec.Command("reg", "delete", `HKCU\Software\Classes\ipm`, "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("unregister protocol: %w: %s", err, out)
	}
	return nil
}
