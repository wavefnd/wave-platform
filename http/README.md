# Wave HTTP

A small HTTP/1.1 server library written in Wave. Playground and Translation
import the same source through `--dep wave_http=<path>/http`. It is a library,
not another running service. It uses the compiler pinned in `../wave-version`
and its bundled standard library.

## Use

From the repository root:

```sh
python3 tools/check-wave-version.py
wavec --dep wave_http="$PWD/http" build http/examples/hello.wave \
  --target-dir /tmp/wave-http-build -o /tmp/wave-http-example
/tmp/wave-http-example
```

The example listens on `127.0.0.1:8093` (`WAVE_HTTP_PORT` overrides the port).
`GET /` returns text, `POST /echo` echoes up to 4 KiB of binary data, and
`GET /empty` demonstrates a bodyless response. No package download is required.
The same dependency flag works for a service outside this repository.

```wave
import("wave_http::request")::{Request, method_is, target_is};
import("wave_http::connection")::{read_request, write_response, close_connection};
```

## API

- `read_request(stream, buffer, capacity, max_body, timeout_ms)` reads one
  request. Reserve at least `8192 + max_body` bytes. Status `200` means a
  complete request, `0` means disconnect, and other statuses are HTTP errors.
  This status is a parsing result; the service chooses its response status.
- `Request` contains method, target, content type, and body offsets/lengths into
  the caller's buffer. Keep that buffer alive and unchanged while handling it.
  The body starts at `body_start` and has `length` bytes; it may contain NULs.
- `method_is`, `target_is`, and `content_type_is` compare those slices. Targets
  are exact, including query strings. No URL decoding or routing is implicit.
- `parse_request(buffer, header_size, max_body)` parses only a complete request
  head, without sockets or body reads.
- `write_response(stream, status, content_type, extra_headers, body, size, head)`
  writes a final `200..599` response with a two-second total write deadline.
  HEAD preserves the representation length without sending a body. `204` and
  `304` send neither a body nor Content-Length. Supply trusted constants for
  content type and extra headers; extra headers must be empty or CRLF-terminated
  and must not override framing or connection headers. Never copy raw client
  input into them. Body data is byte-counted.
- `close_connection(stream)` shuts down writes, briefly drains unread bytes,
  then closes the connection so rejected bodies do not cause proxy resets.

## Boundaries

The library handles HTTP framing, not application policy. Each service owns
its listener, routes, content types, body limit, and lifecycle. Playground stays
in its isolated container; Translation stays a separate internal executable.
The existing Go platform host remains in place. Services can be deployed
separately and reuse this library without calling through Go.

Headers are limited to 8 KiB and 100 fields. Duplicate Host/Content-Length,
ambiguous framing, folded headers, control bytes, and invalid lengths are
rejected. HTTP/1.1 and Content-Length bodies are supported. Chunked requests,
Expect/100-continue, keep-alive, TLS, streaming, and HTTP client functionality
are outside this initial library. TLS remains at the gateway. Responses use
`Connection: close` and `Cache-Control: no-store`.

## Tests

```sh
wavec --dep wave_http="$PWD/http" build http/tests/request.wave \
  --target-dir /tmp/wave-http-tests-build -o /tmp/wave-http-tests
/tmp/wave-http-tests
python3 http/tests/test_server.py
```

Build the example first; the Python suite starts and stops its own copy on a
temporary loopback port. CI also tests both real consumer services, including
fragmented requests, timeouts, HEAD errors, compiler output, and browser Wasm
execution.
