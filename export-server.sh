#!/usr/bin/env bash
set -Eeuo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_root"

if [[ ! -f .env ]]; then
  echo "Missing .env; there is no complete installation to export." >&2
  exit 1
fi

archive_path="${1:-wave-platform-transfer-$(date -u +%Y%m%dT%H%M%SZ).tar.gz}"
if [[ "$archive_path" != /* ]]; then
  archive_path="$project_root/$archive_path"
fi
mkdir -p "$(dirname -- "$archive_path")"

platform_container="$(docker compose ps -q wave-platform)"
platform_was_running=false
if [[ -n "$platform_container" ]]; then
  platform_was_running=true
  docker compose stop wave-platform
fi

restore_service() {
  if [[ "$platform_was_running" == true ]]; then
    docker compose start wave-platform >/dev/null
  fi
}
trap restore_service EXIT

tar -czf "$archive_path" -- .env data
chmod 600 "$archive_path"

restore_service
platform_was_running=false
trap - EXIT

echo "Created $archive_path"
echo "Copy this archive to the new server and run: ./import-server.sh $(basename -- "$archive_path")"
