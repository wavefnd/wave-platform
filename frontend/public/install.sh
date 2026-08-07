#!/usr/bin/env bash
set -euo pipefail

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

resolve_shell_rc() {
    local login_shell="${SHELL:-}"

    if [[ -z "$login_shell" ]] && command -v getent >/dev/null 2>&1; then
        login_shell="$(getent passwd "$(id -u)" 2>/dev/null | cut -d: -f7 || true)"
    fi

    case "${login_shell##*/}" in
        zsh)
            SHELL_RC="$HOME/.zshrc"
            SHELL_PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
            SHELL_RELOAD_COMMAND="source \"$HOME/.zshrc\""
            ;;
        bash)
            SHELL_RC="$HOME/.bashrc"
            SHELL_PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
            SHELL_RELOAD_COMMAND="source \"$HOME/.bashrc\""
            ;;
        fish)
            SHELL_RC="$HOME/.config/fish/config.fish"
            SHELL_PATH_LINE="fish_add_path \"$INSTALL_DIR\""
            SHELL_RELOAD_COMMAND="source \"$HOME/.config/fish/config.fish\""
            ;;
        *)
            SHELL_RC="$HOME/.profile"
            SHELL_PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""
            SHELL_RELOAD_COMMAND=". \"$HOME/.profile\""
            ;;
    esac
}

append_path_if_needed() {
    PATH_CONFIG_UPDATED=0
    resolve_shell_rc

    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *) export PATH="$INSTALL_DIR:$PATH" ;;
    esac

    mkdir -p "$(dirname "$SHELL_RC")"
    touch "$SHELL_RC"

    if ! grep -F "$SHELL_PATH_LINE" "$SHELL_RC" >/dev/null 2>&1; then
        {
            echo
            echo "# Wave"
            echo "$SHELL_PATH_LINE"
        } >> "$SHELL_RC"
        PATH_CONFIG_UPDATED=1
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
INSTALL_PARENT="$(dirname "$INSTALL_DIR")"

mkdir -p "$INSTALL_PARENT"
TMP_DIR="$(mktemp -d "$INSTALL_PARENT/.wave-install.XXXXXX")"
STAGE_DIR="${INSTALL_DIR}.new.$$"
BACKUP_DIR="${INSTALL_DIR}.old.$$"

cleanup() {
    rm -rf "$TMP_DIR" "$STAGE_DIR"
}
trap cleanup EXIT

echo "[1/3] Downloading Wave $VERSION..."
echo "[info] Download: $URL"
curl -fL "$URL" -o "$TMP_DIR/$FILE_NAME"

echo "[2/3] Installing Wave..."
tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"

PACKAGE_DIR="$TMP_DIR/${FILE_NAME%.tar.gz}"
if [[ ! -d "$PACKAGE_DIR" ]]; then
    PACKAGE_DIR="$(find "$TMP_DIR" -mindepth 1 -maxdepth 1 -type d ! -name '.wave-install.*' | head -n 1)"
fi

[[ -n "$PACKAGE_DIR" && -d "$PACKAGE_DIR" ]] || fail "Invalid package layout."
[[ -f "$PACKAGE_DIR/wavec" ]] || fail "Package does not contain wavec."
[[ -d "$PACKAGE_DIR/llvm" ]] || fail "Package does not contain bundled llvm/."

rm -rf "$STAGE_DIR" "$BACKUP_DIR"
mkdir -p "$STAGE_DIR"
cp "$PACKAGE_DIR/wavec" "$STAGE_DIR/wavec"
cp -R "$PACKAGE_DIR/llvm" "$STAGE_DIR/llvm"
chmod +x "$STAGE_DIR/wavec"
chmod +x "$STAGE_DIR/llvm/bin/"* 2>/dev/null || true

if [[ -d "$INSTALL_DIR" ]]; then
    mv "$INSTALL_DIR" "$BACKUP_DIR"
fi

if ! mv "$STAGE_DIR" "$INSTALL_DIR"; then
    if [[ -d "$BACKUP_DIR" && ! -d "$INSTALL_DIR" ]]; then
        mv "$BACKUP_DIR" "$INSTALL_DIR"
    fi
    fail "Unable to activate the new Wave installation."
fi

append_path_if_needed

echo "[3/3] Verifying installation..."
if "$INSTALL_DIR/wavec" --version; then
    echo "Installation completed successfully."
else
    if [[ -d "$BACKUP_DIR" ]]; then
        rm -rf "$INSTALL_DIR"
        mv "$BACKUP_DIR" "$INSTALL_DIR"
    fi
    fail "Installation verification failed."
fi

rm -rf "$BACKUP_DIR"

if [[ "$PATH_CONFIG_UPDATED" -eq 1 ]]; then
    echo "[info] To use 'wavec' in the current terminal, run: $SHELL_RELOAD_COMMAND"
fi