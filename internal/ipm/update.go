package ipm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Version string `json:"version"`
	URL     string `json:"url,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

// CheckForUpdate queries the given endpoint for update information.
// client may be nil, in which case http.DefaultClient is used.
// token is sent as a ****** in the Authorization header.
// currentVersion is included as a query parameter if non-empty.
func CheckForUpdate(ctx context.Context, client *http.Client, endpoint, token, currentVersion string) (*UpdateInfo, error) {
	if client == nil {
		client = http.DefaultClient
	}
	reqURL := endpoint
	if currentVersion != "" {
		reqURL = fmt.Sprintf("%s?current=%s", endpoint, currentVersion)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("update check: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update check: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update check: server returned %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("update check: reading response: %w", err)
	}
	var info UpdateInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("update check: parsing response: %w", err)
	}
	if info.Version == "" {
		return nil, fmt.Errorf("update check: response missing version field")
	}
	return &info, nil
}
