# syntax=docker/dockerfile:1

FROM node:24-bookworm-slim AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json* ./

RUN if [ -f package-lock.json ]; then \
        npm ci; \
    else \
        npm install; \
    fi

COPY frontend/ ./

RUN npm run build


FROM golang:1.25-trixie AS application-builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        clang \
        cmake \
        curl \
        git \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

ARG WAVE_VERSION=0.2.0-pre-beta
ARG VEX_VERSION=0.0.1

COPY frontend/public/install.sh /tmp/install-wave.sh

RUN bash /tmp/install-wave.sh --version "${WAVE_VERSION}" --vex-version "${VEX_VERSION}" \
    && rm -f /tmp/install-wave.sh

ENV PATH="/root/.wave/bin:${PATH}"

RUN wavec --version
RUN vex --version

WORKDIR /src

COPY go.mod go.sum* ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY config/ ./config/
COPY schemas/ ./schemas/
COPY native/ ./native/
COPY wave/ ./wave/

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
    && cc -shared \
        -Wl,-soname,libwave-media-policy.so \
        -o build/wave/libwave-media-policy.so \
        build/wave/media-policy.o \
    && cc -shared \
        -Wl,-soname,libwave-source-analyzer.so \
        -o build/wave/libwave-source-analyzer.so \
        build/wave/source-analyzer.o \
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
    && cp wave/source-analyzer/module.xml build/wave/source-analyzer.xml

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
