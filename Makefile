GHOSTTY_DIR := deps/ghostty
ZIG ?= zig
PKG_CONFIG_PATH := $(GHOSTTY_DIR)/zig-out/share/pkgconfig
LD_LIBRARY_PATH := $(GHOSTTY_DIR)/zig-out/lib
BUILD_FLAGS :=

.PHONY: build build-dynamic build-static build-lib clean test run

build: build-static

build-lib:
	cd $(GHOSTTY_DIR) && $(ZIG) build -Demit-lib-vt -Doptimize=ReleaseFast

build-dynamic: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build -tags dynamic $(BUILD_FLAGS) -o tabdance-test ./cmd/tabdance-test/

build-static: build-lib
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build $(BUILD_FLAGS) -o tabdance-test ./cmd/tabdance-test/

clean:
	rm -f tabdance-test

test: build-static
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test ./...

run: build-dynamic
	LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) ./tabdance-test
