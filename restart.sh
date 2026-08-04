#!/usr/bin/env bash
set -Eeuo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_root"

if [[ ! -f .env ]]; then
  echo "Missing .env. Copy .env.example to .env and configure it first." >&2
  exit 1
fi

git pull --ff-only
docker compose up -d --build --remove-orphans
docker compose restart caddy
docker compose ps
