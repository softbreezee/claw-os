.PHONY: build build-web build-go embed-web dev-build clean release-local install test dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# build-web: full clean install + production build of the dashboard,
# then sync the static export into internal/setup/web where Go's
# //go:embed picks it up. Use this for release builds.
build-web:
	cd web && pnpm install --frozen-lockfile && pnpm build
	rm -rf internal/setup/web
	cp -r web/out internal/setup/web

# embed-web: just sync an already-built web/out into internal/setup/web.
# Use this in the dev loop when you've run `pnpm build` yourself and
# only want to refresh what the Go binary embeds (skips the slow
# `pnpm install --frozen-lockfile`).
embed-web:
	@if [ ! -d web/out ]; then \
		echo "web/out missing — run 'cd web && pnpm build' first"; exit 1; \
	fi
	rm -rf internal/setup/web
	cp -r web/out internal/setup/web

# dev-build: fast local rebuild loop. Assumes you've already run
# `pnpm build` after editing web/. One command syncs the embed dir
# AND rebuilds the binary so a daemon restart picks both up.
# Avoids the historical footgun where running `go build` alone left
# the binary serving a stale internal/setup/web.
dev-build: embed-web
	go build -o bin/pawnix ./cmd/pawnix

build: build-web
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/pawnix ./cmd/pawnix

install: build
	cp bin/pawnix /usr/local/bin/pawnix

test:
	go test ./...

dev: build-web
	air

clean:
	rm -rf bin/ dist/ tmp/

# Build all platforms
release-local: build-web
	@mkdir -p dist
	@# macOS
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_darwin_arm64/pawnix  ./cmd/pawnix
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_darwin_amd64/pawnix  ./cmd/pawnix
	@# Linux
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_linux_arm64/pawnix   ./cmd/pawnix
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_linux_amd64/pawnix   ./cmd/pawnix
	@# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_windows_amd64/pawnix.exe ./cmd/pawnix
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/pawnix_windows_arm64/pawnix.exe ./cmd/pawnix
	@# Package: tar.gz for unix, zip for windows
	@cd dist && for d in pawnix_darwin_* pawnix_linux_*; do tar -czf "$${d}.tar.gz" -C "$$d" pawnix; done
	@cd dist && for d in pawnix_windows_*; do (cd "$$d" && zip -q "../$${d}.zip" pawnix.exe); done
	@echo "Release artifacts:"
	@ls -lh dist/*.tar.gz dist/*.zip 2>/dev/null
