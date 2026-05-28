#!/usr/bin/env sh
# aic install script. Usage:
#   curl -sSL https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh | sh
#   AIC_VERSION=v0.1.0 curl -sSL ... | sh         # pin version
#   AIC_INSTALL_DIR=$HOME/.local/bin curl ... | sh  # custom install dir

set -eu

REPO="learners-superpumped/aic"
BINARY="aic"
VERSION="${AIC_VERSION:-latest}"
INSTALL_DIR="${AIC_INSTALL_DIR:-/usr/local/bin}"

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    darwin) echo "darwin" ;;
    linux) echo "linux" ;;
    *) echo "unsupported OS: $os" >&2; exit 1 ;;
  esac
}

uname_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "x86_64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported arch: $arch" >&2; exit 1 ;;
  esac
}

resolve_version() {
  if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" \
      | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
      echo "failed to resolve latest version" >&2
      echo "no stable release found at https://github.com/$REPO/releases/latest" >&2
      echo "pin a version explicitly, e.g.: AIC_VERSION=v0.1.0 curl ... | sh" >&2
      exit 1
    fi
  fi
}

main() {
  OS=$(uname_os)
  ARCH=$(uname_arch)
  resolve_version
  VERSION_NO_V="${VERSION#v}"

  ARCHIVE="${BINARY}_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
  CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

  TMP=$(mktemp -d)
  trap 'rm -rf "$TMP"' EXIT

  echo "Downloading $URL"
  curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
  curl -fsSL "$CHECKSUMS_URL" -o "$TMP/checksums.txt"

  echo "Verifying checksum"
  if command -v sha256sum >/dev/null 2>&1; then
    SHA_CMD="sha256sum --check -"
  elif command -v shasum >/dev/null 2>&1; then
    SHA_CMD="shasum -a 256 -c -"
  else
    echo "neither sha256sum nor shasum is installed; cannot verify download" >&2
    exit 1
  fi
  # shellcheck disable=SC2086
  (cd "$TMP" && grep " $ARCHIVE\$" checksums.txt | $SHA_CMD)

  tar -C "$TMP" -xzf "$TMP/$ARCHIVE"

  if [ ! -d "$INSTALL_DIR" ]; then
    if mkdir -p "$INSTALL_DIR" 2>/dev/null; then :; else
      echo "Creating $INSTALL_DIR (requires sudo)"
      sudo mkdir -p "$INSTALL_DIR"
    fi
  fi

  if [ ! -w "$INSTALL_DIR" ]; then
    echo "Installing to $INSTALL_DIR (requires sudo)"
    sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
  else
    mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
  fi
  chmod +x "$INSTALL_DIR/$BINARY"

  echo "Installed $("$INSTALL_DIR"/$BINARY --version)"
}

main "$@"
