# Wave Platform

Wave Platform is the official website and community platform for the Wave programming language. It includes the language documentation, email, community posts, questions, release notes, and read-only Git mirrors.

## Run with Docker

Docker Engine with the Compose plugin is required.

```sh
cp .env.example .env
```

Set the public domain, TOTP encryption key, and initial owner values in `.env`, then start the platform:

```sh
./start.sh
```

The local site is available at <http://localhost:8080>. Caddy serves the configured public domain over HTTPS when its DNS records point to the server and the public ports are set to `80` and `443`.

Runtime data is created under `./data` and is not part of the source tree. Back up this directory before upgrading or moving an installation.

## Initial administrator

`start.sh` generates `WAVE_AUTH_ENCRYPTION_KEY` when it is empty. Generate the initial owner TOTP secret:

```sh
head -c 20 /dev/urandom | base32 | tr -d '=\n'
```

Set the generated values and the recovery address in `.env`:

```dotenv
WAVE_ADMIN_DISPLAY_NAME=Wave Administrator
WAVE_ADMIN_USERNAME=wave-admin
WAVE_ADMIN_RECOVERY_EMAIL=owner@example.org
WAVE_ADMIN_TOTP_SECRET=BASE32_SECRET
WAVE_AUTH_ENCRYPTION_KEY=BASE64_KEY
```

Enter `WAVE_ADMIN_TOTP_SECRET` manually in an authenticator app with a 30-second period and six digits. The bootstrap operation is idempotent. Website accounts do not store login passwords. Remove `WAVE_ADMIN_TOTP_SECRET` from `.env` after the owner account exists; keep `WAVE_AUTH_ENCRYPTION_KEY` backed up because it encrypts TOTP secrets at rest.

## Development

The local toolchain requires Go 1.25, Node.js 24, npm, and `wavec` when rebuilding Wave modules.

```sh
make frontend-install
make test
make run
```

The development server listens on <http://127.0.0.1:8080>. Run `make frontend-dev` in another terminal when Vite hot reload is needed.

## Operations

`restart.sh` pulls the current branch with fast-forward only, rebuilds the images, and recreates changed containers:

```sh
./restart.sh
```

To move an installation, export it on the old server and import it into a fresh clone on the new server:

```sh
./export-server.sh
./import-server.sh wave-platform-transfer-YYYYMMDDTHHMMSSZ.tar.gz
```

The transfer archive contains `.env` and the complete `data/` directory. Treat it as a secret.

## Documentation

- [Getting started](docs/getting-started.md)
- [Configuration](docs/configuration.md)
- [Production deployment](docs/deployment.md)
- [Editing language documentation](docs/document-authoring.md)

## License

Wave Platform is licensed under the [Mozilla Public License 2.0](LICENSE). Third-party components retain their respective licenses.
