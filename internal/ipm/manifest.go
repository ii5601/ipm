package ipm

import (
	"encoding/json"
	"fmt"
	"os"
)

// Manifest represents a package manifest file.
type Manifest struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Description string        `json:"description,omitempty"`
	Homepage    string        `json:"homepage,omitempty"`
	Platforms   []string      `json:"platforms,omitempty"`
	Install     []InstallStep `json:"install,omitempty"`
}

// LoadManifest reads and validates a manifest from the given file path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	return loadManifestData(data)
}

func loadManifestData(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate verifies that the manifest contains the required fields.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest missing required field: version")
	}
	for i, step := range m.Install {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("manifest install step %d: %w", i+1, err)
		}
	}
	return nil
}
