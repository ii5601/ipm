# ipm

ipm (Ilya Package Manager) is a small cross-platform package manager prototype written in Go.

## What is implemented

- `trees/` directory for repository trees such as `main`
- JSON package manifests with installation scenarios
- token-based update checks against an external web endpoint
- a simple CLI that works on Windows, Linux and macOS

## Manifest format

Each package manifest is a JSON file:

```json
{
  "name": "demo",
  "version": "1.0.0",
  "description": "Example package",
  "homepage": "https://example.com",
  "platforms": ["windows", "linux", "darwin"],
  "install": [
    {
      "runShell": {
        "command": "echo installing demo"
      }
    }
  ]
}
```

Supported install actions:

- `runShell`
- `downloadAndExecute`
- `addGPGKey`

Dangerous shell commands such as `rm -rf *`, disk formatting, block-device writes,
shutdown/reboot commands and similar destructive patterns are blocked.

## Commands

```bash
go run . package list
go run . package list main
go run . install demo
go run . install main/demo
go run . update check https://example.com/api/update YOUR_TOKEN 1.0.0
```

If `trees/main` exists in the repository, the build preparation script bundles it
into the final binary so regular users can install packages without manual
initialization.
