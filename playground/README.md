# Wave Playground

The homepage, `/playground`, and documentation Wave fences share one editor and
runner. The HTTP/compiler service in `src/` is Wave. The Vue component and the
small JavaScript Wasm/WASI host connect it to the browser; no Go or Python service
compiles or runs playground programs.

All Wave builds use the root `wave-version` pin and the compiler's bundled std.
The same pin applies to platform native modules, WaveEditor, public installers,
and the translation foundation. Version changes require an explicit source
change and validation, including the translation archive lock and browser API
version check.

## Run

```sh
docker compose up --build -d
```

Open `http://localhost:8080/playground`, or use the example on the homepage.
Documentation code blocks open an editable modal and run the selected code.
Fragments and intentionally invalid examples display compiler diagnostics; they
are not silently wrapped or rewritten. Vite development uses the same Compose
gateway for `/api/playground/compile`.

`POST /api/playground/compile` accepts a UTF-8 `text/plain` body with a
Content-Length. A successful response is `application/wasm` with
`X-Wave-Version`. Compile errors return `422` with the compiler's structured JSON
diagnostic lines; resource failures return `503`. Invalid framing and content
types are rejected. The server handles one connection/job at a time and closes
each connection. Request reads have a three-second total deadline.

## Execution boundaries

- Compilation takes place in a separate unprivileged, read-only container, on
  an internal network. It has no platform data, secrets, Docker socket, or host
  filesystem mounts. Only Caddy can reach the service through Compose networking.
- `/work` is a 64 MiB, non-executable tmpfs. Each job is removed before and after
  use. The service uses fixed argv entries, without a command shell. Guest
  programs are never executed natively on the server.
- Compiler limits: 32 KiB source, 10 seconds wall time plus a one-second kill
  grace period, eight seconds CPU, 768 MiB address space, 4 MiB output files, and
  64 file descriptors. Container limits: one CPU, 1 GiB memory, 64 processes,
  no extra capabilities, and no privilege escalation.
- Output targets `wasm32-wasip1`. The bundled LLD is exposed as `wasm-ld`; no
  ambient system LLVM installation is required.
- A fresh browser Worker executes each result, with a three-second deadline,
  a 64 MiB Wasm memory ceiling, and a 64 KiB output cap. Stop terminates the Worker
  and aborts the browser's compile request. A compile already accepted on the
  server finishes within its own limits before the next job is handled.
- The browser host provides bounded stdin/stdout/stderr (output streams are
  combined), empty arguments/environment, clocks, and random bytes. Input is
  supplied before running, followed by EOF. No filesystem, sockets, or arbitrary
  native FFI are exposed. Missing host functions fail explicitly.
- The gateway permits `wasm-unsafe-eval` for Wasm compilation, without permitting
  JavaScript `eval` or inline scripts. Editor content and program output are
  rendered as text.

This is a single-file teaching playground. Multi-file packages, interactive
terminal sessions, native libraries, and OS-specific programs require a local
toolchain. It is not a replacement for a local Wave installation.

## Validation

```sh
python3 tools/check-wave-version.py
python3 playground/tests/test_server.py
node --experimental-strip-types --test frontend/tests/playground-live.test.ts
cd frontend && npm run test:playground && npm run build
```

Integration tests use the live Compose gateway by default; `PLAYGROUND_URL` can
point at the service for direct HTTP checks. `test_http.py` requires that direct
service URL and checks malformed framing, split bodies, and request deadlines.
CI builds the same container and
runs these checks. Browser checks cover homepage execution, documentation modal
execution, stdin, diagnostics, cancellation, deadline termination, and mobile
layout.

The float-formatting helper in the browser host is adapted from the matching
Wave release's `src/runtime/wasm_host.mjs` under MPL-2.0.
