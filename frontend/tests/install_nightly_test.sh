#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source <(sed -n '1,/^resolve_shell_rc()/p' "$ROOT/public/install.sh" | sed '$d')
for tag in nightly vnightly NIGHTLY vNightly; do
    if message="$(normalize_version "$tag" 2>&1)"; then
        echo "accepted forbidden tag: $tag" >&2; exit 1
    fi
    [[ "$message" == *'manual download'* ]]
    if (validate_version "$tag" >/dev/null 2>&1); then exit 1; fi
done
[[ "$(normalize_version 0.2.1-pre-beta)" == v0.2.1-pre-beta ]]
validate_version v0.2.1-pre-beta
curl() {
    case "${*: -1}" in
        *'page=1') printf '%s\n' '[' '  "tag_name": "nightly",' '  "tag_name": "v0.2.1-pre-beta",' ']' ;;
        *) return 1 ;;
    esac
}
[[ "$(resolve_latest_version wavefnd/Wave)" == v0.2.1-pre-beta ]]
curl() {
    case "${*: -1}" in
        *'page=1') printf '%s\n' '[' '  "tag_name": "nightly",' ']' ;;
        *'page=2') printf '%s\n' '[' '  "tag_name": "v0.2.0-pre-beta",' ']' ;;
        *) return 1 ;;
    esac
}
[[ "$(resolve_latest_version wavefnd/Wave)" == v0.2.0-pre-beta ]]
curl() { printf '[]\n'; }
if (resolve_latest_version wavefnd/Wave >/dev/null 2>&1); then exit 1; fi
curl() { return 22; }
if (resolve_latest_version wavefnd/Wave >/dev/null 2>&1); then exit 1; fi
echo 'installer Nightly policy tests passed'
