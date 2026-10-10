# Wave Translation

An internal translation service implemented in Wave. The foundation contains a
Wave TCP server, a bounded HTTP/1.1 parser, health endpoints, and tooling for
installing and testing a repository-pinned Wave release. There is no Go code or Go
service in this directory. Python is used only for bootstrap, builds, and tests.

Translation, model inference, document chunking, persistent jobs, and platform
integration are not implemented yet. `/readyz` deliberately returns `503` with
`model_not_configured`; no endpoint reports untranslated input as a translation.

## Layout

| Path | Responsibility |
| --- | --- |
| `src/main.wave` | Configuration and service entry point |
| `src/http/request.wave` | Bounded HTTP request parsing |
| `src/http/server.wave` | TCP acceptance, deadlines, and HTTP responses |
| `src/translation/backend.wave` | Current model readiness status |
| `tools/toolchain.py` | Pinned release verification, retention, and installation |
| `tools/build.py` | Build with the locked compiler and its matching standard library |
| `wave-toolchain.lock.json` | Pinned release and archive digest |
| `tests/` | Offline tooling tests and real TCP integration tests |

## Build and run

The checked-in lock targets Linux x86-64. Use Python 3.12 or newer and the
compiler's bundled LLVM and standard library. No global Wave installation is
modified or used. From the repository root:

```sh
# Choose a durable store; this example keeps it in your user data directory.
TRANSLATE_STORE="$HOME/.local/share/wave-translation/snapshots"
TRANSLATE_SDK="$HOME/.local/share/wave-translation/sdk-initial"

python3 translate/tools/toolchain.py \
  --store "$TRANSLATE_STORE" --destination "$TRANSLATE_SDK"
PYTHONDONTWRITEBYTECODE=1 python3 translate/tools/build.py \
  --toolchain "$TRANSLATE_SDK" \
  --build-dir /tmp/wave-translation-build --output /tmp/wave-translation
WAVE_TRANSLATE_PORT=8091 /tmp/wave-translation
```

The install destination must be new. Installation checks the archive
size and SHA-256, safely extracts them, and records the lock used. The build tool
requires that receipt to match the selected lock and explicitly sets `--std-root`.
Do not mix the pinned compiler with a previously installed standard library.

```sh
curl -i http://127.0.0.1:8091/healthz
curl -i http://127.0.0.1:8091/readyz
```

The service binds only `127.0.0.1`; the default port is `8091`. An invalid
`WAVE_TRANSLATE_PORT` exits before listening. Stop the foreground process with
Ctrl-C. This foundation is not connected to the platform's Docker Compose stack.

## HTTP contract

| Request | Status | Meaning |
| --- | --- | --- |
| `GET /healthz` | `200` | HTTP service is running |
| `GET /readyz` | `503` | Translation model is not configured |
| `HEAD` on either route | Same as `GET` | Same content length, no response body |
| Unknown path | `404` | No route |
| Other method without a body | `405` | `Allow: GET, HEAD` |

The parser supports origin-form HTTP/1.1 requests with exactly one nonempty Host
header. Routes are exact; query strings do not match these health endpoints.
Headers are limited to 8 KiB including the request line and terminator, and at
most 100 fields. Duplicate Content-Length fields and ambiguous framing are
rejected. A nonzero Content-Length returns `413`; Transfer-Encoding returns
`501`, or `400` when combined with Content-Length. Other HTTP versions return
`505`. Malformed requests return `400`, oversized headers `431`, and incomplete
requests that exceed the read deadline `408`.

Each connection serves one request and closes. Reads and writes each have a
two-second total deadline, including fragmented reads and partial writes. The
server handles connections sequentially. There is no request-body handling,
keep-alive, TLS, concurrency, or public API yet. The parser is an intentionally
limited health-service implementation, not a general-purpose HTTP server.

## Pinned release

All Wave components now use the version in the root `wave-version` file.
The translation lock pins the release archive's size and SHA-256. Installation
rejects a lock for any other version and never resolves a rolling release.
The service must use the compiler's bundled standard library.

Retain the archive in the selected store to reinstall with `--offline`. The
compiler archive is not committed to Git; normal installation downloads the
exact versioned release. Production deployment remains a separate step.

## Tests

```sh
PYTHONDONTWRITEBYTECODE=1 python3 translate/tests/test_tooling.py
WAVE_TRANSLATE_BINARY=/tmp/wave-translation python3 translate/tests/test_http.py
```

Tooling tests run offline and cover version pinning, missing/corrupted
archives, unsafe extraction, and compiler/standard-library
selection. HTTP tests build no mock server: they launch the Wave binary on a
temporary loopback port and exercise split requests, malformed framing, size
limits, HEAD semantics, deadlines, and connection closure. Local socket access
is required. CI runs the offline tooling tests; native integration currently
requires a retained SDK and the explicit build/test commands above.

## Next implementation steps

- [ ] Add Wave document segmentation and reconstruction, preserving Markdown,
  code, links, and stable segment identifiers.
- [ ] Define an inference adapter, evaluate a pretrained model on long release
  notes, and record its license, revision, and runtime requirements.
- [ ] Add token-budgeted translation jobs, retries, and persistent results keyed
  by source content, target language, model, and glossary revisions.
- [ ] Validate completeness and terminology before publishing machine-translated
  blog content; retain a visible link to the English original.
- [ ] Provide retained snapshot distribution and native integration CI before
  connecting the service to production.
- [ ] Add authenticated internal translation requests, bounded concurrency, and
  operational limits before designing the later public API.
