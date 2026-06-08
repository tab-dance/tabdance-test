# Recording — Capturing Real Terminal Sessions

In-depth reference for VHS's recording mode: capturing real terminal sessions as `.tape` files.

## vhs record

```sh
vhs record > cassette.tape
```

Starts a PTY session using `creack/pty`, records all keystrokes, and converts them to tape commands.

### How It Works

1. VHS creates a PTY (pseudo-terminal) using `creack/pty`
2. Spawns a shell in the PTY
3. Captures all keystrokes sent to the PTY
4. Converts raw escape sequences to tape DSL commands using the `EscapeSequences` map
5. Writes the tape commands to stdout

### Escape Sequence Mapping

The `inputToTape()` function maps raw escape sequences to tape commands:

| Escape Sequence | Tape Command |
|----------------|-------------|
| `\x1b[A` | `Up` |
| `\x1b[B` | `Down` |
| `\x1b[C` | `Right` |
| `\x1b[D` | `Left` |
| `\x03` | `Ctrl+C` |
| `\x04` | `Ctrl+D` |
| `\x1b` | `Escape` |
| `\x7f` | `Backspace` |
| `\r` | `Enter` |
| `\t` | `Tab` |
| Printable chars | `Type "char"` |

### Usage

```sh
# Record a session
vhs record > demo.tape

# Do things in the terminal...
# Type commands, run programs, etc.

# Exit the shell to stop recording
exit

# The demo.tape file now contains tape commands
cat demo.tape
```

### Output Format

The recorded tape file contains `Type` commands for printable input and key commands for special keys:

```tape
Type "echo hello"
Enter
Type "ls -la"
Enter
Ctrl+C
Up
Enter
```

### Limitations

- No timing information is captured (all `Type` commands use default speed)
- Mouse events are not recorded
- Output/wait conditions are not captured
- The recording only captures input, not terminal output
