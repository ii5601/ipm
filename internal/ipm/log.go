package ipm

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogPath returns the path to the ipm log file.
func LogPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "ipm", "ipm.log")
}

// WriteLog appends a formatted log line to the ipm log file.
// Errors are silently ignored because logging is best-effort.
func WriteLog(format string, args ...interface{}) {
	path := LogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	_, _ = f.WriteString(line)
}
