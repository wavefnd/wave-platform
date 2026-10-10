# Wave Platform

Wave Platform is the official website and community platform for the Wave programming language. It includes the language documentation, email, community posts, questions, RFC design proposals, release notes, and read-only Git mirrors.

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

Wave components use the version pinned in `wave-version`, with the matching
bundled standard library. `make build` and `make run` reject another compiler.
The homepage, documentation examples, and `/playground` use the Wave compiler
service included in Docker Compose. See [Playground](playground/README.md) for
its execution limits and integration checks.

Wave services share the small [HTTP library](http/README.md). Playground and
Translation remain independent processes with their own routes and lifecycle;
the Go platform remains a single host application.

### LLVM toolchain downloads

The `/toolchains` page reads `/downloads/toolchains/index.json`. In development,
the Go server serves this path from `./toolchains`, the same public directory
mounted by Caddy, so Vite's existing `/downloads` proxy works without Caddy.
Only this public directory is served; hidden files, directory listings, and
symlinks escaping it are rejected. Development responses use `no-cache`.
Production downloads continue to be served by Caddy's read-only mount.
Verified archives, checksums, metadata, and the catalog are committed together
under `toolchains/`. A Git checkout or pull supplies the download files; no
separate upload is required. They are excluded only from Docker build contexts
because Caddy serves the checked-out directory through its read-only mount.

Published files use the layout
`llvm/<version>/<revision>/wave-llvm-<version>-<target>-<revision>.tar.xz`,
with an adjacent `.tar.xz.sha256` file containing the SHA-256 digest and archive
basename. The catalog uses `schema_version: 1` and a `bundles` array; each entry
contains `target`, `llvm_version`, `revision`, `filename`, `size_bytes`, `sha256`,
`url`, and `published_at`. URLs use `https://wave-lang.dev/downloads/toolchains/`
followed by the relative archive path. Supported targets are `linux-riscv64`
and `linux-loong64`; versions must be LLVM 21 and revisions start at `r1`.

Commit verified archives and checksums together with their catalog entries.
Keep published revision paths immutable and deploy through `./restart.sh`.
An absent catalog or an empty `bundles` array displays the empty state; no SDK
or placeholder entry is required while builds are in progress.
Run `cd frontend && npm run test:toolchains` to check catalog and checksum behavior.

## Android development

The native Kotlin and Jetpack Compose app lives in [`android/`](android/README.md).
Its contributor guide covers the pinned toolchain, local server connection, debug
installation, and tests.

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

The transfer archive contains `.env`, the complete `data/` directory, and
`toolchains/` when present. Treat it as a secret.

## Documentation

- [Getting started](docs/getting-started.md)
- [Configuration](docs/configuration.md)
- [Production deployment](docs/deployment.md)
- [Editing language documentation](docs/document-authoring.md)

## License

Wave Platform is licensed under the [Mozilla Public License 2.0](LICENSE). Third-party components retain their respective licenses.
