# Tape Parsing — Programmatic .tape File Processing

In-depth reference for parsing VHS `.tape` files programmatically using the Go sub-packages.

## Importable Packages

VHS is a `package main` CLI application, but three sub-packages are importable:

```go
import (
    "github.com/charmbracelet/vhs/lexer"
    "github.com/charmbracelet/vhs/parser"
    "github.com/charmbracelet/vhs/token"
)
```

## Lexer

```go
func New(input string) *Lexer
func (l *Lexer) NextToken() token.Token
```

Tokenizes the tape input. Recognizes:
- Keywords: `Set`, `Type`, `Sleep`, `Enter`, `Backspace`, `Ctrl`, `Alt`, `Shift`, `Wait`, `Hide`, `Show`, `Output`, `Require`, `Source`, `Screenshot`, `Copy`, `Paste`, `Env`, etc.
- Literals: strings (single/double/backtick-quoted), numbers, JSON objects, regex patterns
- Operators: `@` (timing), `+` (modifier chaining), `%` (percentage), `=` (assignment)
- Comments: `#` to end of line

## Parser

```go
func New() *Parser
func (p *Parser) Parse(l *lexer.Lexer) ([]Command, error)
```

Produces a `[]Command` list from the lexer output.

### Command Type

```go
type Command struct {
    Type    CommandType  // e.g. token.TYPE, token.ENTER, token.SET
    Options string      // e.g. typing speed "@500ms", setting name "FontSize"
    Args    string      // e.g. the string to type, repeat count, setting value
    Source  string
}
```

### CommandType

CommandType corresponds to `token.TokenType` values:

| Type | Description |
|------|-------------|
| `token.TYPE` | Type characters |
| `token.ENTER` | Press Enter |
| `token.BACKSPACE` | Press Backspace |
| `token.TAB` | Press Tab |
| `token.SPACE` | Press Space |
| `token.CTRL` | Ctrl combination |
| `token.ALT` | Alt combination |
| `token.SHIFT` | Shift combination |
| `token.ESCAPE` | Press Escape |
| `token.DELETE` | Press Delete |
| `token.INSERT` | Press Insert |
| `token.UP/DOWN/LEFT/RIGHT` | Arrow keys |
| `token.PAGEUP/PAGEDOWN` | Page navigation |
| `token.SLEEP` | Wait |
| `token.WAIT` | Wait for pattern |
| `token.HIDE` | Pause recording |
| `token.SHOW` | Resume recording |
| `token.SCREENSHOT` | Capture screenshot |
| `token.SET` | Configure setting |
| `token.OUTPUT` | Set output file |
| `token.REQUIRE` | Check dependency |
| `token.SOURCE` | Include tape file |
| `token.COPY` | Copy to clipboard |
| `token.PASTE` | Paste from clipboard |
| `token.ENV` | Set environment variable |
| `token.SCROLLUP` | Scroll up |
| `token.SCROLLDOWN` | Scroll down |

## Token

```go
type TokenType int
type Token struct {
    Type    TokenType
    Literal string
}
```

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/charmbracelet/vhs/lexer"
    "github.com/charmbracelet/vhs/parser"
    "github.com/charmbracelet/vhs/token"
)

func main() {
    tape := `
Output "demo.gif"
Set Shell "bash"
Type "echo hello"
Enter
Sleep 1
`

    l := lexer.New(tape)
    p := parser.New()
    commands, err := p.Parse(l)
    if err != nil {
        panic(err)
    }

    for _, cmd := range commands {
        switch cmd.Type {
        case token.TYPE:
            fmt.Printf("Type: %q (speed: %s)\n", cmd.Args, cmd.Options)
        case token.ENTER:
            fmt.Printf("Enter (count: %s)\n", cmd.Args)
        case token.SLEEP:
            fmt.Printf("Sleep: %s\n", cmd.Args)
        case token.SET:
            fmt.Printf("Set: %s = %s\n", cmd.Options, cmd.Args)
        case token.OUTPUT:
            fmt.Printf("Output: %s\n", cmd.Args)
        }
    }
}
```

## Converting Tape Commands to go-libghostty Input

The parsed commands can be translated to go-libghostty operations for headless testing:

| VHS Command | go-libghostty Equivalent |
|-------------|------------------------|
| `Type "text"` | `KeyEncoder.Encode()` for each character |
| `Enter` | `KeyEncoder.Encode(KeyEvent{Key: KeyEnter})` |
| `Ctrl+C` | `KeyEncoder.Encode(KeyEvent{Key: KeyC, Mods: ModCtrl})` |
| `Sleep 1` | `time.Sleep(1 * time.Second)` |
| `Set Width/Height` | `term.Resize(cols, rows, cellW, cellH)` |
| `Wait /pattern/` | Poll `Formatter.FormatString()` until regex matches |
| `Screenshot` | `RenderState.Update(term)` + custom renderer |
