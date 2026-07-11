#!/usr/bin/env bash
set -euo pipefail

repo="ii5601/ipm"
install_dir="${IPM_INSTALL_DIR:-$HOME/.local/bin}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux|darwin) ;;
  *)
    echo "unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64)
    if [[ "$os" != "darwin" ]]; then
      echo "unsupported architecture: $arch" >&2
      exit 1
    fi
    arch="arm64"
    ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

asset="ipm-${os}-${arch}"
url="https://github.com/${repo}/releases/latest/download/${asset}"

mkdir -p "$install_dir"
curl --fail --location --silent --show-error "$url" -o "$tmp_dir/ipm"
chmod +x "$tmp_dir/ipm"
mv "$tmp_dir/ipm" "$install_dir/ipm"

if [[ "${IPM_SKIP_PROTOCOL:-0}" != "1" ]]; then
  "$install_dir/ipm" protocol register || true
fi

echo "installed ipm to $install_dir/ipm"
echo "ensure $install_dir is in PATH"
