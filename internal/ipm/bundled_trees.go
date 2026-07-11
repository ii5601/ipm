package ipm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const legacyTreesDir = "tree"

func (m *Manager) FindManifest(tree, pkg string) (*Manifest, error) {
	if manifest, err := m.findLocalManifest(tree, pkg); err == nil {
		return manifest, nil
	} else if !errorsIsNotFound(err) {
		return nil, err
	}
	if manifest, err := findBundledManifest(tree, pkg); err == nil {
		return manifest, nil
	} else if !errorsIsNotFound(err) {
		return nil, err
	}
	return nil, fmt.Errorf("package %q not found in tree %q", pkg, tree)
}

func (m *Manager) listBundledAndLocalPackages(tree string) ([]*Manifest, error) {
	seen := map[string]struct{}{}
	var manifests []*Manifest

	local, err := m.listLocalPackages(tree)
	if err != nil {
		return nil, err
	}
	for _, manifest := range local {
		key := manifest.Name + "\x00" + manifest.Version
		seen[key] = struct{}{}
		manifests = append(manifests, manifest)
	}
	for _, manifest := range bundledPackages(tree) {
		key := manifest.Name + "\x00" + manifest.Version
		if _, ok := seen[key]; ok {
			continue
		}
		manifests = append(manifests, manifest)
	}

	sort.Slice(manifests, func(i, j int) bool {
		if manifests[i].Name == manifests[j].Name {
			return manifests[i].Version < manifests[j].Version
		}
		return manifests[i].Name < manifests[j].Name
	})
	return manifests, nil
}

func (m *Manager) listAllTrees() ([]string, error) {
	seen := map[string]struct{}{}
	for _, tree := range bundledTrees() {
		seen[tree] = struct{}{}
	}
	for _, dirName := range []string{treesDir, legacyTreesDir} {
		dir := filepath.Join(m.root, dirName)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("list trees: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				seen[entry.Name()] = struct{}{}
			}
		}
	}

	trees := make([]string, 0, len(seen))
	for tree := range seen {
		trees = append(trees, tree)
	}
	sort.Strings(trees)
	return trees, nil
}

func (m *Manager) findLocalManifest(tree, pkg string) (*Manifest, error) {
	for _, dirName := range []string{treesDir, legacyTreesDir} {
		treeDir := filepath.Join(m.root, dirName, tree)
		entries, err := os.ReadDir(treeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("find manifest: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			manifest, err := loadManifestJSON(filepath.Join(treeDir, entry.Name()))
			if err != nil {
				continue
			}
			if manifest.Name == pkg {
				return manifest, nil
			}
		}
	}
	return nil, fmt.Errorf("package %q not found", pkg)
}

func (m *Manager) listLocalPackages(tree string) ([]*Manifest, error) {
	var manifests []*Manifest
	seenFiles := map[string]struct{}{}
	for _, dirName := range []string{treesDir, legacyTreesDir} {
		treeDir := filepath.Join(m.root, dirName, tree)
		entries, err := os.ReadDir(treeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("list packages: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			path := filepath.Join(treeDir, entry.Name())
			if _, ok := seenFiles[path]; ok {
				continue
			}
			seenFiles[path] = struct{}{}
			manifest, err := loadManifestJSON(path)
			if err != nil {
				continue
			}
			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}

func findBundledManifest(tree, pkg string) (*Manifest, error) {
	for path, content := range bundledTreeFiles {
		if !strings.HasPrefix(path, filepath.ToSlash(filepath.Join(treesDir, tree))+"/") || filepath.Ext(path) != ".json" {
			continue
		}
		manifest, err := loadManifestData([]byte(content))
		if err != nil {
			continue
		}
		if manifest.Name == pkg {
			return manifest, nil
		}
	}
	return nil, fmt.Errorf("package %q not found", pkg)
}

func bundledPackages(tree string) []*Manifest {
	var manifests []*Manifest
	prefix := filepath.ToSlash(filepath.Join(treesDir, tree)) + "/"
	for path, content := range bundledTreeFiles {
		if !strings.HasPrefix(path, prefix) || filepath.Ext(path) != ".json" {
			continue
		}
		manifest, err := loadManifestData([]byte(content))
		if err != nil {
			continue
		}
		manifests = append(manifests, manifest)
	}
	return manifests
}

func bundledTrees() []string {
	seen := map[string]struct{}{}
	for path := range bundledTreeFiles {
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) >= 2 && parts[0] == treesDir {
			seen[parts[1]] = struct{}{}
		}
	}
	trees := make([]string, 0, len(seen))
	for tree := range seen {
		trees = append(trees, tree)
	}
	sort.Strings(trees)
	return trees
}

func errorsIsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
