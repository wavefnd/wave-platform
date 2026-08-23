#!/usr/bin/env bash
set -euo pipefail

WAVE_VERSION=""
VEX_VERSION="${VEX_VERSION:-}"
WAVE_REPO="wavefnd/Wave"
VEX_REPO="wavefnd/Vex"
INSTALL_DIR="${WAVE_INSTALL_DIR:-$HOME/.wave/bin}"

usage() {
    local exit_code="${1:-1}"
    echo "Wave Toolchain Installer"
    echo "Usage:"
    echo "  bash install.sh --version <wave-tag> [--vex-version <vex-tag>]"
    echo "  bash install.sh latest"
    echo "  curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest"
    exit "$exit_code"
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

validate_version() {
    [[ "$1" =~ ^v[0-9A-Za-z][0-9A-Za-z._+-]*$ ]] || fail "Invalid version tag: $1"
}

resolve_latest_version() {
    local repository="$1"
    local response
    local version

    response="$(curl -fsSL -H "Accept: application/vnd.github+json" "https://api.github.com/repos/${repository}/releases?per_page=1")" \
        || fail "Unable to query releases for ${repository}."
    version="$(awk -F '"' '/"tag_name":/ { print $4; exit }' <<< "$response")"
    [[ -n "$version" ]] || fail "Unable to resolve the latest release for ${repository}."
    validate_version "$version"
    printf "%s" "$version"
}

sha256_file() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{ print $1 }'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{ print $1 }'
    else
        fail "SHA-256 verification requires sha256sum or shasum."
    fi
}

verify_checksum() {
    local archive="$1"
    local sums_file="$2"
    local file_name="$3"
    local expected
    local actual

    expected="$(awk -v file="$file_name" '$2 == file || $2 == ("*" file) { print $1; exit }' "$sums_file")"
    [[ "$expected" =~ ^[0-9A-Fa-f]{64}$ ]] || fail "No valid checksum was published for $file_name."
    actual="$(sha256_file "$archive")"
    [[ "$actual" == "$expected" ]] || fail "Checksum verification failed for $file_name."
    echo "[info] Verified SHA-256: $file_name"
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
        --version|--wave-version)
            [[ $# -ge 2 ]] || fail "Missing value after $1."
            WAVE_VERSION="$(normalize_version "$2")"
            shift 2
            ;;
        --vex-version)
            [[ $# -ge 2 ]] || fail "Missing value after --vex-version."
            VEX_VERSION="$(normalize_version "$2")"
            shift 2
            ;;
        latest)
            WAVE_VERSION="$(resolve_latest_version "$WAVE_REPO")"
            VEX_VERSION="$(resolve_latest_version "$VEX_REPO")"
            echo "[info] Latest Wave version: $WAVE_VERSION"
            echo "[info] Latest Vex version: $VEX_VERSION"
            shift
            ;;
        -h|--help)
            usage 0
            ;;
        *)
            usage
            ;;
    esac
done

if [[ -z "$WAVE_VERSION" ]]; then
    fail "Missing Wave version. Use --version <tag> or latest."
fi
validate_version "$WAVE_VERSION"

if [[ -z "$VEX_VERSION" ]]; then
    VEX_VERSION="$(resolve_latest_version "$VEX_REPO")"
    echo "[info] Latest Vex version: $VEX_VERSION"
fi
validate_version "$VEX_VERSION"

echo "[info] Detecting system..."

UNAME_OUT="$(uname -s)"
ARCH="$(uname -m)"

case "$UNAME_OUT:$ARCH" in
    Linux:x86_64|Linux:amd64)
        WAVE_FILE_SUFFIX="x86_64-linux-gnu"
        VEX_FILE_SUFFIX="x86_64-unknown-linux-gnu"
        ;;
    Darwin:arm64|Darwin:aarch64)
        WAVE_FILE_SUFFIX="aarch64-apple-darwin"
        VEX_FILE_SUFFIX="aarch64-apple-darwin"
        ;;
    Darwin:x86_64|Darwin:amd64)
        WAVE_FILE_SUFFIX="x86_64-apple-darwin"
        VEX_FILE_SUFFIX="x86_64-apple-darwin"
        ;;
    Linux:arm64|Linux:aarch64)
        fail "Wave release archives currently support Linux x86_64 only."
        ;;
    *)
        fail "Unsupported system: $UNAME_OUT $ARCH"
        ;;
esac

WAVE_FILE_NAME="wave-${WAVE_VERSION}-${WAVE_FILE_SUFFIX}.tar.gz"
VEX_FILE_NAME="vex-${VEX_VERSION}-${VEX_FILE_SUFFIX}.tar.gz"
WAVE_URL="https://github.com/${WAVE_REPO}/releases/download/${WAVE_VERSION}/${WAVE_FILE_NAME}"
VEX_URL="https://github.com/${VEX_REPO}/releases/download/${VEX_VERSION}/${VEX_FILE_NAME}"
WAVE_SUMS_URL="https://github.com/${WAVE_REPO}/releases/download/${WAVE_VERSION}/SHA256SUMS"
VEX_SUMS_URL="https://github.com/${VEX_REPO}/releases/download/${VEX_VERSION}/SHA256SUMS"
INSTALL_PARENT="$(dirname "$INSTALL_DIR")"

mkdir -p "$INSTALL_PARENT"
TMP_DIR="$(mktemp -d "$INSTALL_PARENT/.wave-install.XXXXXX")"
STAGE_DIR="${INSTALL_DIR}.new.$$"
BACKUP_DIR="${INSTALL_DIR}.old.$$"
WAVE_EXTRACT_DIR="$TMP_DIR/wave"
VEX_EXTRACT_DIR="$TMP_DIR/vex"

cleanup() {
    rm -rf "$TMP_DIR" "$STAGE_DIR"
}
trap cleanup EXIT

echo "[1/4] Downloading Wave $WAVE_VERSION and Vex $VEX_VERSION..."
echo "[info] Download: $WAVE_URL"
curl -fL "$WAVE_URL" -o "$TMP_DIR/$WAVE_FILE_NAME"
echo "[info] Download: $VEX_URL"
curl -fL "$VEX_URL" -o "$TMP_DIR/$VEX_FILE_NAME"
curl -fsSL "$WAVE_SUMS_URL" -o "$TMP_DIR/WAVE_SHA256SUMS"
curl -fsSL "$VEX_SUMS_URL" -o "$TMP_DIR/VEX_SHA256SUMS"

echo "[2/4] Verifying release archives..."
verify_checksum "$TMP_DIR/$WAVE_FILE_NAME" "$TMP_DIR/WAVE_SHA256SUMS" "$WAVE_FILE_NAME"
verify_checksum "$TMP_DIR/$VEX_FILE_NAME" "$TMP_DIR/VEX_SHA256SUMS" "$VEX_FILE_NAME"

echo "[3/4] Installing Wave toolchain..."
mkdir -p "$WAVE_EXTRACT_DIR" "$VEX_EXTRACT_DIR"
tar -xzf "$TMP_DIR/$WAVE_FILE_NAME" -C "$WAVE_EXTRACT_DIR"
tar -xzf "$TMP_DIR/$VEX_FILE_NAME" -C "$VEX_EXTRACT_DIR"

WAVE_PACKAGE_DIR="$WAVE_EXTRACT_DIR/${WAVE_FILE_NAME%.tar.gz}"
VEX_PACKAGE_DIR="$VEX_EXTRACT_DIR/${VEX_FILE_NAME%.tar.gz}"

[[ -d "$WAVE_PACKAGE_DIR" ]] || fail "Invalid Wave package layout."
[[ -f "$WAVE_PACKAGE_DIR/wavec" ]] || fail "Wave package does not contain wavec."
[[ -d "$WAVE_PACKAGE_DIR/llvm" ]] || fail "Wave package does not contain bundled llvm/."
[[ -d "$VEX_PACKAGE_DIR" ]] || fail "Invalid Vex package layout."
[[ -f "$VEX_PACKAGE_DIR/vex" ]] || fail "Vex package does not contain vex."

rm -rf "$STAGE_DIR" "$BACKUP_DIR"
mkdir -p "$STAGE_DIR"
cp -R "$WAVE_PACKAGE_DIR"/. "$STAGE_DIR"/
cp "$VEX_PACKAGE_DIR/vex" "$STAGE_DIR/vex"
mkdir -p "$STAGE_DIR/share/vex"
for notice in COPYRIGHT LICENSE NOTICE README.md; do
    if [[ -f "$VEX_PACKAGE_DIR/$notice" ]]; then
        cp "$VEX_PACKAGE_DIR/$notice" "$STAGE_DIR/share/vex/$notice"
    fi
done
chmod +x "$STAGE_DIR/wavec" "$STAGE_DIR/vex"
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

rollback_install() {
    rm -rf "$INSTALL_DIR"
    if [[ -d "$BACKUP_DIR" ]]; then
        mv "$BACKUP_DIR" "$INSTALL_DIR"
    fi
}

echo "[4/4] Verifying installation..."
if ! "$INSTALL_DIR/wavec" --version; then
    rollback_install
    fail "wavec installation verification failed."
fi
if ! "$INSTALL_DIR/vex" --version; then
    rollback_install
    fail "vex installation verification failed."
fi

rm -rf "$BACKUP_DIR"
append_path_if_needed

echo "Installation completed successfully."
echo "[info] Installed wavec $WAVE_VERSION and vex $VEX_VERSION."

if [[ "$PATH_CONFIG_UPDATED" -eq 1 ]]; then
    echo "[info] To use 'wavec' and 'vex' in the current terminal, run: $SHELL_RELOAD_COMMAND"
fi
