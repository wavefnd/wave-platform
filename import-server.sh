#!/usr/bin/env bash
set -Eeuo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_root"

if [[ $# -ne 1 ]]; then
  echo "Usage: ./import-server.sh <transfer-archive.tar.gz>" >&2
  exit 1
fi

archive_path="$1"
if [[ ! -f "$archive_path" ]]; then
  echo "Archive not found: $archive_path" >&2
  exit 1
fi

while IFS= read -r entry; do
  case "$entry" in
    .env|data|data/*) ;;
    *)
      echo "Unexpected archive entry: $entry" >&2
      exit 1
      ;;
  esac
done < <(tar -tzf "$archive_path")

if [[ -e .env ]]; then
  echo "Refusing to replace the existing .env. Import into a fresh clone." >&2
  exit 1
fi
if [[ -d data ]] && [[ -n "$(find data -mindepth 1 -print -quit)" ]]; then
  echo "Refusing to replace the existing data directory. Import into a fresh clone." >&2
  exit 1
fi

docker compose down --remove-orphans
tar -xzf "$archive_path" -C "$project_root"
chmod 600 .env
docker compose up -d --build --remove-orphans
docker compose ps
