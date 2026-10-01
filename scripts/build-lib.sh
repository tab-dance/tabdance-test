#!/bin/sh
# Build libghostty-vt for a goreleaser build target (format: <os>_<arch>).
# Installs the library, headers, and pkg-config file into
# dist/lib/<target> so parallel goreleaser builds don't clobber
# each other's artifacts.
set -eu

target=${1:?usage: build-lib.sh <os>_<arch>[_<variant>]}
arch=${target#*_}
arch=${arch%%_*}

case "$arch" in
	amd64) zig_target=x86_64-linux-gnu ;;
	arm64) zig_target=aarch64-linux-gnu ;;
	*) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

ZIG=${ZIG:-zig}
prefix="$(pwd)/dist/lib/$target"

cd deps/ghostty
"$ZIG" build -Demit-lib-vt -Doptimize=ReleaseFast -Dtarget="$zig_target" --prefix "$prefix"
