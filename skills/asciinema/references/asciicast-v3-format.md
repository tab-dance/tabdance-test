# Asciicast v3 Format Specification

In-depth reference for the asciicast v3 file format used by asciinema CLI 3.x+.

## Overview

Asciicast v3 is based on **newline-delimited JSON** (JSON Lines). The file consists of a **header line** (JSON object; never a comment) followed by an **event stream** (one JSON array per line). Lines starting with `#` are comments and are ignored by parsers. The file extension is `.cast`, MIME type `application/x-asciicast`.

### Example

```
{"version": 3, "term": {"cols": 80, "rows": 24, "type": "xterm-ghostty"}, "timestamp": 1504467315, "title": "Demo", "env": {"SHELL": "/bin/zsh"}}
# comments may appear after the header
[0.248848, "o", "\u001b[1;31mHello \u001b[32mWorld!\u001b[0m\n"]
[0.752528, "o", "That was ok\rThis is better."]
[0.5, "m", "chapter two"]
[0.3, "r", "80x24"]
[5.0, "x", "0"]
```

## Header (First Line)

### Required Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `version` | Integer | Format version. **Must be `3`**. |
| `term` | Object | Terminal description: `cols`, `rows` required; `type` (TERM), `version` (XTVERSION), and `theme` optional. |

**Minimal valid header:**
```json
{"version": 3, "term": {"cols": 80, "rows": 24}}
```

### Optional Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `timestamp` | Integer | Unix timestamp of recording start. |
| `idle_time_limit` | Float | Max idle time between events. Players reduce inactivity delays to at most this value. `0` means exactly that — every gap collapses to zero and playback finishes instantly — so omit the attribute entirely for unbounded playback. |
| `command` | String | The command that was recorded. |
| `title` | String | Title of the recording. |
| `env` | Object (String→String) | Captured environment (v3 default captures `SHELL`; `TERM` lives in `term.type`). Values must be strings — `null` env values fail v3 parsing (v2 tolerated them). Unknown attributes are ignored, so producers may add extra keys freely. |

## Event Stream

Each event is `[interval, code, data]` where **interval** is the delay in seconds since the **previous event** (or the start of the recording for the first event), accumulated into an absolute time on read. Interval-based timing makes inserting/deleting events easy. The writer in asciinema 3.x quantizes to milliseconds with error diffusion; readers keep microsecond precision.

### Event Types

| Code | Description |
|------|-------------|
| `o` | Output printed by the process (string). An empty data string used to be standard; producers such as tt append a final no-op `[interval, "o", ""]` event so trailing quiet time survives playback. |
| `i` | Input written to the process (string). |
| `m` | Marker/breakpoint for navigation and automation; `data` is a string label (can be empty). |
| `r` | Resize; `data` is `"COLSxROWS"` (e.g. `"90x30"`). |
| `x` | Exit status; `data` is the status as a string (e.g. `"0"`). |

### Comment Lines

Lines starting with `#` (after the header) and blank lines are skipped by parsers. **The first line of the file must always be the header, never a comment.**

## Parser Behavior (verified against asciinema 3.x source)

- **Format detection uses the first line only**: the v3 parser is tried first, then v2, then v1 (whole file); any parse error falls through silently and later event errors never affect detection. Files starting with the zstd magic bytes are decompressed transparently.
- **Only `version == 3` discriminates v3 from v2** — there is no `deny_unknown_fields`, so extra header or `term` keys (even `width`/`height` leftovers) are tolerated and ignored.
- **`term.theme` is strict**: when present, `fg`/`bg` must be exactly 7-character hex strings and `palette` must yield exactly 8 or 16 hex triplets. A malformed theme makes the v3 load fail, and the file then fails the other loaders too (generic “not a v1, v2, v3 asciicast file” error).
- **Event time parsing truncates**: the interval's fractional part is truncated to 6 digits (microseconds, no rounding), then added to the running absolute time (`prev + interval`).
- **Duration** of a cast equals the last event's cumulative time — which is why a trailing no-op output event extends playback.

## Compatibility

| Format version reads/writes |
|------------------------------|
| asciinema CLI 3.x records v3 by default and **reads v1, v2 and v3** (playback + `asciinema convert`). To keep producing v2 from a 3.x CLI use `--output-format asciicast-v2`. |

v3 is **not backward compatible** with v2 (different header schema — `term` object vs top-level `width`/`height` — and interval-based event times vs absolute times). Older players (asciinema player < 3.10.0) and older servers cannot read v3.

## Adoption Timeline

- asciinema CLI **v3.0** (September 2025)
- asciinema player **v3.10.0**
- asciinema server **20250509**
