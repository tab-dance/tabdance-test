GHOSTTY_DIR ?= /home/rsteube/Documents/development/github/ghostty-org/ghostty
PKG_CONFIG_PATH := $(GHOSTTY_DIR)/zig-out/share/pkgconfig
LD_LIBRARY_PATH := $(GHOSTTY_DIR)/zig-out/lib
BUILD_FLAGS :=

.PHONY: build build-dynamic build-static clean test

build: build-static

build-dynamic:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build -tags dynamic $(BUILD_FLAGS) -o tabdance-test ./cmd/tabdance-test/

build-static:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build $(BUILD_FLAGS) -o tabdance-test ./cmd/tabdance-test/

clean:
	rm -f tabdance-test

test:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test ./...

run: build-dynamic
	LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) ./tabdance-test
