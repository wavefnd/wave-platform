# Configuration

Copy `.env.example` to `.env` before starting Docker Compose.

## Website and accounts

| Variable | Purpose | Default |
| --- | --- | --- |
| `WAVE_DOMAIN` | Public web domain and Wave email-address domain | `wave-lang.dev` |
| `WAVE_SITE_ADDRESS` | Caddy site address; use the public domain to enable automatic HTTPS | `http://localhost` |
| `WAVE_PUBLIC_URL` | Canonical public URL used in account email links, `robots.txt`, and `sitemap.xml` | `http://localhost:8080` |
| `WAVE_HTTP_PORT` | Host port for HTTP | `80` |
| `WAVE_HTTPS_PORT` | Host port for HTTPS and HTTP/3 | `443` |
| `WAVE_REGISTRATION_OPEN` | Allow public account registration | `true` |
| `WAVE_SESSION_HOURS` | Login session lifetime | `720` |
| `WAVE_SECURE_COOKIES` | Send session cookies only over HTTPS | `false` |
| `WAVE_ADMIN_DISPLAY_NAME` | Initial owner display name | empty |
| `WAVE_ADMIN_USERNAME` | Initial owner username | empty |
| `WAVE_AUTH_ENCRYPTION_KEY` | Base64-encoded 32-byte key used to encrypt TOTP secrets | empty |
| `WAVE_TOTP_ISSUER` | Name shown by authenticator apps | `Wave Platform` |
| `WAVE_ADMIN_RECOVERY_EMAIL` | Initial owner recovery email | empty |
| `WAVE_ADMIN_TOTP_SECRET` | Initial owner Base32 TOTP secret | empty |
| `WAVE_TURNSTILE_SITE_KEY` | Optional Cloudflare Turnstile public site key | empty |
| `WAVE_TURNSTILE_SECRET_KEY` | Optional Cloudflare Turnstile server secret | empty |

Set `WAVE_SECURE_COOKIES=true` for a public HTTPS deployment. Turnstile remains completely disabled unless both keys are set. The platform sends challenge tokens directly to Cloudflare for verification and does not persist them or collect mouse movement, device fingerprints, or other behavioral telemetry.

`start.sh` generates the encryption key in `.env` if it is empty. Generate the initial owner TOTP secret before first start:

```sh
head -c 20 /dev/urandom | base32 | tr -d '=\n'
```

Recovery links expire after 30 minutes and can be used once. Recovery-email confirmation links expire after 24 hours. Replacing or recovering an authenticator revokes existing browser sessions.

Website sign-in uses TOTP and has no account password. Mail is read and sent through Wave Mail. The platform does not issue client passwords or expose authenticated SMTP submission for external mail applications.

`WAVE_PUBLIC_URL` must be an absolute HTTP or HTTPS URL without a query string or fragment. Public routes receive self-referencing canonical metadata and Schema.org JSON-LD. The sitemap is generated from the currently published English documentation, releases, community posts, and questions. Mail, account, authentication, administration, search-result, and writing routes are excluded from indexing; private route responses also carry `X-Robots-Tag`.

## Runtime user

| Variable | Purpose | Default |
| --- | --- | --- |
| `WAVE_PLATFORM_RUNTIME_UID` | UID used to write `./data` | `1000` |
| `WAVE_PLATFORM_RUNTIME_GID` | GID used to write `./data` | `1000` |

Set these values to the owner of the deployment directory when the server uses a different account.

## Mail

| Variable | Purpose | Default |
| --- | --- | --- |
| `WAVE_SMTP_HOSTNAME` | Public hostname announced by the mail server | `mail.wave-lang.dev` |
| `WAVE_SMTP_HOST_PORT` | Host port mapped to SMTP ingress | `2525` |
| `WAVE_SMTP_MAX_MESSAGE_BYTES` | Maximum accepted message size | `26214400` |
| `WAVE_SMTP_TLS_CERTIFICATE` | SMTP TLS certificate path inside the container | empty |
| `WAVE_SMTP_TLS_KEY` | SMTP TLS private-key path inside the container | empty |
| `WAVE_SMTP_RELAY_ADDRESS` | Optional outbound SMTP relay address | empty |
| `WAVE_SMTP_RELAY_USERNAME` | Relay username | empty |
| `WAVE_SMTP_RELAY_PASSWORD` | Relay password | empty |
| `WAVE_SMTP_RELAY_IMPLICIT_TLS` | Use implicit TLS for the relay | `false` |
| `WAVE_SMTP_DIRECT_DELIVERY` | Deliver outbound mail directly through MX records | `true` |
| `WAVE_DKIM_DOMAIN` | Domain used for DKIM signing | empty |
| `WAVE_DKIM_SELECTOR` | Published DKIM selector | empty |
| `WAVE_DKIM_PRIVATE_KEY` | DKIM private-key path inside the container | empty |

Paths under `./data` are visible as `/app/data` in the application container. Store SMTP certificate and DKIM key files under `./data/mail/tls` and reference their `/app/data/mail/tls/...` paths.
