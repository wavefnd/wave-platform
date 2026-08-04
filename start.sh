#!/usr/bin/env bash
set -Eeuo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_root"

if [[ ! -f .env ]]; then
  cp .env.example .env
  chmod 600 .env
  echo "Created .env. Review the domain and initial owner settings before a public launch."
fi

auth_key="$(awk -F= '$1 == "WAVE_AUTH_ENCRYPTION_KEY" {sub(/^[^=]*=/, ""); print; exit}' .env)"
if [[ -z "$auth_key" ]]; then
  auth_key="$(openssl rand -base64 32 | tr -d '\n')"
  temporary_env="$(mktemp)"
  awk -v value="$auth_key" '
    /^WAVE_AUTH_ENCRYPTION_KEY=/ { print "WAVE_AUTH_ENCRYPTION_KEY=" value; found=1; next }
    { print }
    END { if (!found) print "WAVE_AUTH_ENCRYPTION_KEY=" value }
  ' .env > "$temporary_env"
  chmod 600 "$temporary_env"
  mv "$temporary_env" .env
  echo "Generated WAVE_AUTH_ENCRYPTION_KEY in .env. Keep this file in the server backup."
fi

docker compose up -d --build --remove-orphans
docker compose restart caddy
docker compose ps
