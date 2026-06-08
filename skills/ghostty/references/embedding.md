# Embedding — C API and Application Runtime Architecture

In-depth reference for Ghostty's embedding API (`ghostty.h`), the application runtime abstraction (apprt), and the inspector system.

## ghostty.h — Internal C API

The `include/ghostty.h` header defines the C API consumed primarily by the macOS SwiftUI app. **This is an internal-only API** — a clean-slate C API for wider usage is being designed.

### Core Opaque Types

| C Type | Zig Implementation | Purpose |
|--------|-------------------|---------|
| `ghostty_app_t` | `*embedded.App` | Application instance |
| `ghostty_config_t` | `*configpkg.Config` | Configuration state |
| `ghostty_surface_t` | `*embedded.Surface` | Terminal surface instance |
| `ghostty_inspector_t` | `*CoreInspector` | Terminal inspector handle |

### Application Functions

| Function | Purpose |
|----------|---------|
| `ghostty_app_new(runtime_config, config)` | Create app with callback struct |
| `ghostty_app_tick(app)` | Process event loop |
| `ghostty_app_free(app)` | Free app |

### Surface Functions

| Function | Purpose |
|----------|---------|
| `ghostty_surface_new(app, config)` | Create terminal surface |
| `ghostty_surface_key(surface, key_event)` | Forward keyboard input |
| `ghostty_surface_mouse_button(...)` | Forward mouse button events |
| `ghostty_surface_mouse_motion(...)` | Forward mouse motion |
| `ghostty_surface_mouse_scroll(...)` | Forward scroll events |
| `ghostty_surface_read_selection(surface, text)` | Read selected text |
| `ghostty_surface_read_text(surface, selection, text)` | Read text from arbitrary selection |
| `ghostty_surface_has_selection(surface)` | Check for selection |
| `ghostty_surface_size(surface)` | Get surface dimensions |

### Runtime Callbacks (Host Must Implement)

| Callback | Purpose |
|----------|---------|
| `wakeup_cb` | Notify host to tick event loop |
| `action_cb` | Handle action dispatch |
| `read_clipboard_cb` | Request clipboard content |
| `confirm_read_clipboard_cb` | Confirm unsafe paste |
| `write_clipboard_cb` | Write to clipboard |
| `close_surface_cb` | Close a surface |

### Surface Configuration

```c
typedef struct {
    void* platform_view;       // Platform-specific view handle
    double scale_factor;       // Content scale factor
    uint16_t font_size;        // Initial font size
    const char* working_directory;
    const char* command;
    const char* const* env;
    const char* initial_input;
    ghostty_surface_context_s context;  // Window/tab/split context
} ghostty_surface_config_s;
```

### Platform Support

| Platform | View Type | Rendering |
|----------|-----------|------------|
| macOS | `nsview` (Objective-C) | Metal |
| iOS | `uiview` | Metal |
| Linux/GTK | GTK4 widget | OpenGL/Vulkan |

## libghostty-vt — Standalone VT Library

A **zero-dependency** (not even libc) library for VT parsing and terminal state. This is the first publicly available component.

### Header Structure (`include/ghostty/vt.h`)

```
ghostty/vt/types.h        - Base types and result codes
ghostty/vt/allocator.h    - Custom memory allocators
ghostty/vt/build_info.h   - Compile-time build configuration
ghostty/vt/color.h        - Color types
ghostty/vt/device.h       - Device attributes
ghostty/vt/focus.h        - Focus encoding
ghostty/vt/formatter.h    - Terminal content formatting
ghostty/vt/render.h       - High-performance render state
ghostty/vt/terminal.h     - Complete terminal state machine
ghostty/vt/grid_ref.h     - Grid reference (untracked)
ghostty/vt/grid_ref_tracked.h - Tracked grid reference
ghostty/vt/osc.h          - OSC parser
ghostty/vt/sgr.h          - SGR parser
ghostty/vt/style.h        - Style types
ghostty/vt/sys.h          - System interface (embedder-provided)
ghostty/vt/key/encoder.h   - Key encoding
ghostty/vt/key/event.h    - Key event types
ghostty/vt/kitty_graphics.h - Kitty graphics protocol
ghostty/vt/modes.h        - Terminal modes
ghostty/vt/mouse/encoder.h - Mouse encoding
ghostty/vt/mouse/event.h  - Mouse event types
ghostty/vt/paste.h        - Paste safety validation
ghostty/vt/point.h        - Point/coordinate types
ghostty/vt/screen.h       - Cell and row types
ghostty/vt/selection.h    - Selection types
ghostty/vt/size_report.h  - Size report types
ghostty/vt/wasm.h         - WebAssembly utilities
```

### Status

- **Zig API**: Available for testing
- **C API**: In development
- **WASM target**: Planned
- **Future `libghostty-<x>` libraries**: Input handling, GPU rendering, GTK widgets, Swift frameworks
- **Versioning**: Will be versioned separately from Ghostty the application
- **Alpha quality** (API may change; core logic is production-proven)

## apprt — Application Runtime Abstraction

Ghostty uses an **application runtime abstraction** (`src/apprt/`) to support multiple GUI frontends:

| Runtime | File | Platform |
|---------|------|----------|
| `embedded` | `src/apprt/embedded/` | macOS (SwiftUI/AppKit) |
| `gtk` | `src/apprt/gtk/` | Linux (GTK4) |
| `none` | `src/apprt/none.zig` | Headless (no UI) |

### none Runtime (Headless)

The `none` runtime provides a minimal headless stub:
- `App` struct: Only implements `performIpc()` which always returns `false`
- `Surface` struct: Empty — no surface functionality

Used for builds that need the core library without any GUI.

## Inspector

The Inspector is a built-in debugging tool using **Dear ImGui** (via `dcimgui` C wrapper).

### C API

| Function | Purpose |
|----------|---------|
| `ghostty_surface_inspector(surface)` | Get/create inspector |
| `ghostty_inspector_free(surface)` | Free inspector |
| `ghostty_inspector_set_focus(inspector, bool)` | Set focus |
| `ghostty_inspector_set_content_scale(inspector, double, double)` | Set scale |
| `ghostty_inspector_set_size(inspector, uint32, uint32)` | Set dimensions |
| `ghostty_inspector_key(inspector, key_event)` | Forward key events |
| `ghostty_inspector_text(inspector, text)` | Forward text input |
| `ghostty_inspector_mouse_button/motion/scroll(...)` | Forward mouse |

### Inspector Widgets

| Widget | File | Purpose |
|--------|------|---------|
| Page | `page.zig` | Page inspection |
| PageList | `pagelist.zig` | Page list |
| Screen | `screen.zig` | Screen state |
| Style | `style.zig` | Style information |
| Surface | `surface.zig` | Surface properties |
| Terminal | `terminal.zig` | Terminal state |
| Termio | `termio.zig` | I/O state |

### Inspector Actions

| Action | Description |
|--------|-------------|
| `GHOSTTY_INSPECTOR_TOGGLE` | Toggle inspector |
| `GHOSTTY_INSPECTOR_SHOW` | Show inspector |
| `GHOSTTY_INSPECTOR_HIDE` | Hide inspector |

### Metal Rendering (macOS)

| Function | Purpose |
|----------|---------|
| `ghostty_inspector_metal_init(inspector, device)` | Initialize Metal rendering |
| `ghostty_inspector_metal_render(inspector, command_buffer, drawable)` | Render with Metal |
| `ghostty_inspector_metal_shutdown(inspector)` | Shut down Metal rendering |

## Surface Architecture

A **Surface** (`src/Surface.zig`) represents a single terminal instance:

| Component | Purpose |
|-----------|---------|
| PTY session | Child process I/O |
| Renderer thread | GPU command submission, frame timing |
| I/O thread | PTY reads/writes, VT emulation |
| DerivedConfig | Thread-safe per-surface config snapshot |
| Mailbox | Thread-safe message queue |
| Keyboard state | Pressed keys, key table stack (max depth 8) |

### Threading Model

- **Main thread**: Event loop, input processing, keybinding matching
- **I/O thread**: PTY read, VT parsing, terminal state updates
- **Renderer thread**: GPU rendering, frame timing

Thread-safe communication via `Mailbox.Queue` (blocking queue).
