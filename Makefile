.PHONY: all build generate tidy clean test dist

VERSION ?= dev
COMMIT ?= unknown
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
DIST_DIR ?= dist
PROJECT_NAME ?= shortcut-cli
LDFLAGS = -X github.com/nazar256/shortcut-cli/internal/cli.Version=$(VERSION) -X github.com/nazar256/shortcut-cli/internal/cli.Commit=$(COMMIT) -X github.com/nazar256/shortcut-cli/internal/cli.BuildDate=$(BUILD_DATE)

all: tidy generate build

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/shortcut ./cmd/shortcut

test:
	go test ./...

dist:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		os=$${target%/*}; \
		arch=$${target#*/}; \
		name=$(PROJECT_NAME)_$(VERSION)_$${os}_$${arch}; \
		tmpdir=$$(mktemp -d); \
		CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} go build -trimpath -ldflags "$(LDFLAGS)" -o $${tmpdir}/shortcut ./cmd/shortcut; \
		tar -C $${tmpdir} -czf $(DIST_DIR)/$${name}.tar.gz shortcut; \
		rm -rf $${tmpdir}; \
	done
	cd $(DIST_DIR) && if command -v sha256sum >/dev/null 2>&1; then sha256sum *.tar.gz; else shasum -a 256 *.tar.gz; fi | LC_ALL=C sort > $(PROJECT_NAME)_$(VERSION)_checksums.txt

generate:
	go generate ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/
	rm -rf dist/
	rm -f internal/gen/shortcutv3/shortcut.gen.go
