package main

import (
	"context"
	"errors"
	"fmt"
	"os"

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

	switch args[0] {
	case "init":
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		manager := ipm.NewManager(root)
		if err := manager.Init(); err != nil {
			return err
		}
		fmt.Printf("initialized package root at %s\n", manager.Root())
		return nil
	case "tree":
		return runTreeCommand(args[1:])
	case "manifest":
		return runManifestCommand(args[1:])
	case "package":
		return runPackageCommand(args[1:])
	case "update":
		return runUpdateCommand(args[1:])
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runTreeCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("tree command requires a subcommand")
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			return errors.New("usage: ipm tree create <name> [root]")
		}
		root := "."
		if len(args) > 2 {
			root = args[2]
		}
		manager := ipm.NewManager(root)
		if err := manager.CreateTree(args[1]); err != nil {
			return err
		}
		fmt.Printf("created tree %s\n", args[1])
		return nil
	case "list":
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		manager := ipm.NewManager(root)
		trees, err := manager.ListTrees()
		if err != nil {
			return err
		}
		for _, tree := range trees {
			fmt.Println(tree)
		}
		return nil
	default:
		return fmt.Errorf("unknown tree subcommand %q", args[0])
	}
}

func runManifestCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("manifest command requires a subcommand")
	}

	switch args[0] {
	case "validate":
		if len(args) != 2 {
			return errors.New("usage: ipm manifest validate <file>")
		}
		manifest, err := ipm.LoadManifest(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("manifest %s %s is valid\n", manifest.Name, manifest.Version)
		return nil
	case "add":
		if len(args) < 3 {
			return errors.New("usage: ipm manifest add <tree> <file> [root]")
		}
		root := "."
		if len(args) > 3 {
			root = args[3]
		}
		manager := ipm.NewManager(root)
		target, err := manager.AddManifest(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Printf("stored manifest in %s\n", target)
		return nil
	default:
		return fmt.Errorf("unknown manifest subcommand %q", args[0])
	}
}

func runPackageCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("package command requires a subcommand")
	}

	switch args[0] {
	case "list":
		if len(args) < 2 {
			return errors.New("usage: ipm package list <tree> [root]")
		}
		root := "."
		if len(args) > 2 {
			root = args[2]
		}
		manager := ipm.NewManager(root)
		packages, err := manager.ListPackages(args[1])
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
  ipm init [root]
  ipm tree create <name> [root]
  ipm tree list [root]
  ipm manifest validate <file>
  ipm manifest add <tree> <file> [root]
  ipm package list <tree> [root]
  ipm update check <url> <token> [current-version]`)
}
