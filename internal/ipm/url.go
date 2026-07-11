package ipm

import (
	"fmt"
	"net/url"
	"strings"
)

// Scheme is the URL scheme used by the ipm protocol handler.
const Scheme = "ipm"

// Action represents a parsed ipm:// URL.
type Action struct {
	// Action is the operation to perform (e.g. "install").
	Action string
	// Tree is the optional package tree name.
	Tree string
	// Package is the package name.
	Package string
}

// ParseURL parses an ipm:// URL into an Action.
//
// Supported formats:
//
//	ipm://install/<package>
//	ipm://install/<tree>/<package>
func ParseURL(raw string) (*Action, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid ipm URL %q: %w", raw, err)
	}
	if u.Scheme != Scheme {
		return nil, fmt.Errorf("invalid ipm URL %q: expected scheme %q", raw, Scheme)
	}

	action := u.Host
	if action == "" {
		return nil, fmt.Errorf("invalid ipm URL %q: missing action", raw)
	}

	// u.Path starts with "/" for the remainder of the path.
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	a := &Action{Action: action}
	switch len(parts) {
	case 1:
		a.Package = parts[0]
	case 2:
		a.Tree = parts[0]
		a.Package = parts[1]
	}
	if a.Package == "" {
		return nil, fmt.Errorf("invalid ipm URL %q: missing package", raw)
	}
	return a, nil
}
