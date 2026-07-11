//go:build !windows

package ipm

import (
	"os/exec"
	"runtime"
)

// Notify displays a desktop notification.
// On macOS it uses osascript; on Linux it uses notify-send.
// Errors are silently ignored because notifications are best-effort.
func Notify(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		script := `display notification "` + message + `" with title "` + title + `"`
		_ = exec.Command("osascript", "-e", script).Run()
	case "linux":
		_ = exec.Command("notify-send", title, message).Run()
	}
}
