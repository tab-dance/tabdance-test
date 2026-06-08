# Asciicast v2 Format Specification

In-depth reference for the asciicast v2 file format used by asciinema CLI 2.x+.

## Overview

Asciicast v2 is based on **newline-delimited JSON** (JSON Lines). The file consists of a **header line** (JSON object) followed by an **event stream** (one JSON array per line).

### Example

```
{"version": 2, "width": 80, "height": 24, "timestamp": 1504467315, "title": "Demo", "env": {"TERM": "xterm-256color", "SHELL": "/bin/zsh"}}
[0.248848, "o", "\u001b[1;31mHello \u001b[32mWorld!\u001b[0m\n"]
[1.001376, "o", "That was ok\rThis is better."]
[1.500000, "m", ""]
[2.143733, "o", "Now... "]
[4.050000, "r", "80x24"]
[6.541828, "o", "Bye!"]
```

## Header (First Line)

### Required Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `version` | Integer | Format version. **Must be `2`**. |
| `width` | Integer | Terminal width in columns. |
| `height` | Integer | Terminal height in rows. |

**Minimal valid header:**
```json
{"version": 2, "width": 80, "height": 24}
```

### Optional Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `timestamp` | Integer | Unix timestamp of recording start. |
| `duration` | Float | Total duration in seconds (when known upfront). |
| `idle_time_limit` | Float | Max idle time between frames. Players reduce inactivity delays to at most this value. |
| `command` | String | The command that was recorded. |
| `title` | String | Title of the recording. |
| `env` | Object (String→String) | Captured environment variables. Default: `SHELL` and `TERM` only. Implementations should **not** capture additional env vars unless explicitly requested. |
| `theme` | Object | Color theme of the recorded terminal. |

### Theme Sub-Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `theme.fg` | String | Normal text color in `#rrggbb` CSS format. |
| `theme.bg` | String | Normal background color in `#rrggbb`. |
| `theme.palette` | String | 8 or 16 colors separated by colons, each in `#rrggbb`. |

**Complete header example:**
```json
{
  "version": 2,
  "width": 80,
  "height": 24,
  "timestamp": 1504467315,
  "duration": 123.45,
  "command": "/bin/bash",
  "title": "Demo",
  "idle_time_limit": 2,
  "env": {
    "SHELL": "/bin/bash",
    "TERM": "xterm-256color"
  },
  "theme": {
    "fg": "#d0d0d0",
    "bg": "#212121",
    "palette": "#151515:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#d0d0d0:#505050:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#f5f5f5"
  }
}
```

## Event Stream

Each subsequent line is a **3-element JSON array**:

```
[time, code, data]
```

| Element | Type | Description |
|---------|------|-------------|
| `time` | Float | Seconds since start of recording. |
| `code` | String | Event type: `"o"`, `"i"`, `"m"`, or `"r"`. |
| `data` | Any | Event-specific data. |

### Event Codes

The list is **open to extension**. Tools should **ignore** unknown event codes.

#### `"o"` — Output

Data written to the terminal. `data` is a **string** with valid UTF-8, non-printable codepoints as `\uXXXX`.

```json
[5.0, "o", "hello"]
[0.1, "o", "\u001b[1;31mred text\u001b[0m"]
```

#### `"i"` — Input

Keystrokes/raw data sent from terminal to stdin. `data` is a **string**.

```json
[5.0, "i", "h"]
[5.1, "i", "\r"]
```

> **Note:** asciinema CLI does **not** capture keyboard input by default.

#### `"m"` — Marker

Breakpoints for playback navigation and automation. `data` is a **string** (label, can be empty).

```json
[5.0, "m", ""]
[10.0, "m", "Configuration"]
```

#### `"r"` — Resize

Terminal resize event. `data` is a **string** formatted as `"{COLS}x{ROWS}"`.

```json
[5.0, "r", "100x50"]
```

## Timing

Timing is **relative** — each event's `time` is seconds elapsed since recording start. Events are not required to be in strictly increasing time order, though they typically are.

## File Properties

| Property | Value |
|----------|-------|
| Suggested extension | `.cast` |
| Suggested MIME type | `application/x-asciicast` |

## Design Benefits

The newline-delimited JSON structure enables:

- **Incremental, real-time writing** — not possible with v1's single JSON object
- **Minimal memory usage** — disk space is the only limit
- **Crash safety** — interrupted recordings are not lost
- **Real-time streaming** — can be processed line by line

## v1 vs v2

v2 is **not backward compatible** with v1 due to the structural change (v1 was a single JSON object, v2 is JSON Lines).

**Support timeline:**
- asciinema CLI **v2.0**
- asciinema player **v2.6**
- asciinema server **v20171105**
