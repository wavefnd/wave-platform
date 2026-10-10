GOCACHE ?= /tmp/wave-platform-gocache
WAVEC ?= wavec

.PHONY: frontend-install frontend-dev frontend-build test build run clean wave-version-check

wave-version-check:
	python3 tools/check-wave-version.py "$(WAVEC)"

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

test:
	python3 tools/test-server-transfer.py
	GOCACHE=$(GOCACHE) go test ./...
	cd frontend && npm run typecheck
	cd frontend && npm run test:docs
	cd frontend && npm run test:seo
	cd frontend && npm run test:toolchains
	cd frontend && npm run test:playground

build: wave-version-check frontend-build
	mkdir -p bin
	GOCACHE=$(GOCACHE) go build -trimpath -o bin/wave-platform ./cmd/server

run: wave-version-check frontend-build
	GOCACHE=$(GOCACHE) go run ./cmd/server

clean:
	$(RM) -r frontend/dist bin
