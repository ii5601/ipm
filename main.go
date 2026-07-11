package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ii5601/ipm/internal/ipm"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	// Handle ipm:// URLs passed directly as the first argument.
	if strings.HasPrefix(args[0], ipm.Scheme+"://") {
		return runProtocolURL(args[0])
	}

	switch args[0] {
	case "package":
		return runPackageCommand(args[1:])
	case "install":
		return runInstallCommand(args[1:])
	case "protocol":
		return runProtocolCommand(args[1:])
	case "update":
		return runUpdateCommand(args[1:])
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runInstallCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: ipm install <package>|<tree>/<package> [root]")
	}
	root := "."
	if len(args) > 1 {
		root = args[1]
	}
	manager := ipm.NewManager(root)
	manifest, err := manager.InstallPackage(args[0])
	if err != nil {
		return err
	}
	tree, pkg := parsePackageLabel(args[0], manifest.Name)
	fmt.Printf("installed %s from tree %s\n", pkg, tree)
	return nil
}

func runProtocolCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("protocol command requires a subcommand: register, unregister")
	}
	switch args[0] {
	case "register":
		handler := ""
		if len(args) > 1 {
			handler = args[1]
		}
		if err := ipm.RegisterProtocol(handler); err != nil {
			return err
		}
		fmt.Println("ipm:// protocol handler registered")
		return nil
	case "unregister":
		if err := ipm.UnregisterProtocol(); err != nil {
			return err
		}
		fmt.Println("ipm:// protocol handler unregistered")
		return nil
	default:
		return fmt.Errorf("unknown protocol subcommand %q", args[0])
	}
}

// runProtocolURL handles an ipm:// URL passed directly to the CLI.
func runProtocolURL(raw string) error {
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
		return runInstallCommand([]string{pkg})
	default:
		return fmt.Errorf("unknown ipm:// action %q", action.Action)
	}
}

func runPackageCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("package command requires a subcommand")
	}

	switch args[0] {
	case "list":
		tree := "main"
		root := "."
		if len(args) > 1 {
			tree = args[1]
		}
		if len(args) > 2 {
			root = args[2]
		}
		manager := ipm.NewManager(root)
		packages, err := manager.ListPackages(tree)
		if err != nil {
			return err
		}
		for _, manifest := range packages {
			fmt.Printf("%s %s\n", manifest.Name, manifest.Version)
		}
		return nil
	default:
		return fmt.Errorf("unknown package subcommand %q", args[0])
	}
}

func runUpdateCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("update command requires a subcommand")
	}

	switch args[0] {
	case "check":
		if len(args) < 3 {
			return errors.New("usage: ipm update check <url> <token> [current-version]")
		}
		currentVersion := ""
		if len(args) > 3 {
			currentVersion = args[3]
		}
		info, err := ipm.CheckForUpdate(context.Background(), nil, args[1], args[2], currentVersion)
		if err != nil {
			return err
		}
		fmt.Printf("latest version: %s\n", info.Version)
		if info.URL != "" {
			fmt.Printf("download: %s\n", info.URL)
		}
		if info.Notes != "" {
			fmt.Printf("notes: %s\n", info.Notes)
		}
		return nil
	default:
		return fmt.Errorf("unknown update subcommand %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`ipm - a small cross-platform package manager prototype

Usage:
  ipm package list [tree] [root]
  ipm install <package>|<tree>/<package> [root]
  ipm protocol register [handler-path]
  ipm protocol unregister
  ipm update check <url> <token> [current-version]
  ipm ipm://<action>/...     handle a protocol URL directly`)
}

func parsePackageLabel(ref, fallback string) (tree, pkg string) {
	tree, pkg = "main", fallback
	if parts := strings.SplitN(ref, "/", 2); len(parts) == 2 {
		tree = parts[0]
		pkg = parts[1]
	} else if strings.TrimSpace(ref) != "" {
		pkg = ref
	}
	return tree, pkg
}
