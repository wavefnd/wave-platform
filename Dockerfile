# syntax=docker/dockerfile:1

FROM node:24-bookworm-slim AS frontend-builder

WORKDIR /src

COPY editor/ ./editor/

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json* ./

RUN if [ -f package-lock.json ]; then \
        npm ci; \
    else \
        npm install; \
    fi

COPY frontend/ ./
COPY wavedoc/redirects.json wavedoc/locales.json /src/wavedoc/

RUN npm run build


FROM golang:1.25-trixie AS wave-toolchain

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        clang \
        cmake \
        curl \
        git \
        jq \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY wave-version /opt/wave-version
ARG VEX_VERSION=0.0.1

# All platform Wave components use wave-version and the bundled standard library.
RUN <<'INSTALL'
set -eu
WAVE_VERSION="$(cat /opt/wave-version)"
case "$(uname -m)" in
    x86_64) arch=x86_64 ;;
    aarch64) arch=aarch64 ;;
    *) echo 'Unsupported server toolchain architecture' >&2; exit 1 ;;
esac
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p /root/.wave/bin/share/vex
for product in Wave Vex; do
    if [ "$product" = Wave ]; then
        tag="v${WAVE_VERSION#v}"; name="wave-$tag-$arch-linux-gnu.tar.gz"
    else
        tag="v${VEX_VERSION#v}"; name="vex-$tag-$arch-unknown-linux-gnu.tar.gz"
    fi
    curl --proto '=https' --tlsv1.2 -fsSL --retry 3 \
        "https://api.github.com/repos/wavefnd/$product/releases/tags/$tag" > "$work/release.json"
    digest="$(jq -er --arg name "$name" \
        '[.assets[] | select(.name == $name and .state == "uploaded")] | if length == 1 then .[0].digest else error("missing asset") end | strings | select(test("^sha256:[0-9a-fA-F]{64}$")) | sub("^sha256:"; "")' "$work/release.json")"
    curl --proto '=https' --proto-redir '=https' --tlsv1.2 -fsSL --retry 3 \
        "https://github.com/wavefnd/$product/releases/download/$tag/$name" -o "$work/$name"
    printf '%s  %s\n' "$digest" "$work/$name" | sha256sum -c -
    tar -xzf "$work/$name" -C "$work"
    package="$work/${name%.tar.gz}"
    if [ "$product" = Wave ]; then
        test -f "$package/wavec"; test -d "$package/llvm"
        cp -R "$package"/. /root/.wave/bin/
    else
        cp "$package/vex" /root/.wave/bin/vex
        for notice in COPYRIGHT LICENSE NOTICE README.md; do
            if [ -f "$package/$notice" ]; then cp "$package/$notice" /root/.wave/bin/share/vex/; fi
        done
    fi
done
INSTALL

ENV PATH="/root/.wave/bin:${PATH}"

RUN wavec --version
RUN vex --version

FROM wave-toolchain AS playground-builder
WORKDIR /src
COPY playground/src ./playground/src
RUN wavec --std-root /root/.wave/bin/std build playground/src/main.wave \
    --target-dir /tmp/playground-build -o /tmp/wave-playground

FROM debian:trixie-slim AS playground
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl util-linux \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10002 playground \
    && useradd --uid 10002 --gid playground --no-create-home playground
COPY --from=playground-builder /root/.wave/bin /opt/wave
RUN ln -s ld.lld /opt/wave/llvm/bin/wasm-ld
COPY --from=playground-builder /tmp/wave-playground /usr/local/bin/wave-playground
WORKDIR /work
USER 10002:10002
EXPOSE 8092
CMD ["/usr/local/bin/wave-playground"]

FROM wave-toolchain AS application-builder

WORKDIR /src

COPY go.mod go.sum* ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY wavedoc/ ./wavedoc/
COPY config/ ./config/
COPY schemas/ ./schemas/
COPY native/ ./native/
COPY wave/ ./wave/
COPY editor/ ./editor/

COPY --from=frontend-builder /src/frontend/dist ./web/dist

RUN if [ -f native/CMakeLists.txt ]; then \
        cmake -S native -B build/native \
            -DCMAKE_BUILD_TYPE=Release \
            -DBUILD_SHARED_LIBS=OFF \
        && cmake --build build/native --parallel; \
    else \
        mkdir -p build/native; \
    fi

RUN mkdir -p build/wave \
    && wavec build wave/policy-engine/main.wave \
        --emit=obj \
        -o build/wave/policy-engine.o \
    && wavec build wave/media-policy/main.wave \
        --emit=obj \
        -o build/wave/media-policy.o \
    && wavec build wave/source-analyzer/main.wave \
        --emit=obj \
        -o build/wave/source-analyzer.o \
    && wavec build editor/wave/main.wave \
        --emit=obj \
        -o build/wave/editor.o \
    && cc -shared \
        -Wl,-soname,libwave-media-policy.so \
        -o build/wave/libwave-media-policy.so \
        build/wave/media-policy.o \
    && cc -shared \
        -Wl,-soname,libwave-source-analyzer.so \
        -o build/wave/libwave-source-analyzer.so \
        build/wave/source-analyzer.o \
    && cc -shared \
        -Wl,-soname,libwave-editor.so \
        -o build/wave/libwave-editor.so \
        build/wave/editor.o \
    && cc -Inative/include \
        native/tests/source_analyzer_smoke.c \
        build/wave/source-analyzer.o \
        -o build/wave/source-analyzer-smoke \
    && build/wave/source-analyzer-smoke \
    && rm -f build/wave/source-analyzer-smoke \
    && cc -Inative/include \
        native/tests/media_policy_smoke.c \
        build/wave/media-policy.o \
        -o build/wave/media-policy-smoke \
    && build/wave/media-policy-smoke \
    && rm -f build/wave/media-policy-smoke \
    && cp wave/policy-engine/module.xml build/wave/policy-engine.xml \
    && cp wave/media-policy/module.xml build/wave/media-policy.xml \
    && cp wave/source-analyzer/module.xml build/wave/source-analyzer.xml \
    && cp editor/module.xml build/wave/editor.xml

RUN CGO_ENABLED=1 \
    go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /src/build/wave-platform \
        ./cmd/server


FROM debian:trixie-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system --gid 10001 wave \
    && useradd \
        --system \
        --uid 10001 \
        --gid wave \
        --home-dir /app \
        --shell /usr/sbin/nologin \
        wave

WORKDIR /app

COPY --from=application-builder /src/build/wave-platform ./wave-platform
COPY --from=application-builder /src/build/wave ./wave
COPY --from=application-builder /src/web/dist ./web/dist
COPY --from=application-builder /src/config ./config
COPY --from=application-builder /src/schemas ./schemas

RUN mkdir -p /app/data \
    && chown -R wave:wave /app

ENV WAVE_PLATFORM_ADDRESS=0.0.0.0:8080
ENV WAVE_PLATFORM_DATA_PATH=/app/data
ENV WAVE_PLATFORM_CONFIG=/app/config/production.xml
ENV WAVE_PLATFORM_WEB_PATH=/app/web/dist
ENV WAVE_PLATFORM_WAVE_MODULE_PATH=/app/wave

USER wave

EXPOSE 8080

VOLUME ["/app/data"]

CMD ["./wave-platform"]
