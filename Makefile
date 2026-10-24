GHOSTTY_DIR := deps/ghostty
ZIG ?= zig
LIB := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/lib/libghostty-vt.a
PKG_CONFIG_PATH := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/share/pkgconfig
LD_LIBRARY_PATH := $(CURDIR)/$(GHOSTTY_DIR)/zig-out/lib
BUILD_FLAGS :=

.PHONY: build build-dynamic build-static build-lib clean test test-race fmt vet run

build: build-static

build-lib:
	@test -f $(LIB) || (cd $(GHOSTTY_DIR) && $(ZIG) build -Demit-lib-vt -Doptimize=ReleaseFast)

build-dynamic: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) \
		go build -tags dynamic $(BUILD_FLAGS) -o tt ./cmd/tt/

build-static: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build $(BUILD_FLAGS) -o tt ./cmd/tt/

clean:
	rm -f tt

test: build-static
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test ./...

test-race: build-static
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test -race ./...

fmt:
	gofmt -l -w cmd pkg internal

vet:
	go vet ./cmd/... ./pkg/... ./internal/...

run: build-dynamic
	LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) ./tt
