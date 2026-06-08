# CGo & Build — Linking, Concurrency, and Cross-Compilation

In-depth reference for the cgo integration, linking model, concurrency rules, memory management, and cross-compilation in go-libghostty.

## Linking Model

### Static Linking (Default)

go-libghostty links libghostty statically via pkg-config. The `cgo_static.go` file (build tag `!dynamic`) configures:

```go
// #pkg-config: --static ghostty-vt
```

### Dynamic Linking

Use `-tags dynamic` to link dynamically:

```go
// cgo_dynamic.go (build tag: dynamic)
// #pkg-config: ghostty-vt
```

## Memory Allocation

go-libghostty uses libghostty's own allocator (`Alloc`/`Free`) rather than `C.malloc`/`C.free`:

```go
func Alloc(size uintptr) unsafe.Pointer  // wraps ghostty_alloc
func Free(ptr unsafe.Pointer, size uintptr)  // wraps ghostty_free
```

**Critical on Windows**: Zig's libc and MSVC's CRT maintain separate heaps. Using `C.malloc`/`C.free` would cause heap corruption.

## cgo.Handle and Trampolines

Effect callbacks use C trampolines:

1. `Terminal.handle` stores a `cgo.Handle` pointing back to the `*Terminal`
2. The handle is set as C-side userdata (`GHOSTTY_TERMINAL_OPT_USERDATA`)
3. C callback functions (exported via `//export`) receive the userdata
4. They recover the `cgo.Handle`, look up the `*Terminal`, and dispatch to the Go effect handler

```go
//go:nocheckptr
func handleToPointer(h cgo.Handle) unsafe.Pointer {
    return unsafe.Pointer(h)
}
```

The `//go:nocheckptr` directive suppresses checkptr because `cgo.Handle` is an opaque integer (uintptr), not a real Go pointer.

## Concurrency Rules

### Terminal

- **Not safe for concurrent use**
- Serialize all calls: getters, setters, `VTWrite`, `Resize`, `Close`, and any borrowed handles
- Effect callbacks run **synchronously during VTWrite** — must not call `VTWrite` on the same terminal

### Borrowed Views

Borrowed views are only valid until the next mutating terminal call:

| Type | Borrowed From | Valid Until |
|------|---------------|-------------|
| `GridRef` | `term.GridRef()` | Next VTWrite/Reset/Resize |
| `KittyGraphics` | `term.KittyGraphics()` | Next VTWrite/Reset/Resize |
| `Selection` (from `term.Selection()`) | Terminal | Next VTWrite/Reset/Resize |
| Raw pixel slices | Various | Next VTWrite/Reset/Resize |

### Copied Values

Copied values are regular Go values and may be retained freely:

| Type | Source |
|------|--------|
| `Cell` | `GridRef.Cell()` |
| `Row` | `GridRef.Row()` |
| `Style` | `GridRef.Style()` |
| `ColorRGB` | Color getters |
| `Palette` | Palette getters |

### RenderState

- Hold **exclusive access** to the terminal during `Update()`
- After `Update()` returns, the render state can be read **without** the terminal
- Do **not** call `Update()` concurrently with reads from the same render state
- Iterators are valid only until the next `Update()`

### KeyEncoder / MouseEncoder

- Not safe for concurrent use
- `SetOptFromTerminal` requires serialized access to both encoder and terminal

## Cross-Compilation

go-libghostty supports cross-compilation via **Zig as a C compiler**:

```sh
# Cross-compile for Linux aarch64
GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" go build

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 CC="zig cc -target x86_64-macos" go build

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows-gnu" go build
```

Zig provides a drop-in C compiler that handles cross-compilation seamlessly, including linking against the correct libc for the target platform.

## Build Info

```go
func GetBuildInfo() (BuildInfo, error)
```

Returns compile-time configuration:

| Field | Type | Description |
|-------|------|-------------|
| `SIMD` | `bool` | Whether SIMD-optimized parsing is enabled |
| `KittyGraphics` | `bool` | Whether Kitty graphics support is compiled in |
| `TmuxControlMode` | `bool` | Whether tmux control mode parsing is enabled |
| `Optimize` | `string` | Optimization mode |
| `VersionString` | `string` | Full version string |
| `VersionMajor/Minor/Patch` | `uint32` | Semantic version components |
| `VersionPre` | `string` | Pre-release identifier |
| `VersionBuild` | `string` | Build metadata |

## Error Handling

All C API calls go through `resultError()`:

```go
func resultError(result C.GhosttyResult) error {
    switch Result(result) {
    case ResultSuccess:
        return nil
    default:
        return &Error{Result: Result(result)}
    }
}
```

| Result | Description |
|--------|-------------|
| `ResultSuccess` | Operation succeeded |
| `ResultOutOfMemory` | Memory allocation failed |
| `ResultInvalidValue` | Invalid argument |
| `ResultOutOfSpace` | Buffer too small (check output size) |
| `ResultNoValue` | No value available |

## WASM Target

libghostty-vt includes a `wasm.h` header with WebAssembly-specific allocation helpers. These are **not available in native cgo builds** — they are only for WASM compilation targets.

## Missing / TODO APIs

From the project's `TODO.md`:

**Not yet bound in Go**:
- OSC parser (`osc.h`)
- SGR parser (`sgr.h`)

**WASM-only (not in native cgo build)**:
- WebAssembly allocation helpers (`wasm.h`)

**Partially bound**:
- `ghostty_mode_report_encode()`
- `ghostty_size_report_encode()`
- `ghostty_type_json()`
- `ghostty_style_default()`
- `ghostty_color_rgb_get()`
- `ghostty_formatter_format_buf()`
