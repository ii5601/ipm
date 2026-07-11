//go:build windows

package ipm

import (
	"fmt"
	"os/exec"
)

// Notify displays a desktop notification on Windows using PowerShell's
// BurntToast-style balloon notification via the Windows Script Host.
func Notify(title, message string) {
	// Use PowerShell to show a toast/balloon notification.
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$n = New-Object System.Windows.Forms.NotifyIcon
$n.Icon = [System.Drawing.SystemIcons]::Information
$n.Visible = $true
$n.ShowBalloonTip(5000, '%s', '%s', [System.Windows.Forms.ToolTipIcon]::None)
Start-Sleep -Milliseconds 5500
$n.Dispose()
`, escapePS(title), escapePS(message))
	_ = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Run()
}

// escapePS escapes a string for safe embedding in a PowerShell single-quoted string.
func escapePS(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\'' {
			out = append(out, '\'', '\'')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}
