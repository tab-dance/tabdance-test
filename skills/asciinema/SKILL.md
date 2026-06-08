---
name: asciinema
description: >
  Use when working with asciinema, asciicast recordings, or the asciicast v2 file format —
  recording terminal sessions, playing them back, parsing .cast files, generating recordings
  programmatically, or integrating asciinema into testing pipelines. Triggers on: "asciinema",
  "asciicast", "asciicast v2", ".cast", "cast file", "terminal recording", "session recording",
  "asciinema play", "asciinema rec", "asciinema cat", "idle_time_limit", "JSON Lines",
  "application/x-asciicast".
user-invocable: true
---

# Asciinema & Asciicast v2 In-Depth Reference

Comprehensive reference for [asciinema](https://asciinema.org) — the terminal session recording tool — and the asciicast v2 file format (`.cast`). Covers the format specification, CLI usage, programmatic generation, and integration testing patterns.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference |
||----------|----------|
|| asciicast v2, .cast, format, header, event stream, output, input, marker, resize, timing, theme, JSON Lines, version 2, file format specification | [references/asciicast-v2-format.md](references/asciicast-v2-format.md) |
|| asciinema CLI, asciinema rec, asciinema play, asciinema cat, asciinema upload, recording, playback, authentication, install, configuration | [references/cli-usage.md](references/cli-usage.md) |
|| programmatic generation, Go, write cast file, create recording, embed recording, testing, CI, asciinema player, embed | [references/programmatic.md](references/programmatic.md) |

## Quick Guide

- **What is the asciicast v2 file format?** → [references/asciicast-v2-format.md](references/asciicast-v2-format.md)
- **How do I record a terminal session?** → [references/cli-usage.md](references/cli-usage.md)
- **How do I play back a recording?** → [references/cli-usage.md](references/cli-usage.md)
- **How do I generate .cast files programmatically?** → [references/programmatic.md](references/programmatic.md)
- **How do I embed recordings in a web page?** → [references/programmatic.md](references/programmatic.md)
- **How do I use asciinema for integration testing?** → [references/programmatic.md](references/programmatic.md)
- **What are the event codes (o/i/m/r)?** → [references/asciicast-v2-format.md](references/asciicast-v2-format.md)
- **How does timing work in asciicast files?** → [references/asciicast-v2-format.md](references/asciicast-v2-format.md)

## Cross-Project References

- For Ghostty/libghostty headless terminal testing, use the **ghostty** skill.
- For VHS tape-based testing, use the **vhs** skill.
- For the tabdance-test MCP server (headless terminal testing), use the **tabdance-test** skill.
