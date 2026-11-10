GHOSTTY_DIR := deps/ghostty
ZIG ?= zig
LIB := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/lib/libghostty-vt.a
PKG_CONFIG_PATH := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/share/pkgconfig
LD_LIBRARY_PATH := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/lib
BUILD_FLAGS :=

.PHONY: build build-dynamic build-static build-lib build-darwin-arm64 build-darwin-amd64 clean test test-race fmt vet run

build: build-static

build-lib:
	@test -f $(LIB) || (cd $(GHOSTTY_DIR) && $(ZIG) build -Demit-lib-vt -Doptimize=ReleaseFast)

build-dynamic: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) \
		go build -tags dynamic $(BUILD_FLAGS) -o tt ./cmd/tt/

build-static: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build $(BUILD_FLAGS) -o tt ./cmd/tt/

# Stub .tbd files satisfy Apple frameworks absent from a zig toolchain (no
# Apple SDK is available); dyld resolves the real frameworks on macOS at load
# time. zig-cc-darwin.sh wraps zig cc so the repeated -framework CoreFoundation
# the Go toolchain passes only produces a single dylib load command.
MACOS_STUBS := $(CURDIR)/scripts/macos-stubs
MACOS_CC := $(CURDIR)/scripts/zig-cc-darwin.sh

build-darwin-arm64: build-lib
	cd $(GHOSTTY_DIR) && $(ZIG) build -Demit-lib-vt -Doptimize=ReleaseFast -Dtarget=aarch64-macos --prefix $(CURDIR)/dist/lib/darwin_arm64
	PKG_CONFIG_PATH=$(CURDIR)/dist/lib/darwin_arm64/share/pkgconfig \
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC="$(MACOS_CC) aarch64-macos" \
	go build -ldflags "-w -extldflags '-L $(MACOS_STUBS) -F $(MACOS_STUBS)'" -o tt-darwin-arm64 ./cmd/tt/

build-darwin-amd64: build-lib
	cd $(GHOSTTY_DIR) && $(ZIG) build -Demit-lib-vt -Doptimize=ReleaseFast -Dtarget=x86_64-macos --prefix $(CURDIR)/dist/lib/darwin_amd64
	PKG_CONFIG_PATH=$(CURDIR)/dist/lib/darwin_amd64/share/pkgconfig \
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="$(MACOS_CC) x86_64-macos" \
	go build -ldflags "-w -extldflags '-L $(MACOS_STUBS) -F $(MACOS_STUBS)'" -o tt-darwin-amd64 ./cmd/tt/

clean:
	rm -f tt tt-darwin-arm64 tt-darwin-amd64

test: build-static
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test ./...

test-race: build-static
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test -race ./...

fmt:
	gofmt -l -w cmd pkg internal

vet: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go vet ./cmd/... ./pkg/... ./internal/...

run: build-dynamic
	LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) ./tt
