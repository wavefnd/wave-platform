#!/bin/bash
set -e

VERSION=""
REPO="wavefnd/Wave"
INSTALL_DIR="${WAVE_INSTALL_DIR:-$HOME/.wave/bin}"

usage() {
  echo "Wave Installer"
  echo "Usage:"
  echo "  bash install.sh --version <tag>"
  echo "  bash install.sh latest"
  echo "  curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest"
  exit 1
}

fail() {
  echo "[error] $1" >&2
  exit 1
}

normalize_version() {
  case "$1" in
    v*) printf "%s" "$1" ;;
    *) printf "v%s" "$1" ;;
  esac
}

append_path_if_needed() {
  case ":$PATH:" in
    *":$INSTALL_DIR:"*) return ;;
  esac

  export PATH="$INSTALL_DIR:$PATH"

  if [[ -n "${ZSH_VERSION:-}" ]]; then
    SHELL_RC="$HOME/.zshrc"
  elif [[ -n "${BASH_VERSION:-}" ]]; then
    SHELL_RC="$HOME/.bashrc"
  else
    SHELL_RC="$HOME/.profile"
  fi

  mkdir -p "$(dirname "$SHELL_RC")"
  touch "$SHELL_RC"

  if ! grep -F "export PATH=\"$INSTALL_DIR:\$PATH\"" "$SHELL_RC" >/dev/null 2>&1; then
    {
      echo
      echo "# Wave"
      echo "export PATH=\"$INSTALL_DIR:\$PATH\""
    } >> "$SHELL_RC"
    echo "[info] Added $INSTALL_DIR to PATH in $SHELL_RC"
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || fail "Missing value after --version."
      VERSION="$(normalize_version "$2")"
      shift 2
      ;;
    latest)
      VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=1" | grep -m 1 '"tag_name":' | cut -d '"' -f4)"
      [[ -n "$VERSION" ]] || fail "Unable to resolve latest release."
      echo "[info] Latest version: $VERSION"
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  fail "Missing version. Use --version <tag> or latest."
fi

echo "[info] Detecting system..."

UNAME_OUT="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH_SUFFIX="x86_64" ;;
  arm64|aarch64) ARCH_SUFFIX="aarch64" ;;
  *) fail "Unsupported architecture: $ARCH" ;;
esac

case "$UNAME_OUT" in
  Linux) OS_SUFFIX="linux-gnu" ;;
  Darwin) OS_SUFFIX="apple-darwin" ;;
  *) fail "This OS is not supported yet: $UNAME_OUT" ;;
esac

FILE_SUFFIX="${ARCH_SUFFIX}-${OS_SUFFIX}"
FILE_NAME="wave-${VERSION}-${FILE_SUFFIX}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILE_NAME}"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "[1/3] Downloading Wave $VERSION..."
echo "[info] Download: $URL"
curl -fL "$URL" -o "$TMP_DIR/$FILE_NAME"

echo "[2/3] Installing Wave..."
tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"

PACKAGE_DIR="$TMP_DIR/${FILE_NAME%.tar.gz}"
if [[ ! -d "$PACKAGE_DIR" ]]; then
  PACKAGE_DIR="$(find "$TMP_DIR" -mindepth 1 -maxdepth 1 -type d | head -n 1)"
fi

[[ -n "$PACKAGE_DIR" && -d "$PACKAGE_DIR" ]] || fail "Invalid package layout."
[[ -f "$PACKAGE_DIR/wavec" ]] || fail "Package does not contain wavec."
[[ -d "$PACKAGE_DIR/llvm" ]] || fail "Package does not contain bundled llvm/."

mkdir -p "$INSTALL_DIR"
rm -rf "$INSTALL_DIR/llvm"
cp "$PACKAGE_DIR/wavec" "$INSTALL_DIR/wavec"
cp -R "$PACKAGE_DIR/llvm" "$INSTALL_DIR/llvm"
chmod +x "$INSTALL_DIR/wavec"
chmod +x "$INSTALL_DIR/llvm/bin/"* 2>/dev/null || true

append_path_if_needed

echo "[3/3] Verifying installation..."
if command -v wavec >/dev/null 2>&1; then
  wavec --version
  echo "Installation completed successfully."
else
  fail "Installation failed. Restart your terminal or add $INSTALL_DIR to PATH."
fi
