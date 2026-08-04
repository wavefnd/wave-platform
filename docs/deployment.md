# Production deployment

## DNS

Create records for the website and mail server before starting the public instance.

| Name | Type | Value |
| --- | --- | --- |
| `example.org` | `A` / `AAAA` | Server address |
| `mail.example.org` | `A` / `AAAA` | Server address |
| `example.org` | `MX` | `mail.example.org` |

Publish SPF, DKIM, and DMARC records before sending external mail. Reverse DNS for the server address should resolve to the SMTP hostname.

## Environment

```sh
cp .env.example .env
```

At minimum, set:

```dotenv
WAVE_DOMAIN=example.org
WAVE_SITE_ADDRESS=example.org
WAVE_PUBLIC_URL=https://example.org
WAVE_HTTP_PORT=80
WAVE_HTTPS_PORT=443
WAVE_SECURE_COOKIES=true
WAVE_ADMIN_DISPLAY_NAME=Platform Owner
WAVE_ADMIN_USERNAME=owner
WAVE_ADMIN_RECOVERY_EMAIL=owner@example.net
WAVE_ADMIN_TOTP_SECRET=BASE32_SECRET
WAVE_AUTH_ENCRYPTION_KEY=BASE64_KEY

WAVE_SMTP_HOSTNAME=mail.example.org
WAVE_SMTP_HOST_PORT=25
```

See [Configuration](configuration.md) for relay, TLS, and DKIM settings.

`WAVE_PUBLIC_URL` is also the canonical origin advertised by the generated sitemap and robots policy, so it must match the HTTPS site address used by visitors.

Generate `WAVE_ADMIN_TOTP_SECRET` with the command in [Configuration](configuration.md). `start.sh` fills an empty `WAVE_AUTH_ENCRYPTION_KEY` with a random 32-byte key; store the resulting `.env` in the server secret backup. After the owner has signed in and confirmed the recovery address, remove the bootstrap TOTP secret from `.env`. Cloudflare Turnstile is optional; if enabled, configure both Turnstile keys for the production hostname.

## Start

```sh
./start.sh
```

Caddy obtains and renews the website certificate automatically. Ports `80`, `443`, and SMTP port `25` must be reachable from the Internet. SMTP TLS is configured separately because Caddy does not terminate the mail protocol. Users access mail through Wave Mail; no client submission port is exposed.

With the production port values above, the site remains available at <http://localhost> on the server for local checks. The default development ports use <http://localhost:8080> instead.

## Data and backups

All persistent platform data is stored under `./data`. Caddy certificate state is stored in the `caddy-data` Docker volume.

Stop write traffic or stop the application before taking a filesystem-level backup:

```sh
docker compose stop wave-platform
```

Back up `./data` and the Caddy volumes, then restart the service:

```sh
docker compose start wave-platform
```

## Update

```sh
./restart.sh
```

Check `docker compose ps` and `/health/ready` after the containers become healthy.

## Move to another server

On the old server:

```sh
./export-server.sh
```

Copy the generated archive to a fresh clone on the new server, then run:

```sh
./import-server.sh wave-platform-transfer-YYYYMMDDTHHMMSSZ.tar.gz
```

The export briefly stops the application to produce a consistent copy of `.env` and `data/`. The import refuses to overwrite an existing installation. Caddy obtains a new website certificate after the new server starts.
