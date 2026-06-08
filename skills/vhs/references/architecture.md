# Architecture — How VHS Works Internally

In-depth reference for VHS's internal architecture: the three-process pipeline, frame recording, and rendering.

## Three-Process Pipeline

VHS uses three separate processes:

### 1. ttyd — Terminal Multiplexer

VHS spawns `ttyd` on a random port with specific xterm.js options:

```sh
ttyd --port=<random> --interface=127.0.0.1 \
  -t rendererType=canvas \
  -t disableResizeOverlay=true \
  -t enableSixel=true \
  -t customGlyphs=true \
  --once --writable <shell_command>
```

ttyd hosts a real PTY and renders it via xterm.js in a browser.

### 2. go-rod (Chromium)

A headless Chrome browser connects to the ttyd web page:

- Opens a `rod.Page` pointing at `http://localhost:<port>`
- Sets viewport size via `Page.MustSetViewport(width, height, 0, false)`
- Locates two xterm.js `<canvas>` elements:
  - `canvas.xterm-text-layer` — Text rendering
  - `canvas.xterm-cursor-layer` — Cursor rendering
- Applies terminal options (font, theme, etc.) via `Page.Eval()` calling xterm.js APIs

### 3. ffmpeg

After recording, ffmpeg composites the text and cursor frame sequences into the final output format.

## Frame Recording

The `Record()` method runs a goroutine that, at the configured framerate interval:

1. Calls `CursorCanvas.CanvasToImage("image/png", 1.0)` and `TextCanvas.CanvasToImage("image/png", 1.0)`
2. Writes each frame as two separate PNG files: `frame-text-NNNNN.png` and `frame-cursor-NNNNN.png`
3. If screenshot capture is queued, also triggers screenshot generation

Recording can be **paused** (`PauseRecording()`) and **resumed** (`ResumeRecording()`) — this is how `Hide`/`Show` commands work.

## Key Input System

VHS uses go-rod's keyboard input system (`github.com/go-rod/rod/lib/input`):

### Keymap

Every printable ASCII character is mapped to a `go-rod input.Key`:

```go
var keymap = map[rune]input.Key{
    'a': input.KeyA,  'A': shift(input.KeyA),
    '1': input.Digit1, '!': shift(input.Digit1),
    ' ': input.Space, '\n': input.Enter,
    '\t': input.Tab,  '\b': input.Backspace,
    '←': input.ArrowLeft, '↑': input.ArrowUp,
    '→': input.ArrowRight, '↓': input.ArrowDown,
}
```

### Type Execution

1. Iterate over each rune in the string
2. Look up in keymap → `v.Page.Keyboard.Type(k)`
3. If not found (Unicode/emoji) → `v.Page.MustElement("textarea").Input(string(r))`
4. Sleep for `typingSpeed` between characters

### Ctrl Combinations

Uses go-rod's `KeyActions` API to hold `ControlLeft` and press additional keys:

```go
action := v.Page.KeyActions().Press(input.ControlLeft)
action.Type(input.KeyR)
action.Do()
```

### Scroll

Calls xterm.js's scroll API directly:

```go
v.Page.Eval(fmt.Sprintf("() => term.scrollLines(%d)", direction))
```

## Terminal Resizing

### Browser Viewport (Pixel Dimensions)

During `Setup()`, VHS calculates the effective terminal area:

```go
width := Style.Width - 2*padding - 2*margin
height := Style.Height - 2*padding - 2*margin - bar
vhs.Page.MustSetViewport(width, height, 0, false)
```

### xterm.js Terminal Fit

After setting viewport and font options:

```go
vhs.Page.MustEval("term.fit")
```

xterm.js's `fit` addon calculates columns/rows from pixel dimensions, font size, line height, and letter spacing.

## Rendering Pipeline

After recording, `Render()` calls:

| Output | Function | Encoding |
|--------|----------|----------|
| `.gif` | `MakeGIF(opts)` | ffmpeg with palettegen for optimized GIF |
| `.mp4` | `MakeMP4(opts)` | ffmpeg H.264 |
| `.webm` | `MakeWebM(opts)` | ffmpeg VP9 |
| `.png` | `MakeScreenshots(opts)` | ffmpeg compositing |

### Compositing Layers

ffmpeg composites:
1. Text frames (stream 0) + Cursor frames (stream 1)
2. Optional margin fill overlay
3. Optional window bar overlay (drawn in Go using `image/draw`)
4. Optional border radius mask (drawn in Go using antialiased masks)
5. Loop offset (renames frame files to shift GIF start point)

## Screenshot Command

1. `v.ScreenshotNextFrame(path)` sets `frameCapture = true` and stores the target path
2. On the next frame capture cycle, the recording goroutine calls `makeScreenshot(frameNum)`
3. After recording, `MakeScreenshots()` generates ffmpeg commands that composite the text and cursor frames for the recorded frame numbers into the final PNG screenshots

## VHS Struct (Internal)

```go
type VHS struct {
    Options      *Options
    Errors       []error
    Page         *rod.Page       // go-rod browser page
    browser      *rod.Browser
    TextCanvas   *rod.Element    // xterm.js text canvas
    CursorCanvas *rod.Element    // xterm.js cursor canvas
    mutex        *sync.Mutex
    started      bool
    recording    bool
    tty          *exec.Cmd       // ttyd process
    totalFrames  int
    close        func() error
}
```

**VHS is NOT a Go library** — all types are in the `main` package. Only the lexer/parser/token sub-packages are importable.
