#!/bin/bash
# zig cc wrapper for darwin targets.
#
# The Go toolchain hands the external linker "-framework CoreFoundation"
# twice: crypto/x509 declares the framework through cgo ldflags and the
# framework's dynamic imports contribute it again. zig's Mach-O linker emits
# one dylib load command per -framework argument, so the binary ends up
# listing CoreFoundation twice. dyld refuses to load such a binary with
# "duplicate linked dylib ... " and exits with code 134, which is why this
# wrapper drops repeated -framework pairs before the linker sees them.
#
# Usage: zig-cc-darwin.sh <zig-target> [compiler arguments...]
set -e

target=${1:?usage: zig-cc-darwin.sh <zig-target> [compiler arguments...]}
shift

args=()
seen=" "
while [ $# -gt 0 ]; do
	case $1 in
	-framework)
		name=${2-}
		case $seen in
		*" $name "*)
			# Already linked: skip this -framework pair.
			shift 2
			continue
			;;
		esac
		seen="$seen$name "
		args+=("-framework" "$name")
		shift 2
		;;
	*)
		args+=("$1")
		shift
		;;
	esac
done

exec zig cc -target "$target" "${args[@]}"
