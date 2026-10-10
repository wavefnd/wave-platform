#!/usr/bin/env bash
set -euo pipefail
# Wave installer channel policy: pinned-0.2.1-v1

info() { printf '[info] %s\n' "$*"; }
fail() { printf '[error] %s\n' "$*" >&2; exit 1; }
manual_only() {
    fail 'This installer installs Wave v0.2.1-pre-beta. Install older versions or Nightly manually: https://github.com/wavefnd/Wave/releases'
}
usage() {
    cat <<'HELP'
Wave Toolchain Installer — Wave v0.2.1-pre-beta
Usage: bash install.sh [latest] [--with-vex | --without-vex] [--no-modify-path]
  Default: install Wave and install Vex when its latest release supports this platform.
  --with-vex       Require Vex; fail before installation if its package is absent.
  --without-vex    Install Wave only.
  --no-modify-path Leave shell configuration unchanged.
  WAVE_INSTALL_DIR overrides the dedicated installation directory (~/.wave/bin).
Older versions and Nightly: download manually from https://github.com/wavefnd/Wave/releases
HELP
}
fetch() {
    curl --proto '=https' --proto-redir '=https' --tlsv1.2 -fLsS \
        --connect-timeout 20 --max-time 600 --retry 3 --retry-delay 2 "$@"
}
latest_release() {
    local repository="$1" page=1 response candidate best='null'
    while :; do
        response="$(fetch -H 'Accept: application/vnd.github+json' \
            "https://api.github.com/repos/$repository/releases?per_page=100&page=$page")" \
            || fail "Cannot query $repository releases. Check connectivity or GitHub API limits."
        jq -e 'type == "array"' >/dev/null <<< "$response" || fail "Invalid release response from $repository."
        [[ "$(jq 'length' <<< "$response")" != 0 ]] || break
        candidate="$(jq -c '[.[] | select(.draft == false and (.published_at | type == "string")) |
            select(.tag_name | test("^v[0-9]+\\.[0-9]+\\.[0-9]+(-[0-9A-Za-z.-]+)?$"))] |
            sort_by(.published_at, .id) | last // null' <<< "$response")"
        best="$(jq -cn --argjson a "$best" --argjson b "$candidate" \
            '[$a,$b] | map(select(. != null)) | sort_by(.published_at, .id) | last // null')"
        [[ "$(jq 'length' <<< "$response")" == 100 ]] || break
        page=$((page + 1))
    done
    [[ "$best" != null ]] || fail "No public versioned release is available for $repository."
    printf '%s\n' "$best"
}
pinned_wave_release() {
    local release
    release="$(fetch -H 'Accept: application/vnd.github+json' 'https://api.github.com/repos/wavefnd/Wave/releases/tags/v0.2.1-pre-beta')" || fail 'Cannot fetch pinned Wave release.'
    jq -e '.tag_name == "v0.2.1-pre-beta" and .draft == false' >/dev/null <<< "$release" || fail 'Pinned Wave release metadata does not match.'
    printf '%s\n' "$release"
}
asset() {
    # An absent optional Vex asset is distinct from an invalid published asset.
    local release="$1" name="$2" optional="${3:-false}" selected count
    selected="$(jq -c --arg name "$name" '[.assets[] | select(.name == $name)]' <<< "$release")"
    count="$(jq 'length' <<< "$selected")"
    if [[ "$count" == 0 && "$optional" == true ]]; then printf 'null\n'; return; fi
    [[ "$count" == 1 ]] || fail "Latest release has no unique package: $name. No older version will be selected."
    jq -e '.[0] | .state == "uploaded" and (.digest | type == "string" and test("^sha256:[0-9A-Fa-f]{64}$"))' \
        >/dev/null <<< "$selected" || fail "No valid GitHub SHA-256 was published for $name."
    jq -c '.[0]' <<< "$selected"
}
sha256_file() {
    if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
    elif command -v sha256 >/dev/null 2>&1; then sha256 -q "$1"
    else fail 'SHA-256 verification requires sha256sum, shasum or sha256.'; fi
}
verify_hash() {
    local actual expected="$2"
    [[ "$expected" =~ ^[0-9A-Fa-f]{64}$ ]] || fail "Invalid SHA-256 for $1."
    actual="$(sha256_file "$1")"
    [[ "$(tr '[:upper:]' '[:lower:]' <<< "$actual")" == "$(tr '[:upper:]' '[:lower:]' <<< "$expected")" ]] \
        || fail "SHA-256 verification failed: ${1##*/}"
}
resolve_platform() {
    local os="$1" arch="$2"
    case "$os:$arch" in
        Linux:x86_64|Linux:amd64) WAVE_TARGET=x86_64-unknown-linux-gnu; WAVE_SUFFIX=x86_64-linux-gnu; VEX_SUFFIX=x86_64-unknown-linux-gnu ;;
        Linux:aarch64|Linux:arm64) WAVE_TARGET=aarch64-unknown-linux-gnu; WAVE_SUFFIX=aarch64-linux-gnu; VEX_SUFFIX=aarch64-unknown-linux-gnu ;;
        Linux:riscv64) WAVE_TARGET=riscv64-unknown-linux-gnu; WAVE_SUFFIX=riscv64-linux-gnu; VEX_SUFFIX=riscv64gc-unknown-linux-gnu ;;
        Linux:loongarch64|Linux:loong64) WAVE_TARGET=loongarch64-unknown-linux-gnu; WAVE_SUFFIX=loongarch64-linux-gnu; VEX_SUFFIX=loongarch64-unknown-linux-gnu ;;
        Darwin:arm64|Darwin:aarch64) WAVE_TARGET=aarch64-apple-darwin; WAVE_SUFFIX="$WAVE_TARGET"; VEX_SUFFIX="$WAVE_TARGET" ;;
        Darwin:x86_64|Darwin:amd64) WAVE_TARGET=x86_64-apple-darwin; WAVE_SUFFIX="$WAVE_TARGET"; VEX_SUFFIX="$WAVE_TARGET" ;;
        FreeBSD:amd64|FreeBSD:x86_64) WAVE_TARGET=x86_64-unknown-freebsd; WAVE_SUFFIX="$WAVE_TARGET"; VEX_SUFFIX="$WAVE_TARGET" ;;
        *) fail "Unsupported system: $os $arch" ;;
    esac
}
configure_path() {
    local shell_name="${SHELL:-}" rc line quoted
    shell_name="${shell_name##*/}"
    # Single-quote literal paths rather than injecting shell syntax into the rc file.
    quoted="$(printf '%s' "$INSTALL_DIR" | sed "s/'/'\\\\''/g")"
    case "$shell_name" in
        bash) rc="$HOME/.bashrc"; line="export PATH='$quoted':\$PATH" ;;
        zsh) rc="${ZDOTDIR:-$HOME}/.zshrc"; line="export PATH='$quoted':\$PATH" ;;
        fish)
            rc="$HOME/.config/fish/config.fish"
            quoted="$(printf '%s' "$INSTALL_DIR" | sed "s/\\\\/\\\\\\\\/g; s/'/\\\\'/g")"
            line="fish_add_path '$quoted'" ;;
        *) rc="$HOME/.profile"; line="export PATH='$quoted':\$PATH" ;;
    esac
    mkdir -p "$(dirname "$rc")" || return 1
    touch "$rc" || return 1
    if ! grep -Fxq "$line" "$rc"; then printf '\n# Wave\n%s\n' "$line" >> "$rc" || return 1; fi
    info "PATH configured in $rc. Open a new terminal to use Wave."
}
verify_installation() {
    local directory="$1" target="$2" with_vex="$3" work="$4" actual
    "$directory/wavec" --version || return 1
    actual="$("$directory/wavec" print target-spec --format=json | jq -er .triple)" || return 1
    [[ "$actual" == "$target" ]] || { printf 'Expected %s; compiler reports %s\n' "$target" "$actual" >&2; return 1; }
    cat > "$work/install-smoke.wave" <<'WAVE'
import("std::mem::layout")::{size_of};
fun main() -> i32 {
    if (size_of<i64>() != 8) { return 1; }
    return 0;
}
WAVE
    # Keep output outside the user's working directory and ignore ambient std overrides.
    (cd "$work" && "$directory/wavec" run install-smoke.wave --std-root "$directory/std") || return 1
    if [[ "$with_vex" == true ]]; then "$directory/vex" --version || return 1; fi
}
main() {
    local vex_mode=auto modify_path=true os arch dependency
    while [[ $# -gt 0 ]]; do
        case "$1" in
            latest|--latest) ;;
            --with-vex) [[ "$vex_mode" != off ]] || fail 'Conflicting Vex options.'; vex_mode=required ;;
            --without-vex) [[ "$vex_mode" != required ]] || fail 'Conflicting Vex options.'; vex_mode=off ;;
            --no-modify-path) modify_path=false ;;
            -h|--help) usage; return ;;
            --version*|--wave-version*|--vex-version*|nightly|vnightly|v[0-9]*|[0-9]*) manual_only ;;
            *) fail "Unknown option: $1. Run with --help." ;;
        esac
        shift
    done
    [[ -z "${WAVE_VERSION:-}${VEX_VERSION:-}" ]] || manual_only
    for dependency in curl jq tar awk sed; do command -v "$dependency" >/dev/null 2>&1 || fail "Install $dependency and retry."; done
    os="$(uname -s)"; arch="$(uname -m)"
    if [[ "$os" == Darwin && "$arch" == x86_64 ]] && [[ "$(sysctl -in sysctl.proc_translated 2>/dev/null || true)" == 1 ]]; then arch=arm64; fi
    resolve_platform "$os" "$arch"
    INSTALL_DIR="${WAVE_INSTALL_DIR:-$HOME/.wave/bin}"
    INSTALL_DIR="${INSTALL_DIR%/}"
    [[ "$INSTALL_DIR" != *:* && "$INSTALL_DIR" == /* && "$INSTALL_DIR" != / && "$INSTALL_DIR" != "$HOME" && "$INSTALL_DIR" != *$'\n'* && "$INSTALL_DIR" != *$'\r'* ]] || fail 'Use an absolute, dedicated installation directory.'
    case "${INSTALL_DIR##*/}" in .|..) fail 'Use a dedicated installation directory, not . or ..' ;; esac
    case "$INSTALL_DIR" in /bin|/sbin|/usr/bin|/usr/sbin|/usr/local/bin|/usr/local/sbin) fail 'Do not replace a shared system binary directory.' ;; esac
    [[ ! -L "$INSTALL_DIR" ]] || fail 'The installation directory must not be a symbolic link.'
    if [[ -e "$INSTALL_DIR" && ! -f "$INSTALL_DIR/wavec" ]]; then fail "Refusing to replace a directory not managed by Wave: $INSTALL_DIR"; fi
    local wave_release wave_version wave_name wave_asset vex_release vex_version='' vex_name='' vex_asset='null' install_vex=false
    wave_release="$(pinned_wave_release)"; wave_version="$(jq -r '.tag_name' <<< "$wave_release")"
    wave_name="wave-$wave_version-$WAVE_SUFFIX.tar.gz"; wave_asset="$(asset "$wave_release" "$wave_name")"
    if [[ "$vex_mode" != off ]]; then
        vex_release="$(latest_release wavefnd/Vex)"; vex_version="$(jq -r '.tag_name' <<< "$vex_release")"
        vex_name="vex-$vex_version-$VEX_SUFFIX.tar.gz"; vex_asset="$(asset "$vex_release" "$vex_name" true)"
        if [[ "$vex_asset" == null ]]; then
            [[ "$vex_mode" != required ]] || fail "Latest Vex has no package for $WAVE_TARGET."
            info "Latest Vex has no package for $WAVE_TARGET; installing Wave only."
        else install_vex=true; fi
    fi
    info "Wave $wave_version / $WAVE_TARGET"
    info "Install directory: $INSTALL_DIR"
    local parent
    parent="$(dirname "$INSTALL_DIR")"; mkdir -p "$parent"
    LOCK_DIR="$INSTALL_DIR.install-lock"
    mkdir "$LOCK_DIR" 2>/dev/null || fail "Another install may be running ($LOCK_DIR). If it was interrupted, inspect the directory before retrying."
    TMP_DIR=''; ACTIVATING=false; COMMITTED=false
    trap cleanup EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    TMP_DIR="$(mktemp -d "$parent/.wave-install.XXXXXX")"
    info '[1/4] Downloading and verifying packages'
    fetch "https://github.com/wavefnd/Wave/releases/download/$wave_version/$wave_name" -o "$TMP_DIR/wave.tar.gz"
    verify_hash "$TMP_DIR/wave.tar.gz" "$(jq -r '.digest | sub("^sha256:"; "")' <<< "$wave_asset")"
    if [[ "$install_vex" == true ]]; then
        fetch "https://github.com/wavefnd/Vex/releases/download/$vex_version/$vex_name" -o "$TMP_DIR/vex.tar.gz"
        verify_hash "$TMP_DIR/vex.tar.gz" "$(jq -r '.digest | sub("^sha256:"; "")' <<< "$vex_asset")"
    fi
    info '[2/4] Preparing installation'
    mkdir "$TMP_DIR/wave" "$TMP_DIR/stage"
    tar -xzf "$TMP_DIR/wave.tar.gz" -C "$TMP_DIR/wave"
    local package="$TMP_DIR/wave/${wave_name%.tar.gz}"
    [[ -f "$package/wavec" && -d "$package/llvm/bin" && -f "$package/std/manifest.json" ]] || fail 'Wave package is missing compiler, LLVM or bundled std.'
    cp -R "$package"/. "$TMP_DIR/stage/"
    if [[ "$install_vex" == true ]]; then
        mkdir "$TMP_DIR/vex"; tar -xzf "$TMP_DIR/vex.tar.gz" -C "$TMP_DIR/vex"
        package="$TMP_DIR/vex/${vex_name%.tar.gz}"
        [[ -f "$package/vex" ]] || fail 'Vex package is missing vex.'
        cp "$package/vex" "$TMP_DIR/stage/vex"
        mkdir -p "$TMP_DIR/stage/share/vex"
        local notice
        for notice in COPYRIGHT LICENSE NOTICE README.md; do
            if [[ -f "$package/$notice" ]]; then cp "$package/$notice" "$TMP_DIR/stage/share/vex/"; fi
        done
    fi
    info '[3/4] Activating installation'
    if [[ -d "$INSTALL_DIR" ]]; then mv "$INSTALL_DIR" "$TMP_DIR/previous"; fi
    ACTIVATING=true
    mv "$TMP_DIR/stage" "$INSTALL_DIR"
    info '[4/4] Checking compiler, bundled std and runtime'
    verify_installation "$INSTALL_DIR" "$WAVE_TARGET" "$install_vex" "$TMP_DIR" \
        || fail 'Installation check failed; restoring the previous installation. Check system prerequisites (glibc/system libraries, Apple Command Line Tools, or a compatible FreeBSD base).'
    COMMITTED=true
    if [[ "$modify_path" == true ]]; then
        configure_path || printf '[warning] Wave is installed, but PATH could not be configured. Add %s to PATH manually.\n' "$INSTALL_DIR" >&2
    fi
    info "Installed Wave $wave_version."
    if [[ "$install_vex" == true ]]; then info "Installed Vex $vex_version."; fi
}
cleanup() {
    local status=$?
    trap - EXIT
    if [[ "$ACTIVATING" == true && "$COMMITTED" != true ]]; then
        rm -rf "$INSTALL_DIR"
    fi
    if [[ "$COMMITTED" != true && -n "$TMP_DIR" && -d "$TMP_DIR/previous" ]]; then
        if ! mv "$TMP_DIR/previous" "$INSTALL_DIR"; then
            printf '[error] Restore failed. Previous installation retained at %s/previous\n' "$TMP_DIR" >&2
            rmdir "$LOCK_DIR" 2>/dev/null || true
            exit 1
        fi
    fi
    if [[ -n "$TMP_DIR" ]]; then rm -rf "$TMP_DIR"; fi
    rmdir "$LOCK_DIR" 2>/dev/null || true
    exit "$status"
}
# Sourcing exposes helpers to tests without starting an installation.
if [[ "${BASH_SOURCE[0]:-}" == "$0" || -z "${BASH_SOURCE[0]:-}" ]]; then main "$@"; fi
