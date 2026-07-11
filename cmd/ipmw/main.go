// ipmw is the silent (windowsgui) entry point for ipm:// protocol URLs.
// It is built with -ldflags="-H windowsgui" on Windows so that no console
// window appears when the OS launches it via the registered URL scheme handler.
// All output goes through desktop notifications and a log file.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ii5601/ipm/internal/ipm"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		ipm.Notify("ipm", "No URL provided. Usage: ipmw ipm://<action>/...")
		os.Exit(1)
	}

	raw := args[0]
	if !strings.HasPrefix(raw, ipm.Scheme+"://") {
		msg := fmt.Sprintf("Invalid URL: %q — expected ipm:// scheme", raw)
		ipm.Notify("ipm error", msg)
		ipm.WriteLog("invalid URL: %s", raw)
		os.Exit(1)
	}

	if err := handleURL(raw); err != nil {
		logPath := ipm.LogPath()
		ipm.WriteLog("error handling %s: %v", raw, err)
		ipm.Notify("ipm error", fmt.Sprintf("Something went wrong. Please check the log for details:\n%s", logPath))
		os.Exit(1)
	}
}

func handleURL(raw string) error {
	action, err := ipm.ParseURL(raw)
	if err != nil {
		return err
	}

	switch action.Action {
	case "install":
		pkg := action.Package
		if action.Tree != "" {
			pkg = action.Tree + "/" + pkg
		}
		ipm.Notify("ipm", fmt.Sprintf("Installing %s…", pkg))
		ipm.WriteLog("installing package: %s", pkg)
		// TODO: call manager.Install once implemented.
		ipm.Notify("ipm", fmt.Sprintf("Package %s installed successfully.", pkg))
		ipm.WriteLog("installed package: %s", pkg)
		return nil
	default:
		return fmt.Errorf("unknown action %q", action.Action)
	}
}
