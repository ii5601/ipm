# ipm

ipm (Ilya Package Manager) is a small cross-platform package manager prototype written in Go.

## What is implemented

- `tree/` directory for repository trees such as `Main`, `Game`, `Extra`
- JSON package manifests
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
  "platforms": ["windows", "linux", "darwin"]
}
```

## Commands

```bash
go run . init
go run . tree create Main
go run . manifest validate /path/to/demo.json
go run . manifest add Main /path/to/demo.json
go run . package list Main
go run . update check https://example.com/api/update YOUR_TOKEN 1.0.0
```
