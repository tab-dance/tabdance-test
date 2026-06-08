# Tape DSL — VHS Command Reference

In-depth reference for the VHS `.tape` DSL: all commands, syntax, and usage.

## Overview

VHS tape files use a custom DSL to define terminal interactions. The file goes through a Lexer → Parser → Evaluator pipeline.

### Example

```tape
Output "demo.gif"
Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 600

Type "echo 'Hello, World!'"
Enter
Sleep 1
Wait /Hello/
Screenshot "hello.png"
```

## Commands

### Output

```
Output <path>
```

Output file. Extension determines format:

| Extension | Format |
|-----------|--------|
| `.gif` | Animated GIF |
| `.mp4` | H.264 video |
| `.webm` | VP9 video |
| `.txt` | Plain text (golden file) |
| `.ascii` | Plain text (golden file) |
| directory | PNG frame sequence |

### Require

```
Require <program>
```

Fail early if program not in `$PATH`. Must appear at top of file, before any non-setting command.

### Set

```
Set <Setting> <Value>
```

Configure terminal/recording. See [settings.md](settings.md) for all options.

**Important**: All settings (except `TypingSpeed`) must appear at the top of the tape file, before any non-setting/non-output command. Settings applied after are silently ignored.

### Type

```
Type[@time] "string"
Type[@time] "string with {{.Var}}"
```

Type characters with optional per-command typing speed. Each character is sent individually with `typingSpeed` delay between them.

- `@time` overrides the global `TypingSpeed` for this command (e.g., `Type@500ms "slow"`)
- Template variables `{{.Var}}` are replaced from environment or `Env` commands
- Unicode/emoji characters not in the keymap fall back to DOM input

### Enter

```
Enter[count]
```

Press Enter key. Repeatable: `Enter2` presses Enter twice.

### Backspace

```
Backspace[count]
```

Press Backspace key. Repeatable.

### Tab

```
Tab[@time] [count]
```

Press Tab key. Repeatable. Optional timing override.

### Space

```
Space[count]
```

Press Space key. Repeatable.

### Arrow Keys

```
Up[count]
Down[count]
Left[count]
Right[count]
```

Arrow keys. Repeatable.

### Page Navigation

```
PageUp[count]
PageDown[count]
```

Page Up/Down keys. Repeatable.

### Ctrl

```
Ctrl[+Alt][+Shift]+<char>
```

Control key combinations. Examples:
- `Ctrl+C` — Send Ctrl+C
- `Ctrl+R` — Send Ctrl+R
- `Ctrl+Shift+O` — Send Ctrl+Shift+O
- `Ctrl+Alt+P` — Send Ctrl+Alt+P

Also supports special keys: `Ctrl+Enter`, `Ctrl+Space`, `Ctrl+Left/Right/Up/Down`.

### Alt

```
Alt+<char>
```

Alt modifier + key. Example: `Alt+f`

### Shift

```
Shift+<char>
```

Shift modifier + key.

### Escape

```
Escape
```

Press Escape key.

### Delete / Insert

```
Delete
Insert
```

Special keys.

### Scroll

```
ScrollUp[@time] [count]
ScrollDown[@time] [count]
```

Scroll the terminal viewport. Uses xterm.js `term.scrollLines()`.

### Sleep

```
Sleep <time>
```

Wait (captures frames). Time formats: `0.5`, `2`, `100ms`, `1s`.

### Wait

```
Wait[+Screen|+Line][@timeout] /regex/
```

Wait for regex match on screen or line. Default: `/>$/`, 15s timeout, Line scope.

- `Wait+Screen /pattern/` — Match anywhere on screen
- `Wait+Line /pattern/` — Match on current line (default)
- `Wait@30s /pattern/` — Custom timeout
- `Wait /pattern/` — Default: Line scope, 15s timeout

### Hide / Show

```
Hide
Show
```

`Hide` pauses frame recording (commands still execute). `Show` resumes recording. Useful for setup that shouldn't appear in the output.

### Screenshot

```
Screenshot <path.png>
```

Capture current frame as PNG. The screenshot is composited from text and cursor canvas layers.

### Copy / Paste

```
Copy "text"
Paste
```

`Copy` copies text to system clipboard. `Paste` pastes from system clipboard.

### Env

```
Env <KEY> "value"
```

Set environment variable. Processed before ttyd starts.

### Source

```
Source <path.tape>
```

Include commands from another tape file. No nesting (sourced file cannot itself contain `Source`).

## Comments

```tape
# This is a comment
# Lines starting with # are ignored
```

## String Quoting

Strings can be quoted with single, double, or backtick delimiters:

```tape
Type "double quoted"
Type 'single quoted'
Type `backtick quoted`
```
