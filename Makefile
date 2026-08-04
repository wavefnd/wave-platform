GOCACHE ?= /tmp/wave-platform-gocache

.PHONY: frontend-install frontend-dev frontend-build test build run clean

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

test:
	GOCACHE=$(GOCACHE) go test ./...
	cd frontend && npm run typecheck

build: frontend-build
	mkdir -p bin
	GOCACHE=$(GOCACHE) go build -trimpath -o bin/wave-platform ./cmd/server

run: frontend-build
	GOCACHE=$(GOCACHE) go run ./cmd/server

clean:
	$(RM) -r frontend/dist bin
