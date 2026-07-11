package ipm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const treesDir = "trees"

// Manager manages a package root directory.
type Manager struct {
	root string
}

// NewManager returns a Manager rooted at the given directory.
func NewManager(root string) *Manager {
	return &Manager{root: root}
}

// Root returns the root directory of the manager.
func (m *Manager) Root() string {
	return m.root
}

// Init creates the package root structure (the trees/ subdirectory).
func (m *Manager) Init() error {
	dir := filepath.Join(m.root, treesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("init: %w", err)
	}
	return nil
}

// CreateTree creates a new named tree directory under trees/.
func (m *Manager) CreateTree(name string) error {
	dir := filepath.Join(m.root, treesDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create tree: %w", err)
	}
	return nil
}

// ListTrees returns the names of all available trees.
func (m *Manager) ListTrees() ([]string, error) {
	return m.listAllTrees()
}

// AddManifest copies the manifest file at src into the named tree and returns
// the destination path.
func (m *Manager) AddManifest(tree, src string) (string, error) {
	manifest, err := LoadManifest(src)
	if err != nil {
		return "", err
	}
	treeDir := filepath.Join(m.root, treesDir, tree)
	if err := os.MkdirAll(treeDir, 0o755); err != nil {
		return "", fmt.Errorf("add manifest: %w", err)
	}
	dst := filepath.Join(treeDir, manifest.Name+"-"+manifest.Version+".json")
	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("add manifest: %w", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", fmt.Errorf("add manifest: %w", err)
	}
	return dst, nil
}

// ListPackages returns all manifests stored in the named tree.
func (m *Manager) ListPackages(tree string) ([]*Manifest, error) {
	return m.listBundledAndLocalPackages(tree)
}

// loadManifestJSON reads a manifest JSON file without strict validation.
func loadManifestJSON(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
