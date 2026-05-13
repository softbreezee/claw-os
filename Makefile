.PHONY: build build-web clean release-local install test dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build-web:
	cd web && pnpm install --frozen-lockfile && pnpm build
	rm -rf internal/setup/web
	cp -r web/out internal/setup/web

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
